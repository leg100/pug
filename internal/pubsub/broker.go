package pubsub

import (
	"context"
	"errors"
	"sync"

	"github.com/leg100/pug/internal/resource"
)

const (
	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 1024
)

// ErrSubscriptionTerminated is for use by subscribers to indicate that their
// subscription has been terminated by the broker.
var ErrSubscriptionTerminated = errors.New("broker terminated the subscription")

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Broker allows clients to publish events and subscribe to events
type Broker[T any] struct {
	subs map[chan resource.Event[T]]struct{} // subscriptions
	mu   sync.Mutex                          // sync access to map

	logger Logger
}

func NewBroker[T any](logger Logger) *Broker[T] {
	b := &Broker[T]{
		subs:   make(map[chan resource.Event[T]]struct{}),
		logger: logger,
	}
	return b
}

// Subscribe subscribes the caller to a stream of events. The caller can close
// the subscription by either canceling the context or calling the returned
// unsubscribe function.
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan resource.Event[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan resource.Event[T], subBufferSize)
	b.subs[sub] = struct{}{}

	// when the context is canceled remove the subscriber
	go func() {
		<-ctx.Done()
		b.unsubscribe(sub)
	}()

	return sub
}

// Publish an event to subscribers.
//
// TODO: don't forceably unsubscribe full subscribers: subscribers in pug
// typically aren't setup to re-subscribe
func (b *Broker[T]) Publish(t resource.EventType, payload T) {
	var fullSubscribers []chan resource.Event[T]

	b.mu.Lock()
	for sub := range b.subs {
		select {
		case sub <- resource.Event[T]{Type: t, Payload: payload}:
			continue
		default:
			// could not publish event to subscriber because their buffer is
			// full, so add them to a list for action below
			fullSubscribers = append(fullSubscribers, sub)
		}
	}
	b.mu.Unlock()

	// forceably unsubscribe full subscribers and leave it to them to
	// re-subscribe
	for _, name := range fullSubscribers {
		b.logger.Error("unsubscribing full subscriber", "sub", name, "queue_length", subBufferSize)
		b.unsubscribe(name)
	}
}

func (b *Broker[T]) unsubscribe(sub chan resource.Event[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subs[sub]; !ok {
		// already unsubscribed
		return
	}
	close(sub)
	delete(b.subs, sub)
}
