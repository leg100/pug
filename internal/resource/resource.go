package resource

import (
	"log/slog"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

var NilResource = Resource{}

// Resource is a unique pug entity.
type Resource struct {
	ID ID
	// Resource optionally belongs to a parent.
	Parent *Resource
	// Kind of resource, e.g. module, workspace, etc.
	Kind Kind

	// ident is a human meaningful identification of the resource.
	ident string
}

func New(kind Kind, ident string, parent *Resource) Resource {
	return Resource{
		ID:     ID(uuid.New()),
		Parent: parent,
		Kind:   kind,
		ident:  ident,
	}
}

// String is a human meaningful, or at least readable, identification of the
// resource.
func (r Resource) String() string {
	if r.ident != "" {
		return r.ident
	}
	return base58.Encode(r.ID[:])
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
	if r.Parent == nil {
		return false
	}
	if r.Parent.ID == id {
		return true
	}
	return r.Parent.HasAncestor(id)
}

func (r Resource) Module() *Resource {
	for _, r := range append([]Resource{r}, r.Ancestors()...) {
		if r.Kind == Module {
			return &r
		}
	}
	return nil
}

func (r Resource) Workspace() *Resource {
	for _, r := range append([]Resource{r}, r.Ancestors()...) {
		if r.Kind == Workspace {
			return &r
		}
	}
	return nil
}

func (r Resource) LogValue() slog.Value {
	return slog.StringValue(r.String())
}
