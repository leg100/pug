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
// TODO: there is the potential for a subscriber to become full, i.e. its
// buffered channel is full, in which case the broker will block until the
// channel has free capacity again. This should only happen in extremis, e.g. a
// user has a shit-load of modules/workspaces and invokes a massive number of
// parallel tasks, which in turn publishes a shit-load of events. And if it
// happens, it *should* only happen briefly before the subscriber consumes from
// its channel, freeing up capacity. But if the subscriber does not consume
// because it has blocked on something else indefinitely then the broker will
// block indefinitely.
//
// We need need to know when this happens, via some sort of surfacing of
// metrics that does not get blocked itself...
func (b *Broker[T]) Publish(t resource.EventType, payload T) {
	b.mu.Lock()
	for sub := range b.subs {
		sub <- resource.Event[T]{Type: t, Payload: payload}
	}
	b.mu.Unlock()
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
