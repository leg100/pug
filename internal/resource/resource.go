package resource

var GlobalResource = Resource{}

type Resource struct {
	ID
	Parent *Resource
}

func New(kind Kind, parent Resource) Resource {
	return Resource{
		ID:     NewID(kind),
		Parent: &parent,
	}
}

func (r Resource) HasAncestor(id ID) bool {
	// Every entity is considered an ancestor of the nil ID (equivalent to the
	// ID of the "global" entity).
	if id == GlobalID {
		return true
	}
	if r.Parent == nil {
		// Parent has no parents, so go no further
		return false
	}
	if r.Parent.ID == id {
		return true
	}
	// Check parents of parent
	return r.Parent.HasAncestor(id)
}

// Ancestors provides a list of successive parents, starting with the direct parents.
func (r Resource) Ancestors() (ancestors []Resource) {
	if r.Parent == nil {
		return
	}
	ancestors = append(ancestors, *r.Parent)
	return append(ancestors, r.Parent.Ancestors()...)
}

func (r Resource) getAncestorKind(k Kind) *Resource {
	for _, par := range r.Ancestors() {
		if par.Kind == k {
			return &par
		}
	}
	return nil
}

func (r Resource) Module() *Resource {
	return r.getAncestorKind(Module)
}

func (r Resource) Workspace() *Resource {
	return r.getAncestorKind(Workspace)
}

func (r Resource) Run() *Resource {
	return r.getAncestorKind(Run)
}

// RowParent implements tui/table.ResourceValue
func (r Resource) RowParent() *Resource {
	return r.Parent
}
