package bus

import (
	"sync"

	"goAlarmMonitoring/pkg/types"

	"github.com/eapache/queue"
)

type subscriber struct {
	queue  *queue.Queue
	done   chan struct{}
	notify chan struct{}
	mu     sync.Mutex
}

type EventBus struct {
	subscribers map[<-chan types.Event]*subscriber
	mu          sync.RWMutex
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewEventBus() *EventBus {
	eb := &EventBus{
		subscribers: make(map[<-chan types.Event]*subscriber),
		stop:        make(chan struct{}),
	}
	return eb
}

func (eb *EventBus) Subscribe() (<-chan types.Event, func()) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	outCh := make(chan types.Event, 100)
	sub := &subscriber{
		queue:  queue.New(),
		done:   make(chan struct{}),
		notify: make(chan struct{}, 1),
	}

	eb.subscribers[outCh] = sub

	eb.wg.Add(1)
	go eb.bufferLoop(sub, outCh)

	unsubscribe := func() {
		eb.mu.Lock()
		delete(eb.subscribers, outCh)
		eb.mu.Unlock()
		close(sub.done)
	}

	return outCh, unsubscribe
}

func (eb *EventBus) Publish(event types.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, sub := range eb.subscribers {
		sub.mu.Lock()
		wasEmpty := sub.queue.Length() == 0
		sub.queue.Add(event)
		sub.mu.Unlock()

		if wasEmpty {
			select {
			case sub.notify <- struct{}{}:
			default:
			}
		}
	}
}

func (eb *EventBus) bufferLoop(sub *subscriber, outCh chan<- types.Event) {
	defer eb.wg.Done()
	defer close(outCh)

	for {
		sub.mu.Lock()
		if sub.queue.Length() == 0 {
			sub.mu.Unlock()
			select {
			case <-sub.done:
				return
			case <-eb.stop:
				return
			case <-sub.notify:
				continue
			}
		}

		event := sub.queue.Remove().(types.Event)
		sub.mu.Unlock()

		select {
		case outCh <- event:
		case <-sub.done:
			return
		case <-eb.stop:
			return
		}
	}
}

func (eb *EventBus) Close() {
	close(eb.stop)
	eb.wg.Wait()
	eb.mu.Lock()
	eb.subscribers = nil
	eb.mu.Unlock()
}
