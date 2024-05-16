package workspace

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	resource.Common

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
		Common: resource.New(resource.Workspace, mod),
		Name:   name,
	}, nil
}

func (ws *Workspace) String() string {
	return ws.Name
}

func (ws *Workspace) ModuleID() resource.ID {
	return ws.Parent.GetID()
}

func (ws *Workspace) ModulePath() string {
	return ws.Parent.String()
}

func (ws *Workspace) TerraformEnv() string {
	return fmt.Sprintf("TF_WORKSPACE=%s", ws.Name)
}

func (ws *Workspace) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", ws.Name),
	)
}
