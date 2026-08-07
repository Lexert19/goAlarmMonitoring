package bus

import (
	"goAlarmMonitoring/pkg/types"
	"sync"

	"github.com/eapache/queue"
)

type Event = types.Event

type EventPublisher interface {
	Publish(event Event)
}

type EventSubscriber interface {
	Subscribe() (<-chan Event, func())
}

type subscriber struct {
	queue  *queue.Queue
	done   chan struct{}
	notify chan struct{}
	mu     sync.Mutex
}

type EventBus struct {
	subscribers map[<-chan Event]*subscriber
	mu          sync.RWMutex
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewEventBus() *EventBus {
	eb := &EventBus{
		subscribers: make(map[<-chan Event]*subscriber),
		stop:        make(chan struct{}),
	}
	return eb
}

func (eb *EventBus) Subscribe() (<-chan Event, func()) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	outCh := make(chan Event, 100)
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

func (eb *EventBus) Publish(event Event) {
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

func (eb *EventBus) bufferLoop(sub *subscriber, outCh chan<- Event) {
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

		event := sub.queue.Remove().(Event)
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
