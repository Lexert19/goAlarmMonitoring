package sensor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"

	"github.com/eapache/queue"
	"github.com/google/uuid"
)

type RegistryCommander interface {
	AddSimulatedSensor(ctx context.Context) (Sensor, error)
	Reconfigure(id uuid.UUID, key string, value interface{}) error
	ListAll() []Sensor
}

type ManualSensor struct {
	publisher bus.EventPublisher
	registry  RegistryCommander
	cancel    context.CancelFunc
	running   bool
	mu        sync.Mutex
	wg        sync.WaitGroup
	id        uuid.UUID
	cmdQueue  *queue.Queue
	notifyCh  chan struct{}
}

func NewManualSensor(pub bus.EventPublisher, reg RegistryCommander) *ManualSensor {
	return &ManualSensor{
		publisher: pub,
		registry:  reg,
		id:        uuid.Must(uuid.NewV7()),
		cmdQueue:  queue.New(),
		notifyCh:  make(chan struct{}, 1),
	}
}

func (m *ManualSensor) ID() uuid.UUID {
	return m.id
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
	fmt.Println("  add - Add new simulated sensor")
	fmt.Println("  list - List all active sensors")
	fmt.Println("  reconfig <uuid> <key> <value> - Reconfigure sensor by exact UUID")

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

				m.processCommand(ctx, line)
			}
		}
	}
}

func (m *ManualSensor) processCommand(ctx context.Context, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	parts := strings.Fields(line)
	if len(parts) == 1 {
		switch parts[0] {
		case "m":
			m.publisher.Publish(types.NewSensorEvent(types.Motion, types.INFO, m.id))
			return
		case "d":
			m.publisher.Publish(types.NewSensorEvent(types.Door, types.WARNING, m.id))
			return
		case "s":
			m.publisher.Publish(types.NewSensorEvent(types.Smoke, types.CRITICAL, m.id))
			return
		case "add":
			s, err := m.registry.AddSimulatedSensor(ctx)
			if err != nil {
				fmt.Printf("Failed to add sensor: %v\n", err)
			} else {
				fmt.Printf("Added sensor ID: %s\n", s.ID())
			}
			return
		case "list", "ls":
			sensors := m.registry.ListAll()
			fmt.Println("\n--- Active Sensors ---")
			for _, s := range sensors {
				fmt.Printf("ID: %s | Running: %v\n", s.ID(), s.IsRunning())
			}
			fmt.Println("----------------------")
			return
		}
	}

	if parts[0] == "reconfig" {
		if len(parts) != 4 {
			fmt.Println("Usage: reconfig <uuid> <key> <value>")
			return
		}

		parsedID, err := uuid.Parse(parts[1])
		if err != nil {
			fmt.Printf("Invalid UUID format: %v\n", err)
			return
		}

		key := parts[2]
		rawVal := parts[3]
		var val interface{}

		switch key {
		case "motion", "door", "smoke":
			f, err := strconv.ParseFloat(rawVal, 64)
			if err != nil || f < 0 || f > 1 {
				fmt.Println("Value must be float between 0.0 and 1.0")
				return
			}
			val = f
		case "ticker":
			i, err := strconv.Atoi(rawVal)
			if err != nil || i <= 0 {
				fmt.Println("Value must be positive integer")
				return
			}
			val = i
		default:
			fmt.Println("Unknown key. Available: motion, door, smoke, ticker")
			return
		}

		if err := m.registry.Reconfigure(parsedID, key, val); err != nil {
			fmt.Printf("Reconfigure error: %v\n", err)
		} else {
			fmt.Printf("Sensor %s reconfigured successfully\n", parsedID)
		}
		return
	}
}
