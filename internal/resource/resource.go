package resource

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

// Resource is a unique pug entity.
type Resource struct {
	ID uuid.UUID
	// Resource optionally belongs to a parent.
	Parent *Resource
}

func New(parent *Resource) Resource {
	return Resource{
		ID:     uuid.New(),
		Parent: parent,
	}
}

// ID provides a concise human readable ID
func (r Resource) String() string {
	return base58.Encode(r.ID[:])
}

func FromString(s string) (Resource, error) {
	decoded := base58.Decode(s)
	raw, err := uuid.ParseBytes(decoded)
	if err != nil {
		return Resource{}, err
	}
	return Resource(raw), nil
}
