package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"log/slog"

	"github.com/go-logfmt/logfmt"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// Message is the event payload for a log message
type Message struct {
	Time       time.Time
	Level      string
	Message    string `json:"msg"`
	Attributes map[string]string

	// id uniquely identifies the log message.
	id resource.ID
}

func (r Message) ID() resource.ID {
	return r.id
}

// HasAncestor is only implemented in order to satisfy the tableItem interface
// used in the tui table
func (r Message) HasAncestor(id resource.ID) bool {
	return true
}

type Logger struct {
	Logger   *slog.Logger
	Messages []Message

	broker *pubsub.Broker[Message]
	file   io.Writer
}

// NewLogger constructs a slog logger with the appropriate log level. The logger
// emits messages as pug events.
func NewLogger(level string) *Logger {
	f, _ := os.OpenFile("pug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)

	logger := &Logger{
		broker: pubsub.NewBroker[Message](),
		file:   f,
	}

	handler := slog.NewTextHandler(
		logger,
		&slog.HandlerOptions{
			Level: slog.Level(levels[level]),
		},
	)
	logger.Logger = slog.New(handler)

	return logger
}

func (l *Logger) Write(p []byte) (int, error) {
	msgs := make([]Message, 0, 1)
	d := logfmt.NewDecoder(bytes.NewReader(p))
	for d.ScanRecord() {
		msg := Message{
			id: resource.ID(uuid.New()),
		}
		for d.ScanKeyval() {
			switch string(d.Key()) {
			case "time":
				parsed, err := time.Parse(time.RFC3339, string(d.Value()))
				if err != nil {
					return 0, fmt.Errorf("parsing time: %w", err)
				}
				msg.Time = parsed
			case "level":
				msg.Level = string(d.Value())
			case "msg":
				msg.Message = string(d.Value())
			default:
				if msg.Attributes == nil {
					msg.Attributes = make(map[string]string)
				}
				msg.Attributes[string(d.Key())] = string(d.Value())
			}
		}
		msgs = append(msgs, msg)
		l.broker.Publish(resource.CreatedEvent, msg)
	}
	if d.Err() != nil {
		return 0, d.Err()
	}
	l.Messages = append(l.Messages, msgs...)
	if _, err := l.file.Write(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Subscribe to log messages.
func (l *Logger) Subscribe(ctx context.Context) (<-chan resource.Event[Message], func()) {
	return l.broker.Subscribe(ctx)
}
