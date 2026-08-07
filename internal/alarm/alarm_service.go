package alarm

import (
	"time"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"

	"github.com/google/uuid"
)

type Alarm struct {
	ID        uuid.UUID
	DeviceID  uuid.UUID
	Type      types.EventType
	CreatedAt time.Time
}

type AlarmService struct {
	publisher bus.EventPublisher
	repo      AlarmRepository
}

func NewAlarmService(pub bus.EventPublisher, repo AlarmRepository) *AlarmService {
	return &AlarmService{
		publisher: pub,
		repo:      repo,
	}
}

func (as *AlarmService) TriggerAlarm(deviceID uuid.UUID, alarmType types.EventType, createdAt time.Time) error {
	a := Alarm{
		ID:        uuid.Must(uuid.NewV7()),
		DeviceID:  deviceID,
		Type:      alarmType,
		CreatedAt: createdAt,
	}
	if err := as.repo.Save(a); err != nil {
		return err
	}

	as.publisher.Publish(types.AlarmCreatedEvent{
		AlarmID:   a.ID,
		DeviceID:  a.DeviceID,
		AlarmType: a.Type,
		CreatedAt: a.CreatedAt,
	})
	return nil
}
