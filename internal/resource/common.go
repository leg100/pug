package resource

// Common provides functionality common to all resources.
type Common struct {
	ID
	Parent Resource

	// direct dependencies
	dependencies []ID
}

func New(kind Kind, parent Resource) Common {
	return Common{
		ID:     NewID(kind),
		Parent: parent,
	}
}

func (r Common) Dependencies() []ID {
	return r.dependencies
}

func (r Common) WithDependencies(deps ...ID) Common {
	r.dependencies = deps
	return r
}
