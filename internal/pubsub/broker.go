package pubsub

import (
	"sync"

	"github.com/leg100/pug/internal/resource"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Broker allows clients to publish events and subscribe to events
type Broker[T any] struct {
	subs   map[chan resource.Event[T]]struct{} // subscriptions
	mu     sync.Mutex                          // sync access to map
	wg     sync.WaitGroup                      // sync closure of subscriptions
	logger Logger
}

// NewBroker constructs a pub/sub broker.
func NewBroker[T any](logger Logger) *Broker[T] {
	b := &Broker[T]{
		subs:   make(map[chan resource.Event[T]]struct{}),
		logger: logger,
	}
	return b
}

// Shutdown the broker, terminating any subscriptions.
func (b *Broker[T]) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Wait for any in-flight go-routines in Publish() to finish sending to
	// subscriber channels.
	b.wg.Wait()

	// Remove each subscriber entry, so Publish() cannot send any further
	// messages, and close each subscriber's channel, so the subscriber knows to
	// stop consuming messages.
	for ch := range b.subs {
		delete(b.subs, ch)
		close(ch)
	}
}

// Subscribe subscribes the caller to a stream of events. The returned channel
// is closed when the broker is shutdown.
func (b *Broker[T]) Subscribe() <-chan resource.Event[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan resource.Event[T])
	b.subs[sub] = struct{}{}
	return sub
}

// Publish an event to subscribers.
func (b *Broker[T]) Publish(t resource.EventType, payload T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for sub := range b.subs {
		b.wg.Add(1)
		go func() {
			sub <- resource.Event[T]{Type: t, Payload: payload}
			b.wg.Done()
		}()
	}
}
