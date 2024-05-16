package resource

import (
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

// GlobalID is the zero value of ID, representing the ID of the abstract
// top-level "global" entity to which all resources belong.
var GlobalID = ID{}

// IDEncodedMaxLen is the max length of an encoded ID (it can sometimes encode
// to something shorter).
const IDEncodedMaxLen = 27

// ID is a unique identifier for a pug entity.
type ID struct {
	id   uuid.UUID
	Kind Kind
}

func NewID(kind Kind) ID {
	return ID{
		id:   uuid.New(),
		Kind: kind,
	}
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s", id.Kind.String(), base58.Encode(id.id[:]))
}

// GetID allows ID to be accessed via an interface value.
func (id ID) GetID() ID {
	return id
}

// GetKind allows Kind to be accessed via an interface value.
func (id ID) GetKind() Kind {
	return id.Kind
}
