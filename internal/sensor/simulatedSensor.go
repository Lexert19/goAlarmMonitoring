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
	bus        *bus.EventBus
	config     *config.Config
	cancel     context.CancelFunc
	running    bool
	mu         sync.Mutex
	id         uuid.UUID
	motionProb float64
	doorProb   float64
	smokeProb  float64
	tickerSec  int
	restartCh  chan struct{}
	wg         sync.WaitGroup
}

func NewSimulatedSensor(b *bus.EventBus, cfg *config.Config) *SimulatedSensor {
	return &SimulatedSensor{
		bus:        b,
		config:     cfg,
		id:         uuid.Must(uuid.NewV7()),
		motionProb: cfg.MotionProb,
		doorProb:   cfg.DoorProb,
		smokeProb:  cfg.SmokeProb,
		tickerSec:  cfg.TickerSec,
		restartCh:  make(chan struct{}, 1),
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

	s.wg.Add(1)
	go s.run(ctx)
	return nil
}

func (s *SimulatedSensor) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.running = false
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *SimulatedSensor) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *SimulatedSensor) ReconfigureOne(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch key {
	case "motion":
		v, ok := value.(float64)
		if !ok || v < 0 || v > 1 {
			return fmt.Errorf("invalid motion prob")
		}
		s.motionProb = v
	case "door":
		v, ok := value.(float64)
		if !ok || v < 0 || v > 1 {
			return fmt.Errorf("invalid door prob")
		}
		s.doorProb = v
	case "smoke":
		v, ok := value.(float64)
		if !ok || v < 0 || v > 1 {
			return fmt.Errorf("invalid smoke prob")
		}
		s.smokeProb = v
	case "ticker":
		v, ok := value.(int)
		if !ok || v <= 0 {
			return fmt.Errorf("invalid ticker sec")
		}
		if s.tickerSec != v {
			s.tickerSec = v
			select {
			case s.restartCh <- struct{}{}:
			default:
			}
		}
	default:
		return fmt.Errorf("unknown key")
	}
	return nil
}

func (s *SimulatedSensor) run(ctx context.Context) {
	defer s.wg.Done()

	makeTicker := func() *time.Ticker {
		s.mu.Lock()
		defer s.mu.Unlock()
		return time.NewTicker(time.Duration(s.tickerSec) * time.Second)
	}
	ticker := makeTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Sensor stopped.")
			return
		case <-ticker.C:
			s.mu.Lock()
			motionProb := s.motionProb
			doorProb := s.doorProb
			smokeProb := s.smokeProb
			s.mu.Unlock()

			if rand.Float64() < motionProb {
				s.publish(types.Motion, types.INFO)
			}
			if rand.Float64() < doorProb {
				s.publish(types.Door, types.WARNING)
			}
			if rand.Float64() < smokeProb {
				s.publish(types.Smoke, types.CRITICAL)
			}
		case <-s.restartCh:
			ticker.Stop()
			ticker = makeTicker()
		}
	}
}

func (s *SimulatedSensor) publish(typ types.EventType, level types.Level) {
	s.bus.Publish(types.NewEvent(typ, level, s.id))
}
