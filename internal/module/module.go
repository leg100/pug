package module

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
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
}

// New constructs a module. Workdir is the pug working directory, and path is
// the module path relative to the working directory.
func New(workdir internal.Workdir, path string) *Module {
	return &Module{
		Common:  resource.New(resource.Module, resource.GlobalResource),
		Path:    path,
		Workdir: workdir,
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

// findModules finds root modules that are descendents of the workdir and
// returns their paths relative to the workdir.
//
// A root module is deemed to be a directory that contains a .tf file that
// contains a backend block.
func findModules(logger logging.Interface, workdir internal.Workdir) (modules []string, err error) {
	found := make(map[string]struct{})
	walkfn := func(path string, d fs.DirEntry, walkerr error) error {
		// skip directories that have already been identified as containing a
		// root module
		if _, ok := found[filepath.Dir(path)]; ok {
			return nil
		}
		if walkerr != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".terraform", ".terragrunt-cache":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "terragrunt.hcl" {
			found[filepath.Dir(path)] = struct{}{}
			// skip walking remainder of parent directory
			return fs.SkipDir
		}
		if filepath.Ext(path) != ".tf" {
			return nil
		}
		cfg, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// only the hclwrite pkg seems to have the ability to walk HCL blocks,
		// so this is what is used even though no writing is performed.
		file, diags := hclwrite.ParseConfig(cfg, path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			logger.Error("reloading modules: parsing hcl", "path", path, "error", diags)
			return nil
		}
		for _, block := range file.Body().Blocks() {
			if block.Type() == "terraform" {
				for _, nested := range block.Body().Blocks() {
					if nested.Type() == "backend" || nested.Type() == "cloud" {
						found[filepath.Dir(path)] = struct{}{}
						// skip walking remainder of parent directory
						return fs.SkipDir
					}
				}
			}
		}
		return nil
	}
	if err := filepath.WalkDir(workdir.String(), walkfn); err != nil {
		return nil, err
	}
	// Strip parent prefix from paths before returning
	modules = make([]string, len(found))
	for i, f := range maps.Keys(found) {
		stripped, err := filepath.Rel(workdir.String(), f)
		if err != nil {
			return nil, err
		}
		modules[i] = stripped
	}
	return modules, nil
}
