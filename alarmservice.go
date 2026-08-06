package main

import (
	"context"
	"fmt"
	"sync"
)

type AlarmService struct {
	bus    *EventBus
	logger *Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewAlarmService(bus *EventBus, logger *Logger) *AlarmService {
	return &AlarmService{
		bus:    bus,
		logger: logger,
	}
}

func (as *AlarmService) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	as.cancel = cancel

	ch := as.bus.Subscribe()
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
				as.logger.Log(event)

				if event.Type == Smoke {
					fmt.Printf("Smoke detected at %s!\n", event.Time.Format("15:04:05"))
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
}
