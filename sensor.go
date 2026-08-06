package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Sensor struct {
	bus     *EventBus
	config  *Config
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
}

func NewSensor(bus *EventBus, config *Config) *Sensor {
	return &Sensor{
		bus:    bus,
		config: config,
	}
}

func (s *Sensor) Start(ctx context.Context) error {
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

func (s *Sensor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.running = false
}

func (s *Sensor) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Sensor) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.TickerSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Sensor stopped.")
			return
		case <-ticker.C:
			if rand.Float64() < s.config.MotionProb {
				s.publishEvent(Motion, INFO)
			}
			if rand.Float64() < s.config.DoorProb {
				s.publishEvent(Door, WARNING)
			}
			if rand.Float64() < s.config.SmokeProb {
				s.publishEvent(Smoke, CRITICAL)
			}
		}
	}
}

func (s *Sensor) publishEvent(typ EventType, level Level) {
	event := Event{
		Type:  typ,
		Time:  time.Now(),
		Level: level,
	}
	s.bus.Publish(event)
}

func (s *Sensor) ManualEvent(typ EventType) {
	level := INFO
	switch typ {
	case Door:
		level = WARNING
	case Smoke:
		level = CRITICAL
	}
	s.publishEvent(typ, level)
}
