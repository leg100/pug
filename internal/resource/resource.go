package resource

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

// GlobalResource is the zero value of resource, representing the top-level
// entity to which all resources belong.
var GlobalResource = Resource{}

// Resource is an entity of a particular kind and occupies a particular position
// in the pug hierarchy w.r.t to other resources.
type Resource struct {
	// Resource optionally belongs to a parent.
	Parent *Resource
	// Kind of resource, e.g. module, workspace, etc.
	Kind Kind

	// id uniquely identifies the resource.
	id ID
	// ident uniquely identifies the resource and is human meaningful; it takes
	// precedence over the id if non-empty.
	ident string
}

// Entity is a unique identifiable thing.
type Entity interface {
	ID() ID
}

func New(kind Kind, ident string, parent *Resource) Resource {
	return Resource{
		id:     ID(uuid.New()),
		Parent: parent,
		Kind:   kind,
		ident:  ident,
	}
}

func (r Resource) ID() ID {
	return r.id
}

// String is a human meaningful, or at least readable, identification of the
// resource.
func (r Resource) String() string {
	if r.ident != "" {
		return r.ident
	}
	encodedID := base58.Encode(r.id[:])
	kind := strings.ToLower(r.Kind.String())
	return fmt.Sprintf("%s-%s", kind, encodedID)
}

// Ancestors provides a list of successive parents, starting with the direct parents.
func (r Resource) Ancestors() (ancestors []Resource) {
	if r.Parent == nil {
		return
	}
	ancestors = append(ancestors, *r.Parent)
	return append(ancestors, r.Parent.Ancestors()...)
}

// HasAncestor checks whether the given id is an ancestor of the resource.
func (r Resource) HasAncestor(id ID) bool {
	// Every resource is considered an ancestor of the nil ID (equivalent to the
	// ID of the "site" or "global"...).
	if id == GlobalID {
		return true
	}
	if r.Parent == nil {
		return false
	}
	if r.Parent.id == id {
		return true
	}
	return r.Parent.HasAncestor(id)
}

func (r Resource) GetKind(k Kind) *Resource {
	for _, r := range append([]Resource{r}, r.Ancestors()...) {
		if r.Kind == k {
			return &r
		}
	}
	return nil
}

func (r Resource) Module() *Resource {
	return r.GetKind(Module)
}

func (r Resource) Workspace() *Resource {
	return r.GetKind(Workspace)
}

func (r Resource) Run() *Resource {
	return r.GetKind(Run)
}

func (r Resource) LogValue() slog.Value {
	return slog.StringValue(r.String())
}
