package alarm

import (
	"context"
	"fmt"
	"sync"
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
	subscriber  bus.EventSubscriber
	publisher   bus.EventPublisher
	repo        AlarmRepository
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup
}

func NewAlarmService(sub bus.EventSubscriber, pub bus.EventPublisher, repo AlarmRepository) *AlarmService {
	return &AlarmService{
		subscriber: sub,
		publisher:  pub,
		repo:       repo,
	}
}

func (as *AlarmService) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	as.cancel = cancel

	ch, unsub := as.subscriber.Subscribe()
	as.unsubscribe = unsub

	as.wg.Add(1)
	go func() {
		defer as.wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("AlarmService stopped.")
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				if sensorEvent, ok := event.(types.SensorEvent); ok {
					if sensorEvent.Type == types.Smoke {
						a := Alarm{
							ID:        uuid.Must(uuid.NewV7()),
							DeviceID:  sensorEvent.DeviceID,
							Type:      sensorEvent.Type,
							CreatedAt: sensorEvent.Time,
						}
						if err := as.repo.Save(a); err == nil {
							as.publisher.Publish(types.AlarmCreatedEvent{
								AlarmID:   a.ID,
								DeviceID:  a.DeviceID,
								AlarmType: a.Type,
								CreatedAt: a.CreatedAt,
							})
						}
					}
				}
			}
		}
	}()
}

func (as *AlarmService) Stop() {
	if as.cancel != nil {
		as.cancel()
	}
	as.wg.Wait()
	if as.unsubscribe != nil {
		as.unsubscribe()
	}
}
