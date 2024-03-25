package resource

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

// GlobalID is the zero value of ID, representing the ID of the abstract
// top-level "global" entity to which all resources belong.
var GlobalID = ID{}

// IDEncodedMaxLen is the max length of an encoded ID (it can sometimes encode
// to something shorter).
const IDEncodedMaxLen = 27

type ID struct {
	id uuid.UUID
	// human uniquely identifies the resource and is human meaningful; it takes
	// precedence over the id if non-empty.
	human string
	// Kind of resource, e.g. module, workspace, etc.
	kind Kind
}

func newID(k Kind, human string) ID {
	return ID{
		id:    uuid.New(),
		human: human,
		kind:  k,
	}
}

func (id ID) String() string {
	if id.human != "" {
		return id.human
	}
	return fmt.Sprintf("%s-%s", id.kind.String(), base58.Encode(id.id[:]))
}

func (id ID) LogValue() slog.Value {
	return slog.StringValue(id.String())
}

func IDFromString(s string) (ID, error) {
	encKind, encID, found := strings.Cut(s, "-")
	if !found {
		return ID{}, fmt.Errorf("invalid identifier: %s", s)
	}

	// decode kind
	kind, err := kindString(encKind)
	if err != nil {
		return ID{}, fmt.Errorf("decoding identifier: %w", err)
	}

	// decode id
	decoded := base58.Decode(encID)
	id, err := uuid.ParseBytes(decoded)
	if err != nil {
		return ID{}, fmt.Errorf("decoding identifier: %w", err)
	}

	return ID{id: id, kind: kind}, nil
}
