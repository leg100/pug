package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	// Uniquely identifies workspace and the module it belongs to.
	resource.Resource

	// The workspace's current or last active run.
	CurrentRun *resource.Resource

	AutoApply bool
}

func New(module resource.Resource, name string) *Workspace {
	return &Workspace{
		Resource: resource.New(resource.Workspace, name, &module),
	}
}

func (ws *Workspace) TerraformEnv() string {
	return TerraformEnv(ws.String())
}

func (ws *Workspace) Name() string {
	return ws.String()
}

func (ws *Workspace) Module() resource.Resource {
	return *ws.Parent
}

func (ws *Workspace) ModulePath() string {
	return ws.Parent.String()
}

func (ws *Workspace) PugDirectory() string {
	return PugDirectory(ws.String())
}

func PugDirectory(name string) string {
	return filepath.Join(".pug", name)
}

func TerraformEnv(name string) string {
	return fmt.Sprintf("TF_WORKSPACE=%s", name)
}
