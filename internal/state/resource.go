package state

import (
	"encoding/json"

	"github.com/leg100/pug/internal/resource"
)

// Resource is a pug state resource.
type Resource struct {
	resource.Common

	WorkspaceID resource.ID
	Address     ResourceAddress
	Attributes  map[string]any
	Tainted     bool
}

func (r *Resource) String() string {
	return string(r.Address)
}

func newResource(workspaceID resource.ID, addr ResourceAddress, attrs json.RawMessage) (*Resource, error) {
	res := &Resource{
		Common:      resource.New(resource.StateResource, resource.GlobalResource),
		WorkspaceID: workspaceID,
		Address:     addr,
	}
	if err := json.Unmarshal(attrs, &res.Attributes); err != nil {
		return nil, err
	}
	return res, nil
}

type ResourceAddress string
