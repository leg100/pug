package resource

// Resource is a unique pug entity
type Resource interface {
	// GetID retrieves the unique identifier for the resource.
	GetID() MonotonicID
	// String is a human-readable identifier for the resource. Not necessarily
	// unique across pug.
	String() string
}
