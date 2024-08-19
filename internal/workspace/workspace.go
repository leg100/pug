package workspace

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
)

type Workspace struct {
	resource.Common

	Name string
	Cost string
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
	return TerraformEnv(ws.Name)
}

func (ws *Workspace) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", ws.Name),
	)
}

// VarsFile returns the filename of the workspace's terraform variables file
// and whether it exists or not.
func (ws *Workspace) VarsFile(workdir internal.Workdir) (string, bool) {
	fname := fmt.Sprintf("%s.tfvars", ws.Name)
	mod := ws.Module().(*module.Module)
	path := filepath.Join(workdir.String(), mod.Path, fname)
	_, err := os.Stat(path)
	return fname, err == nil
}

func TerraformEnv(workspaceName string) string {
	return fmt.Sprintf("TF_WORKSPACE=%s", workspaceName)
}
