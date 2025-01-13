package plan

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/machine"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
)

type plan struct {
	resource.ID

	ModuleID                 resource.ID
	WorkspaceID              resource.ID
	ModulePath               string
	HasChanges               bool
	ArtefactsPath            string
	Destroy                  bool
	TargetAddrs              []state.ResourceAddress
	DisableMachineReadableUI bool

	targetArgs         []string
	terragrunt         bool
	planFile           bool
	varsFileArg        *string
	envs               []string
	moduleDependencies []resource.ID

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
	modules    moduleGetter
	workspaces workspaceGetter
	broker     *pubsub.Broker[*plan]
	terragrunt bool
}

func (f *factory) newPlan(workspaceID resource.ID, opts CreateOptions) (*plan, error) {
	ws, err := f.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	mod, err := f.modules.Get(ws.ModuleID)
	if err != nil {
		return nil, fmt.Errorf("retrieving module: %w", err)
	}
	plan := &plan{
		ID:                 resource.NewID(resource.Plan),
		ModuleID:           mod.ID,
		WorkspaceID:        ws.ID,
		ModulePath:         mod.Path,
		Destroy:            opts.Destroy,
		TargetAddrs:        opts.TargetAddrs,
		planFile:           opts.planFile,
		terragrunt:         f.terragrunt,
		envs:               []string{ws.TerraformEnv()},
		moduleDependencies: mod.Dependencies(),
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

func (r *plan) planPath() string {
	return filepath.Join(r.ArtefactsPath, "plan")
}

func (r *plan) args() []string {
	args := append([]string{"-input"}, r.targetArgs...)
	if !r.DisableMachineReadableUI {
		args = append(args, "-json")
	}
	return args
}

func (r *plan) planTaskSpec() task.Spec {
	// TODO: assert planFile is true first
	spec := task.Spec{
		Identifier:  PlanTask,
		ModuleID:    &r.ModuleID,
		WorkspaceID: &r.WorkspaceID,
		Path:        r.ModulePath,
		Env:         r.envs,
		Execution: task.Execution{
			TerraformCommand: []string{"plan"},
			Args:             append(r.args(), "-out", r.planPath()),
		},
		// TODO: explain why plan is blocking (?)
		Blocking:    true,
		Description: "plan",
		AfterCreate: func(t *task.Task) {
			r.taskID = &t.ID
		},
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			reader := t.NewReader(false)
			var (
				changes bool
				report  Report
			)
			if r.DisableMachineReadableUI {
				out, err := io.ReadAll(reader)
				if err != nil {
					return nil, err
				}
				changes, report, err = parsePlanReport(string(out))
			} else {
				messages, err := machine.Unmarshal(reader)
				if err != nil {
					return nil, fmt.Errorf("unmarshaling machine readable UI output: %w", err)
				}
				var found bool
				for _, msg := range messages {
					if summary, ok := msg.(*machine.ChangeSummaryMsg); ok {
						report = Report{
							Additions:    summary.Changes.Add,
							Changes:      summary.Changes.Change,
							Destructions: summary.Changes.Remove,
						}
						// TODO: parse outputs too, and take that into account
						// as to whether there are any changes.
						changes = report.Additions != 0 || report.Changes != 0 || report.Destructions != 0
						found = true
					}
				}
				if !found {
					return nil, errors.New("no change summary found in machine readable UI output")
				}
			}
			r.HasChanges = changes
			return report, nil
		},
	}
	if r.varsFileArg != nil {
		spec.Execution.Args = append(spec.Execution.Args, *r.varsFileArg)
	}
	if r.Destroy {
		spec.Execution.Args = append(spec.Execution.Args, "-destroy")
		spec.Description += " (destroy)"
	}
	return spec
}

const (
	PlanTask  task.Identifier = "plan"
	ApplyTask task.Identifier = "apply"
)

func (r *plan) applyTaskSpec() (task.Spec, error) {
	if r.planFile && !r.HasChanges {
		return task.Spec{}, errors.New("plan does not have any changes to apply")
	}
	spec := task.Spec{
		Identifier:  ApplyTask,
		ModuleID:    &r.ModuleID,
		WorkspaceID: &r.WorkspaceID,
		Path:        r.ModulePath,
		Execution: task.Execution{
			TerraformCommand: []string{"apply"},
			Args:             r.args(),
		},
		Env:         r.envs,
		Blocking:    true,
		Description: "apply",
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
	// If terragrunt is in use then respect module dependencies.
	if r.terragrunt {
		spec.Dependencies = &task.Dependencies{
			ModuleIDs: r.moduleDependencies,
			// Module dependencies are reversed for a destroy.
			InverseDependencyOrder: r.Destroy,
		}
	}
	if r.planFile {
		spec.Execution.Args = append(spec.Execution.Args, r.planPath())
	} else {
		if r.varsFileArg != nil {
			spec.Execution.Args = append(spec.Execution.Args, *r.varsFileArg)
		}
		spec.Execution.Args = append(spec.Execution.Args, "-auto-approve")
	}
	if r.Destroy {
		if !r.planFile {
			spec.Execution.Args = append(spec.Execution.Args, "-destroy")
		}
		spec.Description += " (destroy)"
	}
	return spec, nil
}
