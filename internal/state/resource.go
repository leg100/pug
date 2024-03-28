package state

import "github.com/leg100/pug/internal/resource"

// Resource is a pug state resource.
type Resource struct {
	// resource.Resource fully identifies Resource, including the workspace it
	// belongs to.
	resource.Resource
}

func newResource(ws resource.Resource, sfr StateFileResource) *Resource {
	path := ResourcePath{
		name: sfr.Name,
		typ:  sfr.Type,
	}
	return &Resource{
		Resource: resource.New(resource.StateResource, path.String(), &ws),
	}
}

// ResourcePath is the path for a terraform resource, i.e. its type and name.
type ResourcePath struct {
	typ  string
	name string
}

func (p ResourcePath) String() string {
	return p.typ + "." + p.name
}
