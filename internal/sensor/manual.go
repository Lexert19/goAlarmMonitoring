package sensor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"

	"github.com/eapache/queue"
	"github.com/google/uuid"
)

type ManualSensor struct {
	eventBus       *bus.EventBus
	cancel         context.CancelFunc
	running        bool
	mu             sync.Mutex
	wg             sync.WaitGroup
	id             uuid.UUID
	reconfigurator CommandProcessor
	queue          *queue.Queue
}

func NewManualSensor(b *bus.EventBus) *ManualSensor {
	return &ManualSensor{
		eventBus: b,
		id:       uuid.Must(uuid.NewV7()),
		queue:    queue.New(),
	}
}

func (m *ManualSensor) SetReconfigurator(r CommandProcessor) {
	m.reconfigurator = r
}

func (m *ManualSensor) ID() uuid.UUID {
	return m.id
}

func (m *ManualSensor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return fmt.Errorf("manual sensor already running")
	}
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.running = true
	go m.run(ctx)
	return nil
}

func (m *ManualSensor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.running = false
}

func (m *ManualSensor) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *ManualSensor) ReconfigureOne(key string, value interface{}) error {
	return fmt.Errorf("manual sensor does not support reconfiguration")
}

func (m *ManualSensor) run(ctx context.Context) {
	fmt.Println("Commands:")
	fmt.Println("  m - Motion (INFO)")
	fmt.Println("  d - Door (WARNING)")
	fmt.Println("  s - Smoke (CRITICAL)")
	fmt.Println("  reconfig <sensor_id> <key> <value> - reconfigure sensor")

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			m.queue.Add(line)
		}
	}()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Manual sensor stopped.")
			os.Stdin.Close()
			return
		case <-ticker.C:
			for m.queue.Length() > 0 {
				line := m.queue.Remove().(string)
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) == 1 && len(parts[0]) == 1 {
					ch := parts[0][0]
					var typ types.EventType
					var level types.Level
					switch ch {
					case 'm':
						typ, level = types.Motion, types.INFO
					case 'd':
						typ, level = types.Door, types.WARNING
					case 's':
						typ, level = types.Smoke, types.CRITICAL
					default:
						fmt.Println("Unknown command. Available: m, d, s")
						continue
					}
					event := types.NewEvent(typ, level, m.id)
					go m.eventBus.Publish(event)
					fmt.Printf("Manual event published: %s %s\n", typ, level)
				} else {
					if m.reconfigurator != nil {
						m.reconfigurator.ProcessCommand(line)
					} else {
						fmt.Println("Reconfigurator not set")
					}
				}
			}
		}
	}
}
