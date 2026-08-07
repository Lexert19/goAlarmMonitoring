package logger

import (
	"context"
	"fmt"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"
)

type Logger struct {
	bus         *bus.EventBus
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup
}

func NewLogger(b *bus.EventBus) *Logger {
	return &Logger{
		bus: b,
	}
}
func (l *Logger) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	ch, unsub := l.bus.Subscribe()
	l.unsubscribe = unsub

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
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
	l.wg.Wait()
	if l.unsubscribe != nil {
		l.unsubscribe()
	}
}

func (l *Logger) Log(event types.Event) {
	fmt.Printf("[%s] %s [DeviceID:%s] %s\n", event.Level, event.Time.Format("15:04:05"), event.DeviceID, event.Type)
}
