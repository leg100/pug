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

// MonotonicID is an identifier based on an ever-increasing serial number, and a
// kind to differentiate it from other kinds of identifiers.
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

// String provides a human readable representation of the identifier.
func (id MonotonicID) String() string {
	return fmt.Sprintf("#%d", id.Serial)
}
