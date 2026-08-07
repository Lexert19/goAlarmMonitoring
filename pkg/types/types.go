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

type AlarmCreatedEvent struct {
	AlarmID   uuid.UUID
	DeviceID  uuid.UUID
	AlarmType EventType
	CreatedAt time.Time
}
