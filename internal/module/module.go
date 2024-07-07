package module

import (
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
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

// findModules finds root modules that are descendents of the workdir and
// returns options for creating equivalent pug modules.
//
// A root module is deemed to be a directory that contains a .tf file that
// contains a backend or cloud block, or in the case of terragrunt, a
// terragrunt.hcl file containing a remote_state block.
func findModules(logger logging.Interface, workdir internal.Workdir) (modules []Options, err error) {
	walkfn := func(path string, d fs.DirEntry, walkerr error) error {
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
		if filepath.Ext(path) == ".tf" || d.Name() == "terragrunt.hcl" {
			backend, found, err := detectBackend(path)
			if err != nil {
				logger.Error("reloading modules: parsing hcl", "path", path, "error", err)
				return nil
			}
			if !found {
				return nil
			}
			// Strip workdir from module path
			stripped, err := filepath.Rel(workdir.String(), filepath.Dir(path))
			if err != nil {
				return err
			}
			modules = append(modules, Options{
				Path:    stripped,
				Backend: backend,
			})
			// skip walking remainder of parent directory
			return fs.SkipDir
		}
		return nil
	}
	if err := filepath.WalkDir(workdir.String(), walkfn); err != nil {
		return nil, err
	}
	return
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
