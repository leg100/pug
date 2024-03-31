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
	Attributes []Attr

	// Serial uniquely identifies the message (within the scope of the logger it
	// was emitted from). The higher the Serial number the newer the message.
	Serial uint
}

type Attr struct {
	Key   string
	Value string
}

type Logger struct {
	Logger   *slog.Logger
	Messages []Message

	broker *pubsub.Broker[Message]
	file   io.Writer

	serial uint
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
		msg := Message{Serial: l.serial}
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
				msg.Attributes = append(msg.Attributes, Attr{
					Key:   string(d.Key()),
					Value: string(d.Value()),
				})
			}
		}
		msgs = append(msgs, msg)
		l.broker.Publish(resource.CreatedEvent, msg)
		l.serial++
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
func (l *Logger) Subscribe(ctx context.Context) <-chan resource.Event[Message] {
	return l.broker.Subscribe(ctx)
}

func BySerialDesc(i, j Message) int {
	if i.Serial < j.Serial {
		return 1
	}
	return -1
}
