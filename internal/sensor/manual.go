package sensor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"

	"github.com/google/uuid"
)

type ManualSensor struct {
	eventBus *bus.EventBus
	cancel   context.CancelFunc
	running  bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	id       uuid.UUID
}

func NewManualSensor(b *bus.EventBus) *ManualSensor {
	return &ManualSensor{
		eventBus: b,
		id:       uuid.Must(uuid.NewV7()),
	}
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

func (m *ManualSensor) run(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("  m - Motion (INFO)")
	fmt.Println("  d - Door (WARNING)")
	fmt.Println("  s - Smoke (CRITICAL)")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Manual sensor stopped.")
			return
		default:
			if !scanner.Scan() {
				return
			}
			text := scanner.Text()
			if len(text) == 0 {
				continue
			}
			ch := text[0]
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
		}
	}
}
