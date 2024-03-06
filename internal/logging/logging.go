package logging

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

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

// Message is the event payload type
type Message struct {
	Time    time.Time
	Level   string
	Message string `json:"msg"`
	Error   string
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

	handler := slog.NewJSONHandler(
		logger,
		&slog.HandlerOptions{
			Level: slog.Level(levels[level]),
		},
	)
	logger.Logger = slog.New(handler)

	return logger
}

func (l *Logger) Write(p []byte) (int, error) {
	var msg Message
	if err := json.Unmarshal(p, &msg); err != nil {
		return 0, err
	}

	l.Messages = append(l.Messages, msg)
	l.broker.Publish(resource.CreatedEvent, msg)
	if _, err := l.file.Write(p); err != nil {
		return 0, err
	}

	return len(p), nil
}

// Subscribe to log messages.
func (l *Logger) Subscribe(ctx context.Context) (<-chan resource.Event[Message], func()) {
	return l.broker.Subscribe(ctx)
}
