package internal

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type module struct {
	path        string
	initialized bool
}

func (m module) String() string      { return m.path }
func (m module) Title() string       { return m.path }
func (m module) Description() string { return m.path }
func (m module) FilterValue() string { return m.path }

func (m module) init(runner *runner) (*task, error) {
	task, err := runner.run(taskspec{
		prog:      "tofu",
		args:      []string{"init"},
		path:      m.path,
		exclusive: isPluginCacheUsed(),
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

// findModules finds root modules that are descendents of the given path and
// returns their paths. Determining what is a root module is difficult and
// relies on a set of heuristics:
//
// (a) check if path contains a .terraform directory, else fallback to:
// (b) path has a .tf file containing a backend block
//
// (a) will only succeed if the module has already been initialized, i.e. terraform
// init has been run, whereas (b) is necessary if it has not.
func findModules(parent string) (modules []module, err error) {
	err = filepath.WalkDir(parent, func(path string, d fs.DirEntry, walkerr error) error {
		// skip files in directories that have already been identified as containing a
		// root module
		if slices.ContainsFunc(modules, func(m module) bool {
			return filepath.Dir(path) == m.path
		}) {
			return nil
		}
		if walkerr != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".terraform" {
				modules = append(modules, module{
					path:        filepath.Dir(path),
					initialized: true,
				})
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
		f, diags := hclwrite.ParseConfig(cfg, path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			slog.Error("finding modules: parsing hcl", "path", path, "error", diags)
			return nil
		}
		for _, block := range f.Body().Blocks() {
			if block.Type() == "terraform" {
				for _, nested := range block.Body().Blocks() {
					if nested.Type() == "backend" || nested.Type() == "cloud" {
						modules = append(modules, module{
							path: filepath.Dir(path),
						})
						return nil
					}
				}
			}
		}
		return nil
	})
	return
}
