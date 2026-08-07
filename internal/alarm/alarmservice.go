package alarm

import (
	"context"
	"fmt"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"
)

type AlarmService struct {
	bus         *bus.EventBus
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup
}

func NewAlarmService(b *bus.EventBus) *AlarmService {
	return &AlarmService{
		bus: b,
	}
}

func (as *AlarmService) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	as.cancel = cancel

	ch, unsub := as.bus.Subscribe()
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
				if event.Type == types.Smoke {
					fmt.Printf("Smoke detected at %s (DeviceID: %s)\n", event.Time.Format("15:04:05"), event.DeviceID)
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
