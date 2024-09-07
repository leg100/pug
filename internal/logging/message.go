package logging

import (
	"time"

	"github.com/leg100/pug/internal/resource"
)

// Message is the event payload for a log message
type Message struct {
	// A message is a pug resource, but only insofar as it makes it easier to
	// handle consistently alongside all other resources (modules, workspaces,
	// etc) in the TUI.
	resource.ID

	Time       time.Time
	Level      string
	Message    string `json:"msg"`
	Attributes []Attr
}

type Attr struct {
	Key   string
	Value string

	// An attribute is a pug resource, but only insofar as it makes it easier to
	// handle consistently alongside all other resources (modules, workspaces,
	// etc) in the TUI.
	resource.ID
}
