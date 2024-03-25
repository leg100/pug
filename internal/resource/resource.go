package resource

import (
	"log/slog"
)

// GlobalResource is the zero value of resource, representing the top-level
// entity to which all resources belong.
var GlobalResource = Resource{}

// Resource is an entity of a particular kind and occupies a particular position
// in the pug hierarchy w.r.t to other resources.
type Resource struct {
	// Resource optionally belongs to a parent.
	Parent *Resource

	// id uniquely identifies the resource.
	id ID
}

// Entity is a unique identifiable thing.
type Entity interface {
	ID() ID
}

func New(kind Kind, human string, parent *Resource) Resource {
	return Resource{
		id:     newID(kind, human),
		Parent: parent,
	}
}

func (r Resource) ID() ID {
	return r.id
}

func (r Resource) Kind() Kind {
	return r.id.kind
}

// String is a human meaningful, or at least readable, identification of the
// resource.
func (r Resource) String() string {
	return r.id.String()
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
		if r.Kind() == k {
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
