package bus

import (
	"sync"

	"goAlarmMonitoring/pkg/types"
)

type EventBus struct {
	subscribers []chan types.Event
	mu          sync.Mutex
	bufferSize  int
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		subscribers: make([]chan types.Event, 0),
		bufferSize:  bufferSize,
	}
}

func (eb *EventBus) Subscribe() <-chan types.Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan types.Event, eb.bufferSize)
	eb.subscribers = append(eb.subscribers, ch)
	return ch
}

func (eb *EventBus) Publish(event types.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, ch := range eb.subscribers {
		ch <- event
	}
}

func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, ch := range eb.subscribers {
		close(ch)
	}
}
