package logging

import (
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/lmittmann/tint"
	"golang.org/x/exp/maps"
)

const DefaultLevel = "info"

var levels = map[string]slog.Level{
	"debug":      slog.LevelDebug,
	DefaultLevel: slog.LevelInfo,
	"warn":       slog.LevelWarn,
	"error":      slog.LevelError,
}

// ValidLevels returns valid strings for choosing a log level. Returns the
// default log level first.
func ValidLevels() []string {
	keys := maps.Keys(levels)
	slices.SortFunc(keys, func(a, b string) int {
		if a == DefaultLevel {
			return -1
		}
		if b == DefaultLevel {
			return 1
		}
		// Sort remaining in alphabetical order.
		if a < b {
			return -1
		}
		return 1
	})
	return keys
}

// NewLogger constructs Logger, a slog wrapper with additional functionality.
func NewLogger(opts Options) *Logger {
	level := slog.Level(levels[opts.Level])
	logger := &Logger{
		enricher: &enricher{},
	}
	if opts.TUI {
		logger.Broker = pubsub.NewBroker[Message](logger)
		logger.writer = &writer{table: resource.NewTable(logger.Broker)}
		logger.logger = slog.New(slog.NewTextHandler(
			io.MultiWriter(append(opts.AdditionalWriters, logger.writer)...),
			&slog.HandlerOptions{Level: level},
		))
	} else {
		logger.logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      level,
				TimeFormat: time.Kitchen,
			}),
		)
	}
	return logger
}

type Options struct {
	// The log level of the logger
	Level string
	// TUI toggles TUI mode: logs are emitted as Pug events and persisted in
	// memory for later retrieval. When set to false, logs are instead written
	// directly to stderr.
	TUI bool
	// Any additional writers the log handler should write to. Only takes effect
	// when in TUI mode.
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
