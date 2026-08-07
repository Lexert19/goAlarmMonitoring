package types

import (
	"time"

	"github.com/google/uuid"
)

type EventType int

const (
	Motion EventType = iota
	Door
	Smoke
)

type Event interface {
	GetDeviceID() uuid.UUID
	GetTime() time.Time
}

func (e EventType) String() string {
	switch e {
	case Motion:
		return "Motion"
	case Door:
		return "Door"
	case Smoke:
		return "Smoke"
	default:
		return "Unknown"
	}
}

type Level string

const (
	INFO     Level = "INFO"
	WARNING  Level = "WARNING"
	CRITICAL Level = "CRITICAL"
)

type SensorEvent struct {
	DeviceID uuid.UUID
	Type     EventType
	Time     time.Time
	Level    Level
}

func NewSensorEvent(typ EventType, level Level, deviceID uuid.UUID) SensorEvent {
	return SensorEvent{
		DeviceID: deviceID,
		Type:     typ,
		Time:     time.Now(),
		Level:    level,
	}
}

func (e SensorEvent) GetDeviceID() uuid.UUID { return e.DeviceID }
func (e SensorEvent) GetTime() time.Time     { return e.Time }

type AlarmCreatedEvent struct {
	AlarmID   uuid.UUID
	DeviceID  uuid.UUID
	AlarmType EventType
	CreatedAt time.Time
}

func (e AlarmCreatedEvent) GetDeviceID() uuid.UUID { return e.DeviceID }
func (e AlarmCreatedEvent) GetTime() time.Time     { return e.CreatedAt }
