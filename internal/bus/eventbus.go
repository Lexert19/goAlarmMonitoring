package bus

import (
	"sync"

	"goAlarmMonitoring/pkg/types"
)

type EventBus struct {
	subscribers []chan types.Event
	mu          sync.Mutex
	queue       []types.Event
	queueMu     sync.Mutex
	cond        *sync.Cond
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewEventBus() *EventBus {
	eb := &EventBus{
		subscribers: make([]chan types.Event, 0),
		queue:       make([]types.Event, 0),
		stop:        make(chan struct{}),
	}
	eb.cond = sync.NewCond(&eb.queueMu)
	eb.wg.Add(1)
	go eb.dispatchLoop()
	return eb
}

func (eb *EventBus) Subscribe() <-chan types.Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan types.Event, 1000)
	eb.subscribers = append(eb.subscribers, ch)
	return ch
}

func (eb *EventBus) Publish(event types.Event) {
	eb.queueMu.Lock()
	eb.queue = append(eb.queue, event)
	eb.queueMu.Unlock()
	eb.cond.Signal()
}

func (eb *EventBus) dispatchLoop() {
	defer eb.wg.Done()
	for {
		eb.queueMu.Lock()
		for len(eb.queue) == 0 {
			eb.cond.Wait()
			select {
			case <-eb.stop:
				eb.queueMu.Unlock()
				return
			default:
			}
		}
		event := eb.queue[0]
		eb.queue = eb.queue[1:]
		eb.queueMu.Unlock()

		eb.mu.Lock()
		for _, ch := range eb.subscribers {
			go func(c chan types.Event, e types.Event) {
				c <- e
			}(ch, event)
		}
		eb.mu.Unlock()
	}
}

func (eb *EventBus) Close() {
	close(eb.stop)
	eb.cond.Signal()
	eb.wg.Wait()
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, ch := range eb.subscribers {
		close(ch)
	}
}
