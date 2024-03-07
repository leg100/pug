package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	// Uniquely identifies workspace and the module it belongs to.
	resource.Resource

	// True if workspace is the current workspace for the module.
	Current bool
}

func New(module resource.Resource, name string, current bool) *Workspace {
	return &Workspace{
		Resource: resource.New(resource.Workspace, name, &module),
		Current:  current,
	}
}

func (ws *Workspace) TerraformEnv() string {
	return fmt.Sprintf("TF_WORKSPACE=%s", ws)
}

func (ws *Workspace) Name() string {
	return ws.String()
}

func (ws *Workspace) Module() resource.Resource {
	return *ws.Parent
}

func (ws *Workspace) PugDirectory() string {
	return PugDirectory(ws.String())
}

func PugDirectory(name string) string {
	return filepath.Join(".pug", name)
}
