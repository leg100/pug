package resource

import "time"

// Common provides functionality common to all resources.
type Common struct {
	ID
	Parent Resource

	createdAt time.Time
}

func New(kind Kind, parent Resource) Common {
	return Common{
		ID:        NewID(kind),
		Parent:    parent,
		createdAt: time.Now(),
	}
}

func (r Common) CreatedAt() time.Time {
	return r.createdAt
}

func (r Common) GetParent() Resource {
	return r.Parent
}

func (r Common) HasAncestor(id ID) bool {
	// Every entity is considered an ancestor of the nil ID (equivalent to the
	// ID of the "global" entity).
	if id == GlobalID {
		return true
	}
	if r.Parent == nil {
		// Parent has no parents, so go no further
		return false
	}
	if r.Parent.GetID() == id {
		return true
	}
	// Check parents of parent
	return r.Parent.HasAncestor(id)
}

// Ancestors provides a list of successive parents, starting with the direct parents.
func (r Common) Ancestors() (ancestors []Resource) {
	if r.Parent == nil {
		return
	}
	ancestors = append(ancestors, r.Parent)
	return append(ancestors, r.Parent.Ancestors()...)
}

func (r Common) getAncestorKind(k Kind) Resource {
	for _, parent := range r.Ancestors() {
		if parent.GetKind() == k {
			return parent
		}
	}
	return nil
}

func (r Common) Module() Resource {
	return r.getAncestorKind(Module)
}

func (r Common) Workspace() Resource {
	return r.getAncestorKind(Workspace)
}

func (r Common) Run() Resource {
	return r.getAncestorKind(Run)
}
