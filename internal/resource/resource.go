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
	// String is a human-readable identifier for the resource. Not necessarily
	// unique across pug.
	String() string
	// Dependencies returns both direct dependencies and indirect dependencies that
	// ancestor resources have.
	Dependencies() []ID
}
