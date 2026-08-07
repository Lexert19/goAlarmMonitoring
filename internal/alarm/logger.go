package alarm

import (
	"context"
	"fmt"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"
)

type AlarmLogger struct {
	subscriber  bus.EventSubscriber
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup
}

func NewAlarmLogger(sub bus.EventSubscriber) *AlarmLogger {
	return &AlarmLogger{
		subscriber: sub,
	}
}

func (al *AlarmLogger) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	al.cancel = cancel

	ch, unsub := al.subscriber.Subscribe()
	al.unsubscribe = unsub

	al.wg.Add(1)
	go func() {
		defer al.wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("AlarmLogger stopped.")
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				if alarmEvent, ok := event.(types.AlarmCreatedEvent); ok {
					al.Log(alarmEvent)
				}
			}
		}
	}()
}

func (al *AlarmLogger) Stop() {
	if al.cancel != nil {
		al.cancel()
	}
	al.wg.Wait()
	if al.unsubscribe != nil {
		al.unsubscribe()
	}
}

func (al *AlarmLogger) Log(e types.AlarmCreatedEvent) {
	fmt.Printf("[ALARM LOGGER] New Alarm Created! AlarmID: %s | DeviceID: %s | Type: %s | Time: %s\n",
		e.AlarmID, e.DeviceID, e.AlarmType, e.CreatedAt.Format("15:04:05"))
}
