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

func newWorkspace(module resource.Resource, name string, current bool) *Workspace {
	return &Workspace{
		Resource: resource.New(resource.Workspace, name, &module),
		Current:  current,
	}
}

func (ws *Workspace) TerraformEnv() string {
	return fmt.Sprintf("TF_WORKSPACE=%s", ws)
}

func PugDirectory(path, name string) string {
	return filepath.Join(path, ".pug", name)
}
