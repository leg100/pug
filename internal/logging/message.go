package logging

import (
	"time"

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

	// A message is a pug resource, but only insofar as it makes it easier to
	// handle consistently alongside all other resources (modules, workspaces,
	// etc) in the TUI.
	resource.Common
}

type Attr struct {
	Key   string
	Value string

	// An attribute is a pug resource, but only insofar as it makes it easier to
	// handle consistently alongside all other resources (modules, workspaces,
	// etc) in the TUI.
	resource.Common
}
