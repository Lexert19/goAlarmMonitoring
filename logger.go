package main

import (
	"context"
	"fmt"
)

type Logger struct {
	bus    *EventBus
	cancel context.CancelFunc
}

func NewLogger(bus *EventBus) *Logger {
	return &Logger{
		bus: bus,
	}
}

func (l *Logger) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	ch := l.bus.Subscribe()
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Logger stopped.")
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				l.Log(event)
			}
		}
	}()
}

func (l *Logger) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *Logger) Log(event Event) {
	fmt.Printf("[%s] %s: %s\n", event.Level, event.Time.Format("15:04:05"), event.Type)
}
