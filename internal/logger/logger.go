package logger

import (
	"context"
	"fmt"
	"sync"

	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/pkg/types"
)

type Logger struct {
	subscriber  bus.EventSubscriber
	cancel      context.CancelFunc
	unsubscribe func()
	wg          sync.WaitGroup
}

func NewLogger(sub bus.EventSubscriber) *Logger {
	return &Logger{
		subscriber: sub,
	}
}

func (l *Logger) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	ch, unsub := l.subscriber.Subscribe()
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

func (l *Logger) Log(event any) {
	switch e := event.(type) {
	case types.SensorEvent:
		fmt.Printf("[%s] %s [DeviceID:%s] %s\n", e.Level, e.Time.Format("15:04:05"), e.DeviceID, e.Type)
	case types.AlarmCreatedEvent:
		fmt.Printf("[ALARM EVENT] %s [AlarmID:%s] [DeviceID:%s] AlarmType:%s\n",
			e.CreatedAt.Format("15:04:05"), e.AlarmID, e.DeviceID, e.AlarmType)
	}
}
