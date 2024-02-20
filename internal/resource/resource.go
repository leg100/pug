package resource

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

type Resource interface {
	ID() string
}

// ID uniquely identifies a resource.
type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

// ID provides a concise human readable ID
func (id ID) ID() string {
	return base58.Encode(id[:])
}
