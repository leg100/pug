package logging

import (
	"context"
	"io"
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
func NewLogger(opts Options) *Logger {
	logger := &Logger{}
	broker := pubsub.NewBroker[Message](logger)
	writer := &writer{broker: broker}

	handler := slog.NewTextHandler(
		io.MultiWriter(append(opts.AdditionalWriters, writer)...),
		&slog.HandlerOptions{
			Level: slog.Level(levels[opts.Level]),
		},
	)

	logger.logger = slog.New(handler)
	logger.broker = broker
	logger.writer = writer
	logger.enricher = &enricher{}

	return logger
}

type Options struct {
	// The log level of the logger
	Level string
	// Any additional writers the log handler should write to.
	AdditionalWriters []io.Writer
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
