package logging

import (
	"time"

	"github.com/leg100/pug/internal/resource"
)

// Message is the event payload for a log message
type Message struct {
	ID         resource.MonotonicID
	Time       time.Time
	Level      string
	Message    string `json:"msg"`
	Attributes []Attr
}

func (m Message) GetID() resource.ID { return m.ID }

type Attr struct {
	ID    resource.MonotonicID
	Key   string
	Value string
}

func (a Attr) GetID() resource.ID { return a.ID }
