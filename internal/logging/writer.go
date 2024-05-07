package logging

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/leg100/pug/internal/resource"
)

// writer is a slog TextHandler writer that both keeps the log records in
// memory and emits and them as pug events.
type writer struct {
	table  *resource.Table[Message]
	serial uint
}

func (w *writer) Write(p []byte) (int, error) {
	d := logfmt.NewDecoder(bytes.NewReader(p))
	for d.ScanRecord() {
		msg := Message{
			Serial:   w.serial,
			Resource: resource.New(resource.Log, resource.GlobalResource),
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
				msg.Attributes = append(msg.Attributes, Attr{
					Key:   string(d.Key()),
					Value: string(d.Value()),
				})
			}
		}
		w.table.Add(msg.ID, msg)
		w.serial++
	}
	if d.Err() != nil {
		return 0, d.Err()
	}
	return len(p), nil
}
