package pubsub

import (
	"context"
	"sync"

	"github.com/leg100/pug/internal/resource"
)

const bufferSize = 1024

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
	done   chan struct{}                       // close when broker is shutting down
	logger Logger
}

// NewBroker constructs a pub/sub broker.
func NewBroker[T any](logger Logger) *Broker[T] {
	b := &Broker[T]{
		subs:   make(map[chan resource.Event[T]]struct{}),
		done:   make(chan struct{}),
		logger: logger,
	}
	return b
}

// Shutdown the broker, terminating any subscriptions.
func (b *Broker[T]) Shutdown() {
	close(b.done)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Remove each subscriber entry, so Publish() cannot send any further
	// messages, and close each subscriber's channel, so the subscriber cannot
	// consume any more messages.
	for ch := range b.subs {
		delete(b.subs, ch)
		close(ch)
	}
}

// Subscribe subscribes the caller to a stream of events. The returned channel
// is closed when the broker is shutdown.
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan resource.Event[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if broker has shutdown and if so return closed channel
	select {
	case <-b.done:
		ch := make(chan resource.Event[T])
		close(ch)
		return ch
	default:
	}

	// Subscribe
	sub := make(chan resource.Event[T], bufferSize)
	b.subs[sub] = struct{}{}

	// Unsubscribe when context is done.
	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		// Check if broker has shutdown and if so do nothing
		select {
		case <-b.done:
			return
		default:
		}

		delete(b.subs, sub)
		close(sub)
	}()

	return sub
}

// Publish an event to subscribers.
func (b *Broker[T]) Publish(t resource.EventType, payload T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for sub := range b.subs {
		select {
		case sub <- resource.Event[T]{Type: t, Payload: payload}:
		case <-b.done:
			return
		}
	}
}
