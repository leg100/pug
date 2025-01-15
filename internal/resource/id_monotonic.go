package resource

import (
	"fmt"
	"sync"
)

var (
	// nextMonotonicID provides the next monotonic ID for each kind
	nextMonotonicID map[Kind]uint = make(map[Kind]uint)
	mu              sync.Mutex
)

// MonotonicID is a unique identifier for a pug resource.
type MonotonicID struct {
	Serial uint
	Kind   Kind
}

func NewMonotonicID(kind Kind) MonotonicID {
	mu.Lock()
	defer mu.Unlock()

	id := nextMonotonicID[kind]
	nextMonotonicID[kind]++

	return MonotonicID{
		Serial: id,
		Kind:   kind,
	}
}

// String provides a human readable description.
func (id MonotonicID) String() string {
	return fmt.Sprintf("#%d", id.Serial)
}

// GetID implements Identifiable
func (id MonotonicID) GetID() ID {
	return id
}
