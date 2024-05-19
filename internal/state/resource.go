package state

import "github.com/leg100/pug/internal/resource"

// Resource is a pug state resource.
type Resource struct {
	resource.Common

	Address ResourceAddress
	Tainted bool
}

func newResource(ws resource.Resource, addr ResourceAddress) *Resource {
	return &Resource{
		Common:  resource.New(resource.StateResource, ws),
		Address: addr,
	}
}

type ResourceAddress string
