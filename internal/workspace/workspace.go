package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	// Parent module info
	ModuleID   resource.Resource
	ModulePath string

	// Uniquely identifies workspace
	resource.Resource

	// Name of workspace
	Name string

	// True if workspace is the current workspace for the module.
	Current bool
}

func (ws *Workspace) String() string {
	return fmt.Sprintf("%s:%s", ws.ModulePath, ws.Name)
}

func (ws *Workspace) PugDirectory() string {
	return filepath.Join(ws.ModulePath, ".pug", ws.Name)
}
