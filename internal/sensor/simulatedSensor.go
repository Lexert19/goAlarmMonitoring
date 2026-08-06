package sensor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/pkg/types"

	"github.com/google/uuid"
)

type SimulatedSensor struct {
	bus     *bus.EventBus
	config  *config.Config
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
	id      uuid.UUID
}

func NewSimulatedSensor(b *bus.EventBus, cfg *config.Config) *SimulatedSensor {
	return &SimulatedSensor{
		bus:    b,
		config: cfg,
		id:     uuid.Must(uuid.NewV7()),
	}
}

func (s *SimulatedSensor) ID() uuid.UUID {
	return s.id
}

func (s *SimulatedSensor) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return fmt.Errorf("sensor already running")
	}
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true
	go s.run(ctx)
	return nil
}

func (s *SimulatedSensor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.running = false
}

func (s *SimulatedSensor) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *SimulatedSensor) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.TickerSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Sensor stopped.")
			return
		case <-ticker.C:
			if rand.Float64() < s.config.MotionProb {
				s.publish(types.Motion, types.INFO)
			}
			if rand.Float64() < s.config.DoorProb {
				s.publish(types.Door, types.WARNING)
			}
			if rand.Float64() < s.config.SmokeProb {
				s.publish(types.Smoke, types.CRITICAL)
			}
		}
	}
}

func (s *SimulatedSensor) publish(typ types.EventType, level types.Level) {
	go s.bus.Publish(types.NewEvent(typ, level, s.id))
}
