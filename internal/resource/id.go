package resource

import (
	"log/slog"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

// GlobalID is the zero value of ID, representing the ID of the abstract
// top-level "global" entity to which all resources belong.
var GlobalID = ID(uuid.Nil)

// IDEncodedMaxLen is the max length of an encoded ID (it can sometimes encode
// to something shorter).
const IDEncodedMaxLen = 22

type ID uuid.UUID

func (id ID) String() string {
	return base58.Encode(id[:])
}

func (id ID) LogValue() slog.Value {
	return slog.StringValue(id.String())
}

func IDFromString(id string) (ID, error) {
	decoded := base58.Decode(id)
	raw, err := uuid.ParseBytes(decoded)
	if err != nil {
		return ID{}, nil
	}
	return ID(raw), nil
}
