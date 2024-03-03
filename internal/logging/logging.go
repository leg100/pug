package logging

import (
	"context"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/slog"
)

// levels is the slog mapping of levels to integers:
// https://pkg.go.dev/log/slog#Level
var levels = map[string]int{
	"debug": -4,
	"info":  0,
	"warn":  4,
	"error": 8,
}

// Message is the event payload type
type Message []byte

type logger struct {
	Logger *slog.Logger
	broker *pubsub.Broker[Message]
}

// NewLogger constructs a slog logger with the appropriate log level. The logger
// writes its log messages as pug events.
func NewLogger(level string) *logger {
	broker := pubsub.NewBroker[Message]()

	// Configure logging
	handler := slog.NewTextHandler(
		&writer{broker},
		&slog.HandlerOptions{
			Level: slog.Level(levels[level]),
		},
	)
	return &logger{
		Logger: slog.New(handler),
		broker: broker,
	}
}

// Subscribe to log messages.
func (l *logger) Subscribe(ctx context.Context) (<-chan resource.Event[Message], func()) {
	return l.broker.Subscribe(ctx)
}

type writer struct {
	*pubsub.Broker[Message]
}

func (lw *writer) Write(p []byte) (int, error) {
	lw.Publish(resource.CreatedEvent, Message(p))
	return len(p), nil
}
