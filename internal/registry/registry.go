package registry

import (
	"context"
	"fmt"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/sensor"

	"github.com/google/uuid"
)

type DeviceRegistry interface {
	AddSimulatedSensor(ctx context.Context) (sensor.Sensor, error)
	Register(s sensor.Sensor) error
	Get(id uuid.UUID) (sensor.Sensor, bool)
	Reconfigure(id uuid.UUID, key string, value interface{}) error
	ListAll() []sensor.Sensor
	StopAll()
}

type MemoryDeviceRegistry struct {
	mu      sync.RWMutex
	sensors map[uuid.UUID]sensor.Sensor
	pub     bus.EventPublisher
	cfg     *config.Config
}

func NewMemoryDeviceRegistry(pub bus.EventPublisher, cfg *config.Config) *MemoryDeviceRegistry {
	return &MemoryDeviceRegistry{
		sensors: make(map[uuid.UUID]sensor.Sensor),
		pub:     pub,
		cfg:     cfg,
	}
}

func (r *MemoryDeviceRegistry) AddSimulatedSensor(ctx context.Context) (sensor.Sensor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := sensor.NewSimulatedSensor(r.pub, r.cfg)
	if err := s.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start sensor: %w", err)
	}

	r.sensors[s.ID()] = s
	return s, nil
}

func (r *MemoryDeviceRegistry) Register(s sensor.Sensor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sensors[s.ID()]; exists {
		return fmt.Errorf("sensor with ID %s already registered", s.ID())
	}

	r.sensors[s.ID()] = s
	return nil
}

func (r *MemoryDeviceRegistry) Get(id uuid.UUID) (sensor.Sensor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.sensors[id]
	return s, exists
}

func (r *MemoryDeviceRegistry) Reconfigure(id uuid.UUID, key string, value interface{}) error {
	r.mu.RLock()
	s, exists := r.sensors[id]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("sensor with ID %s not found", id)
	}

	return s.ReconfigureOne(key, value)
}

func (r *MemoryDeviceRegistry) ListAll() []sensor.Sensor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]sensor.Sensor, 0, len(r.sensors))
	for _, s := range r.sensors {
		list = append(list, s)
	}
	return list
}

func (r *MemoryDeviceRegistry) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.sensors {
		s.Stop()
	}
}
