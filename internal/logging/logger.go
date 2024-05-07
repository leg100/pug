package logging

import (
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
	writer := &writer{table: resource.NewTable(broker)}

	handler := slog.NewTextHandler(
		io.MultiWriter(append(opts.AdditionalWriters, writer)...),
		&slog.HandlerOptions{
			Level: slog.Level(levels[opts.Level]),
		},
	)

	logger.logger = slog.New(handler)
	logger.Broker = broker
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
	writer *writer

	*pubsub.Broker[Message]
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

// List lists the log messages received thus far.
func (l *Logger) List() []Message {
	return l.writer.table.List()
}

// Get retrieves a log message by ID.
func (l *Logger) Get(id resource.ID) (Message, error) {
	return l.writer.table.Get(id)
}
