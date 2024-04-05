package state

// Resource is a pug state resource.
type Resource struct {
	Address ResourceAddress
	Status  ResourceStatus
}

type ResourceStatus string

type ResourceAddress string

const (
	// Idle means the resource is idle (no tasks are currently operating on
	// it).
	Idle ResourceStatus = "idle"
	// Removing means the resource is in the process of being removed.
	Removing ResourceStatus = "removing"
	// Tainting means the resource is in the process of being tainted.
	Tainting ResourceStatus = "tainting"
	// Tainted means the resource is currently tainted
	Tainted ResourceStatus = "tainted"
)
