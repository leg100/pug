package logging

import (
	"context"
	"log/slog"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// NewLogger constructs Logger, a slog wrapper with additional functionality.
func NewLogger(level string) *Logger {
	broker := pubsub.NewBroker[Message]()
	writer := &writer{broker: broker}

	handler := slog.NewTextHandler(
		writer,
		&slog.HandlerOptions{
			Level: slog.Level(levels[level]),
		},
	)
	return &Logger{
		logger:   slog.New(handler),
		broker:   broker,
		writer:   writer,
		enricher: &enricher{},
	}
}

// Logger wraps slog, providing further functionality such as emitting log
// records as pug events, and enriching records with further attributes.
type Logger struct {
	logger *slog.Logger
	broker *pubsub.Broker[Message]
	writer *writer

	*enricher
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, l.enrich(args...)...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, l.enrich(args...)...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, l.enrich(args...)...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, l.enrich(args...)...)
}

// Subscribe to log messages.
func (l *Logger) Subscribe(ctx context.Context) <-chan resource.Event[Message] {
	return l.broker.Subscribe(ctx)
}

// Messages provides the log messages received thus far.
func (l *Logger) Messages() []Message {
	return l.writer.Messages
}
