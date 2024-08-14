package resource

// GlobalResource is an abstract top-level pug resource from which all other pug
// resources are ultimately spawned.
var GlobalResource Resource = Common{}

// Resource is a unique pug entity spawned from another entity.
type Resource interface {
	// GetID retrieves the unique identifier for the resource.
	GetID() ID
	// GetKind retrieves the kind of resource.
	GetKind() Kind
	// GetParent retrieves the resource's parent, the resource from which the
	// resource was spawned.
	GetParent() Resource
	// HasAncestor determines whether the resource has an ancestor with the
	// given ID.
	HasAncestor(ID) bool
	// Ancestors retrieves a list of the resource's ancestors, nearest first.
	Ancestors() []Resource
	// String is a human-readable identifier for the resource. Not necessarily
	// unique across pug.
	String() string
	// Module retrieves the resource's module. Returns nil if the resource does
	// not have a module ancestor.
	Module() Resource
	// Workspace retrieves the resource's workspcae. Returns nil if the resource does
	// not have a workspace ancestor.
	Workspace() Resource
	// Dependencies returns both direct dependencies and indirect dependencies that
	// ancestor resources have.
	Dependencies() []ID
}
