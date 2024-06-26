package module

import (
	"log/slog"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
)

// Module is a terraform root module.
type Module struct {
	resource.Common

	// Pug working directory
	Workdir internal.Workdir

	// Path relative to pug working directory
	Path string

	// The module's current workspace.
	CurrentWorkspaceID *resource.ID

	// Modules that depend on this module. Only valid when using terragrunt.
	Dependencies []*Module
}

// New constructs a module. Workdir is the pug working directory, and path is
// the module path relative to the working directory.
func New(workdir internal.Workdir, path string, deps []string) *Module {
	return &Module{
		Common:       resource.New(resource.Module, resource.GlobalResource),
		Path:         path,
		Workdir:      workdir,
		Dependencies: deps,
	}
}

func (m *Module) String() string {
	return m.Path
}

// FullPath returns the absolute path to the module.
func (m *Module) FullPath() string {
	return filepath.Join(m.Workdir.String(), m.Path)
}

func (m *Module) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("path", m.Path),
	)
}
