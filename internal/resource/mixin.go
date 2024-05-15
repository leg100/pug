package resource

// GlobalResource is an abstract top-level pug resource from which all other pug
// resources are spawned.
var GlobalResource Resource = Mixin{}

// Mixin is incorporated into pug resources to provide functionality common to
// all resources.
type Mixin struct {
	ID
	Parent Resource
}

func New(kind Kind, parent Resource) Mixin {
	return Mixin{
		ID:     NewID(kind),
		Parent: parent,
	}
}

func (r Mixin) GetParent() Resource {
	return r.Parent
}

func (r Mixin) HasAncestor(id ID) bool {
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
func (r Mixin) Ancestors() (ancestors []Resource) {
	if r.Parent == nil {
		return
	}
	ancestors = append(ancestors, r.Parent)
	return append(ancestors, r.Parent.Ancestors()...)
}

func (r Mixin) getAncestorKind(k Kind) Resource {
	for _, parent := range r.Ancestors() {
		if parent.GetKind() == k {
			return parent
		}
	}
	return nil
}

func (r Mixin) Module() Resource {
	return r.getAncestorKind(Module)
}

func (r Mixin) Workspace() Resource {
	return r.getAncestorKind(Workspace)
}

func (r Mixin) Run() Resource {
	return r.getAncestorKind(Run)
}
