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
	// Kind of resource, e.g. module, workspace, etc.
	Kind Kind

	// ident is a human meaningful identification of the resource.
	ident string
}

func New(kind Kind, ident string, parent *Resource) Resource {
	return Resource{
		ID:     uuid.New(),
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
	getAncestors(r, ancestors)
	return
}

func getAncestors(resource Resource, ancestors []Resource) {
	if resource.Parent == nil {
		return
	}
	getAncestors(*resource.Parent, append(ancestors, *resource.Parent))
}

// HasAncestor checks whether the given id is an ancestor of the resource.
func (r Resource) HasAncestor(id uuid.UUID) bool {
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
