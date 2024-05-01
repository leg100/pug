package workspace

import (
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	resource.Resource

	Name string

	// The workspace's current or last active run.
	CurrentRunID *resource.ID

	AutoApply bool
}

func New(mod *module.Module, name string) (*Workspace, error) {
	if name != url.PathEscape(name) {
		return nil, fmt.Errorf("invalid workspace name: %s", name)
	}
	return &Workspace{
		Resource: resource.New(resource.Workspace, mod.Resource),
		Name:     name,
	}, nil
}

func (ws *Workspace) ModuleID() resource.ID {
	return ws.Parent.ID
}

func (ws *Workspace) TerraformEnv() string {
	return fmt.Sprintf("TF_WORKSPACE=%s", ws.Name)
}

// PugDirectory returns the workspace's pug directory, relative to its module
// directory.
func (ws *Workspace) PugDirectory() string {
	return filepath.Join(internal.PugDirectory, ws.Name)
}

func (ws *Workspace) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", ws.Name),
	)
}
