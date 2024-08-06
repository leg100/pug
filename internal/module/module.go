package module

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
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

	// The module's backend type
	Backend string
}

type Options struct {
	// Path is the module path relative to the working directory.
	Path string
	// Backend is the type of terraform backend
	Backend string
}

// New constructs a module. Workdir is the pug working directory, and path is
// the module path relative to the working directory.
func New(workdir internal.Workdir, opts Options) *Module {
	return &Module{
		Common:  resource.New(resource.Module, resource.GlobalResource),
		Workdir: workdir,
		Path:    opts.Path,
		Backend: opts.Backend,
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

// find finds root modules that are descendents of the workdir and
// returns options for creating equivalent pug modules.
//
// A root module is deemed to be a directory that contains a .tf file that
// contains a backend or cloud block, or in the case of terragrunt, a
// terragrunt.hcl file.
//
// find returns two channels: the first streams discovered modules (in the form
// of Options structs for creating the module in pug); the second streams any
// errors encountered.
//
// find closes both channels when finished.
//
// TODO: take a done channel parameter
func find(workdir internal.Workdir) (<-chan Options, <-chan error) {
	modules := make(chan Options)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup
		errc <- filepath.WalkDir(workdir.String(), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				errc <- err
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case ".terraform", ".terragrunt-cache":
					return filepath.SkipDir
				}
				return nil
			}

			var isTerragrunt bool
			switch {
			case d.Name() == "terragrunt.hcl":
				isTerragrunt = true
				fallthrough
			case filepath.Ext(path) == ".tf":
				wg.Add(1)
				go func() {
					defer wg.Done()
					backend, found, err := detectBackend(path)
					if err != nil {
						errc <- err
						return
					}
					if !isTerragrunt && !found {
						// Not a terragrunt module, nor a vanilla terraform module with a
						// backend config, so skip.
						return
					}
					//if isTerragrunt && backend == "" {
					// Unless terragrunt.hcl directly contains a `remote_state`
					// block then Pug doesn't have a way of determining the backend
					// type (not unless it evaluates terragrunt's language and
					// follows `find_in_parent` etc. to locate the effective
					// remote_state, which is perhaps a future exercise...).
					//}
					// Strip workdir from module path
					stripped, err := filepath.Rel(workdir.String(), filepath.Dir(path))
					if err != nil {
						errc <- err
						return
					}
					modules <- Options{
						Path:    stripped,
						Backend: backend,
					}
				}()
				// skip walking remainder of parent directory
				return fs.SkipDir
			}

			return nil
		})
		go func() {
			wg.Wait()
			close(modules)
			close(errc)
		}()
	}()
	return modules, errc
}

type terragrunt struct {
	RemoteState *terragruntRemoteState `hcl:"remote_state,block"`
	Remain      hcl.Body               `hcl:",remain"`
}

type terragruntRemoteState struct {
	Backend string   `hcl:"backend,attr"`
	Remain  hcl.Body `hcl:",remain"`
}

type terraform struct {
	Terraform *terraformBlock `hcl:"terraform,block"`
	Remain    hcl.Body        `hcl:",remain"`
}

type terraformBlock struct {
	Backend *terraformBackend `hcl:"backend,block"`
	Cloud   *terraformCloud   `hcl:"cloud,block"`
	Remain  hcl.Body          `hcl:",remain"`
}

type terraformBackend struct {
	Type   string   `hcl:"type,label"`
	Remain hcl.Body `hcl:",remain"`
}

type terraformCloud struct {
	Remain hcl.Body `hcl:",remain"`
}

// detectBackend parses the HCL file at the given path and detects whether it
// found a backend configuration, together with the type of backend it found.
func detectBackend(path string) (string, bool, error) {
	f, err := hclparse.NewParser().ParseHCLFile(path)
	if err != nil {
		return "", false, err
	}
	// Detect terraform backend
	var terraform terraform
	if diags := gohcl.DecodeBody(f.Body, nil, &terraform); diags != nil {
		return "", false, diags
	}
	if terraform.Terraform != nil {
		if terraform.Terraform.Backend != nil {
			return terraform.Terraform.Backend.Type, true, nil
		}
		if terraform.Terraform.Cloud != nil {
			return "cloud", true, nil
		}
	}
	// Detect terragrunt remote state configuration
	var remoteStateBlock terragrunt
	if diags := gohcl.DecodeBody(f.Body, nil, &remoteStateBlock); diags != nil {
		return "", false, diags
	}
	if remoteStateBlock.RemoteState != nil {
		return remoteStateBlock.RemoteState.Backend, true, nil
	}
	return "", false, nil
}
