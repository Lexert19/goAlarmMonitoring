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

type Event struct {
	DeviceID uuid.UUID
	Type     EventType
	Time     time.Time
	Level    Level
}

func NewEvent(typ EventType, level Level, deviceID uuid.UUID) Event {
	return Event{
		DeviceID: deviceID,
		Type:     typ,
		Time:     time.Now(),
		Level:    level,
	}
}
