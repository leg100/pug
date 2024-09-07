package resource

import (
	"fmt"
	"sync"
)

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

func (id ID) String() string {
	return fmt.Sprintf("#%d", id.Serial)
}

// GetID allows ID to be accessed via an interface value.
func (id ID) GetID() ID {
	return id
}

// GetKind allows Kind to be accessed via an interface value.
func (id ID) GetKind() Kind {
	return id.Kind
}
