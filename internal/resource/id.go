package resource

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

type ID uuid.UUID

func (id ID) String() string {
	return base58.Encode(id[:])
}

func IDFromString(id string) (ID, error) {
	decoded := base58.Decode(id)
	raw, err := uuid.ParseBytes(decoded)
	if err != nil {
		return ID{}, nil
	}
	return ID(raw), nil
}
