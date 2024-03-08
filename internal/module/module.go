package module

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/leg100/pug/internal/resource"
	"golang.org/x/exp/maps"
)

type Module struct {
	resource.Resource

	// Name of its current workspace
	Current string
}

// Path is the path to the module relative to the pug working directory. The
// path also uniquely identifies a module.
func New(path string) *Module {
	return &Module{
		Resource: resource.New(resource.Module, path, nil),
	}
}

func (m *Module) Path() string { return m.String() }

// findModules finds root modules that are descendents of the given path and
// returns their paths. Determining what is a root module is difficult and
// relies on a set of heuristics:
//
// (a) check if path contains a .terraform directory, else fallback to:
// (b) path has a .tf file containing a backend block
//
// (a) will only succeed if the module has already been initialized, i.e. terraform
// init has been run, whereas (b) is necessary if it has not.
func findModules(parent string) (modules []string, err error) {
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
			if d.Name() == ".terraform" {
				found[filepath.Dir(path)] = struct{}{}
				// skip walking .terraform/
				return filepath.SkipDir
			}
			return nil
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
			slog.Error("finding modules: parsing hcl", "path", path, "error", diags)
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
	if err := filepath.WalkDir(parent, walkfn); err != nil {
		return nil, err
	}
	// Strip parent prefix from paths before returning
	modules = make([]string, len(found))
	for i, f := range maps.Keys(found) {
		stripped, err := filepath.Rel(parent, f)
		if err != nil {
			return nil, err
		}
		modules[i] = stripped
	}
	return modules, nil
}
