package main

import (
	"sync"
)

type EventBus struct {
	subscribers []chan Event
	mu          sync.Mutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make([]chan Event, 0),
	}
}

func (eb *EventBus) Subscribe() <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan Event, 20)
	eb.subscribers = append(eb.subscribers, ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, ch := range eb.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, ch := range eb.subscribers {
		close(ch)
	}
}
