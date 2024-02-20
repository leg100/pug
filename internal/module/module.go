package module

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/leg100/pug/internal/task"
)

type Status string

const (
	// Module does not have a .terraform dir
	Uninitialized Status = "uninitalized"
	// Terraform init is being run
	Initializing Status = "initializing"
	// Terraform init ran successfully
	Initialized Status = "initialized"
	// Terraform init ran unsuccessfully
	Misconfigured Status = "misconfigured"
	// TODO: handle state in which a command has been run, i.e. terraform plan
	// or terraform workspace list, but it failed with error indicating the
	// module needs re-initalizing.
)

type Module struct {
	// Path is the path to the module relative to the pug working directory. The
	// path uniquely identifies a module.
	Path string

	Status Status

	// call this whenever state is updated
	callback func(*Module)
}

type factory struct {
	program  string
	callback func(*Module)
}

func (f *factory) newModule(path string, init bool) (*Module, error) {
	mod := &Module{
		Path:     path,
		Status:   Uninitialized,
		callback: f.callback,
	}
	if init {
		mod.Status = Initialized
	}
	return mod, nil
}

func (m *Module) ID() string          { return m.Path }
func (m *Module) String() string      { return m.Path }
func (m *Module) Title() string       { return m.Path }
func (m *Module) Description() string { return m.Path }
func (m *Module) FilterValue() string { return m.Path }

// Enqueuable determines whether a task belonging to the module can be moved
// onto the global queue or not.
func (m *Module) Enqueuable(queue []*task.Task) bool {
	if m.Status != Initialized {
		// Cannot enqueue tasks for a module that is not in the initialized
		// state.
		return false
	}
	if len(queue) > 0 && queue[0].Kind == InitTask {
		// A queued InitTask blocks pending task (and a queue for a module can
		// only contain *one* InitTask, so we're safe just checking the first
		// item).
		return false
	}
	return true
}

func (m *Module) updateStatus(status Status) {
	m.Status = status
	m.callback(m)
}

//Args: []string{"apply", "-auto-approve", "-input=false"},

// FindModules finds root modules that are descendents of the given path and
// returns their paths. Determining what is a root module is difficult and
// relies on a set of heuristics:
//
// (a) check if path contains a .terraform directory, else fallback to:
// (b) path has a .tf file containing a backend block
//
// (a) will only succeed if the module has already been initialized, i.e. terraform
// init has been run, whereas (b) is necessary if it has not.
func (f *factory) findModules(parent string) (modules []*Module, err error) {
	walkfn := func(path string, d fs.DirEntry, walkerr error) error {
		// skip files in directories that have already been identified as containing a
		// root module
		if slices.ContainsFunc(modules, func(m *Module) bool {
			return filepath.Dir(path) == m.Path
		}) {
			return nil
		}
		if walkerr != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".terraform" {
				mod, err := f.newModule(filepath.Dir(path), true)
				if err != nil {
					return err
				}
				modules = append(modules, mod)
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
						mod, err := f.newModule(filepath.Dir(path), false)
						if err != nil {
							return err
						}
						modules = append(modules, mod)
						return nil
					}
				}
			}
		}
		return nil
	}
	return modules, filepath.WalkDir(parent, walkfn)
}
