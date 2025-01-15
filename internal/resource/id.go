package resource

import (
	"fmt"
	"sync"
)

// Identity is anything that uniquely differentiates a resource from another.
type Identity any

type Identifiable interface {
	GetID() Identity
}

var (
	// nextID provides the next ID for each kind
	nextID map[Kind]uint = make(map[Kind]uint)
	mu     sync.Mutex
)

// ID is a unique identifier for a pug resource.
type ID struct {
	Serial uint
	Kind   Kind
}

func NewID(kind Kind) ID {
	mu.Lock()
	defer mu.Unlock()

	id := nextID[kind]
	nextID[kind]++

	return ID{
		Serial: id,
		Kind:   kind,
	}
}

// String provides a human readable description.
func (id ID) String() string {
	return fmt.Sprintf("#%d", id.Serial)
}

// GetID implements Identifiable
func (id ID) GetID() Identity {
	return id
}
