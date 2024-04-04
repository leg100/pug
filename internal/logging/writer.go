package logging

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
)

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

// writer is a slog TextHandler writer that both keeps the log records in
// memory and emits and them as pug events.
type writer struct {
	Messages []Message

	broker *pubsub.Broker[Message]
	serial uint
}

func (b *writer) Write(p []byte) (int, error) {
	msgs := make([]Message, 0, 1)
	d := logfmt.NewDecoder(bytes.NewReader(p))
	for d.ScanRecord() {
		msg := Message{Serial: b.serial}
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
		b.broker.Publish(resource.CreatedEvent, msg)
		b.serial++
	}
	if d.Err() != nil {
		return 0, d.Err()
	}
	b.Messages = append(b.Messages, msgs...)
	return len(p), nil
}
