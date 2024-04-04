package state

// Resource is a pug state resource.
type Resource struct {
	Address ResourceAddress
	Status  ResourceStatus
}

type ResourceStatus string

const (
	// Idle means the resource is idle (no tasks are currently operating on
	// it).
	Idle ResourceStatus = "idle"
	// Removing means the resource is in the process of being removed.
	Removing = "removing"
	// Tainting means the resource is in the process of being tainted.
	Tainting = "tainting"
	// Tainted means the resource is currently tainted
	Tainted = "tainted"
)

func newResource(sfr StateFileResource) *Resource {
	return &Resource{
		Address: ResourceAddress{
			name: sfr.Name,
			typ:  sfr.Type,
		},
		Status: Idle,
	}
}

// ResourceAddress is the path for a terraform resource, i.e. its type and name.
type ResourceAddress struct {
	typ  string
	name string
}

func (p ResourceAddress) String() string {
	return p.typ + "." + p.name
}
