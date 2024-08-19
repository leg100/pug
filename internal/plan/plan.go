package plan

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type plan struct {
	resource.Common

	HasChanges    bool
	ArtefactsPath string
	Destroy       bool
	TargetAddrs   []state.ResourceAddress

	targetArgs  []string
	terragrunt  bool
	planFile    bool
	varsFileArg *string

	// taskID is the ID of the plan task, and is only set once the task is
	// created.
	taskID *resource.ID
}

type CreateOptions struct {
	// TargetAddrs creates a plan targeting specific resources.
	TargetAddrs []state.ResourceAddress
	// Destroy creates a plan to destroy all resources.
	Destroy bool
	// planFile is true if a plan file is first created with `terraform plan
	// -out plan.file`.
	planFile bool
}

type factory struct {
	dataDir    string
	workdir    internal.Workdir
	workspaces workspaceGetter
	broker     *pubsub.Broker[*plan]
	terragrunt bool
}

func (f *factory) newPlan(workspaceID resource.ID, opts CreateOptions) (*plan, error) {
	ws, err := f.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	plan := &plan{
		Common:      resource.New(resource.Plan, ws),
		Destroy:     opts.Destroy,
		TargetAddrs: opts.TargetAddrs,
		planFile:    opts.planFile,
		terragrunt:  f.terragrunt,
	}
	if opts.planFile {
		plan.ArtefactsPath = filepath.Join(f.dataDir, fmt.Sprintf("%d", plan.Serial))
		if err := os.MkdirAll(plan.ArtefactsPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating run artefacts directory: %w", err)
		}
	}
	for _, addr := range plan.TargetAddrs {
		plan.targetArgs = append(plan.targetArgs, fmt.Sprintf("-target=%s", addr))
	}
	if fname, ok := ws.VarsFile(f.workdir); ok {
		flag := fmt.Sprintf("-var-file=%s", fname)
		plan.varsFileArg = &flag
	}
	return plan, nil
}

func (r *plan) WorkspaceID() resource.ID {
	return r.Parent.GetID()
}

func (r *plan) WorkspaceName() string {
	return r.Parent.String()
}

func (r *plan) ModulePath() string {
	return r.Parent.GetParent().String()
}

func (r *plan) planPath() string {
	return filepath.Join(r.ArtefactsPath, "plan")
}

func (r *plan) args() []string {
	return append([]string{"-input"}, r.targetArgs...)
}

func (r *plan) planTaskSpec(logger logging.Interface) task.Spec {
	// TODO: assert planFile is true first
	spec := task.Spec{
		Parent:  r.Workspace(),
		Path:    r.ModulePath(),
		Env:     []string{workspace.TerraformEnv(r.WorkspaceName())},
		Command: []string{"plan"},
		Args:    append(r.args(), "-out", r.planPath()),
		// TODO: explain why plan is blocking (?)
		Blocking:    true,
		Description: "plan",
		AfterCreate: func(t *task.Task) {
			r.taskID = &t.ID
		},
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			out, err := io.ReadAll(t.NewReader(false))
			if err != nil {
				return nil, err
			}
			changes, report, err := parsePlanReport(string(out))
			if err != nil {
				return nil, err
			}
			r.HasChanges = changes
			return report, nil
		},
	}
	if r.varsFileArg != nil {
		spec.Args = append(spec.Args, *r.varsFileArg)
	}
	if r.Destroy {
		spec.Args = append(spec.Args, "-destroy")
		spec.Description += " (destroy)"
	}
	return spec
}

func (r *plan) applyTaskSpec(logger logging.Interface) (task.Spec, error) {
	if r.planFile && !r.HasChanges {
		return task.Spec{}, errors.New("plan does not have any changes to apply")
	}
	spec := task.Spec{
		Parent:      r.Workspace(),
		Path:        r.ModulePath(),
		Command:     []string{"apply"},
		Args:        r.args(),
		Env:         []string{workspace.TerraformEnv(r.WorkspaceName())},
		Blocking:    true,
		Description: "apply",
		// If terragrunt is in use then respect module dependencies.
		RespectModuleDependencies: r.terragrunt,
		// Module dependencies are reversed for a destroy.
		InverseDependencyOrder: r.Destroy,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			out, err := io.ReadAll(t.NewReader(false))
			if err != nil {
				return nil, err
			}
			if r.planFile {
				// Plan file can now be safely removed
				_ = os.RemoveAll(r.ArtefactsPath)
			}
			report, err := parseApplyReport(string(out))
			if err != nil {
				return nil, err
			}
			return report, nil
		},
	}
	if r.planFile {
		spec.Args = append(spec.Args, r.planPath())
	} else {
		if r.varsFileArg != nil {
			spec.Args = append(spec.Args, *r.varsFileArg)
		}
		spec.Args = append(spec.Args, "-auto-approve")
	}
	if r.Destroy {
		if !r.planFile {
			spec.Args = append(spec.Args, "-destroy")
		}
		spec.Description += " (destroy)"
	}
	return spec, nil
}

func IsApplyTask(t *task.Task) bool {
	return len(t.Command) > 0 && t.Command[0] == "apply"
}
