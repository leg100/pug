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
}

func New(kind Kind, parent *Resource) Resource {
	return Resource{
		ID:     uuid.New(),
		Parent: parent,
	}
}

// ID provides a concise human readable ID
func (r Resource) String() string {
	return base58.Encode(r.ID[:])
}

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
