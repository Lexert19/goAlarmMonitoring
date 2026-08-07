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
	cmdQueue       *queue.Queue
	notifyCh       chan struct{}
}

func NewManualSensor(b *bus.EventBus) *ManualSensor {
	return &ManualSensor{
		eventBus: b,
		id:       uuid.Must(uuid.NewV7()),
		cmdQueue: queue.New(),
		notifyCh: make(chan struct{}, 1),
	}
}

func (m *ManualSensor) ID() uuid.UUID {
	return m.id
}

func (m *ManualSensor) SetReconfigurator(r CommandProcessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reconfigurator = r
}

func (m *ManualSensor) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *ManualSensor) ReconfigureOne(key string, value interface{}) error {
	return fmt.Errorf("manual sensor does not support reconfiguration")
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

	m.wg.Add(2)
	go m.readStdin(ctx)
	go m.run(ctx)
	return nil
}

func (m *ManualSensor) Stop() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.running = false
	m.mu.Unlock()

	os.Stdin.Close()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		fmt.Println("Manual sensor stop timed out")
	}
}

func (m *ManualSensor) readStdin(ctx context.Context) {
	defer m.wg.Done()

	lines := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			m.mu.Lock()
			m.cmdQueue.Add(line)
			m.mu.Unlock()

			select {
			case m.notifyCh <- struct{}{}:
			default:
			}
		}
	}
}

func (m *ManualSensor) run(ctx context.Context) {
	defer m.wg.Done()

	fmt.Println("Commands:")
	fmt.Println("  m - Motion (INFO)")
	fmt.Println("  d - Door (WARNING)")
	fmt.Println("  s - Smoke (CRITICAL)")
	fmt.Println("  reconfig <sensor_id> <key> <value> - reconfigure sensor")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Manual sensor stopped.")
			return

		case <-m.notifyCh:
			for {
				m.mu.Lock()
				if m.cmdQueue.Length() == 0 {
					m.mu.Unlock()
					break
				}
				line := m.cmdQueue.Remove().(string)
				m.mu.Unlock()

				m.processCommand(line)
			}
		}
	}
}

func (m *ManualSensor) processCommand(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
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
			return
		}
		event := types.NewEvent(typ, level, m.id)
		m.eventBus.Publish(event)
		fmt.Printf("Manual event published: %s %s\n", typ, level)
	} else {
		m.mu.Lock()
		reconfig := m.reconfigurator
		m.mu.Unlock()

		if reconfig != nil {
			reconfig.ProcessCommand(line)
		} else {
			fmt.Println("Reconfigurator not set")
		}
	}
}
