package run

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type Plan struct {
	resource.Common

	// taskID is the ID of the plan task, and is only set once the task is
	// created.
	taskID *resource.ID

	PlanReport    *Report
	ApplyReport   *Report
	HasChanges    bool
	ArtefactsPath string
	Destroy       bool
	TargetAddrs   []state.ResourceAddress

	applyOnly    bool
	defaultArgs  []string
	planFileArgs []string
	terragrunt   bool
}

type CreateOptions struct {
	// TargetAddrs creates a plan targeting specific resources.
	TargetAddrs []state.ResourceAddress
	// Destroy creates a plan to destroy all resources.
	Destroy bool
	// applyOnly skips the plan task and goes directly to creating an apply task
	// without a plan file.
	applyOnly bool
}

type factory struct {
	dataDir    string
	workspaces workspaceGetter
	broker     *pubsub.Broker[*Plan]
	terragrunt bool
}

func (f *factory) newPlan(workspaceID resource.ID, opts CreateOptions) (*Plan, error) {
	ws, err := f.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	plan := &Plan{
		Common:      resource.New(resource.Plan, ws),
		Destroy:     opts.Destroy,
		TargetAddrs: opts.TargetAddrs,
		applyOnly:   opts.applyOnly,
		defaultArgs: []string{"-input=false"},
	}
	if !opts.applyOnly {
		plan.ArtefactsPath = filepath.Join(f.dataDir, fmt.Sprintf("%d", plan.Serial))
		if err := os.MkdirAll(plan.ArtefactsPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating run artefacts directory: %w", err)
		}
	}
	if plan.Destroy {
		plan.planFileArgs = append(plan.planFileArgs, "-destroy")
	}
	for _, addr := range plan.TargetAddrs {
		plan.planFileArgs = append(plan.planFileArgs, fmt.Sprintf("-target=%s", addr))
	}
	if fname, ok := ws.VarsFile(); ok {
		plan.planFileArgs = append(plan.planFileArgs, fmt.Sprintf("-var-file=%s", fname))
	}
	return plan, nil
}

func (r *Plan) WorkspaceID() resource.ID {
	return r.Parent.GetID()
}

func (r *Plan) WorkspaceName() string {
	return r.Parent.String()
}

func (r *Plan) ModulePath() string {
	return r.Parent.GetParent().String()
}

func (r *Plan) planPath() string {
	return filepath.Join(r.ArtefactsPath, "plan")
}

func (r *Plan) planTaskSpec(logger logging.Interface) task.Spec {
	spec := task.Spec{
		Parent:  r.Workspace(),
		Path:    r.ModulePath(),
		Env:     []string{workspace.TerraformEnv(r.WorkspaceName())},
		Command: []string{"plan"},
		Args:    append(r.defaultArgs, r.planFileArgs...),
		// TODO: explain why plan is blocking (?)
		Blocking:    true,
		Description: "plan",
		AfterCreate: func(t *task.Task) {
			r.taskID = &t.ID
		},
		AfterExited: func(t *task.Task) {
			if err := r.finishPlan(t.NewReader(false)); err != nil {
				logger.Error("finishing plan", "error", err, "run", r)
			}
		},
	}
	if !r.applyOnly {
		spec.Args = append(spec.Args, "-out", r.planPath())
	}
	if r.Destroy {
		spec.Description += " (destroy)"
	}
	return spec
}

func (r *Plan) finishPlan(reader io.Reader) (err error) {
	out, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if err := r.parsePlanOutput(out); err != nil {
		return err
	}
	return nil
}

func (r *Plan) parsePlanOutput(out []byte) error {
	changes, report, err := parsePlanReport(string(out))
	if err != nil {
		return err
	}
	r.PlanReport = &report
	r.HasChanges = changes
	return nil
}

func (r *Plan) applyTaskSpec(logger logging.Interface) (task.Spec, error) {
	if !r.applyOnly && r.HasChanges {
		return task.Spec{}, errors.New("plan does not have any changes to apply")
	}
	spec := task.Spec{
		Parent:      r.Workspace(),
		Path:        r.ModulePath(),
		Command:     []string{"apply"},
		Args:        r.defaultArgs,
		Env:         []string{workspace.TerraformEnv(r.WorkspaceName())},
		Blocking:    true,
		Description: "apply",
		// If terragrunt is in use then respect module dependencies.
		RespectModuleDependencies: r.terragrunt,
		// Module dependencies are reversed for a destroy.
		InverseDependencyOrder: r.Destroy,
		AfterExited: func(t *task.Task) {
			if err := r.finishApply(t.NewReader(false)); err != nil {
				logger.Error("finishing apply", "error", err, "workspace", r.Workspace())
				return
			}
		},
	}
	if r.applyOnly {
		spec.Args = append(spec.Args, r.planFileArgs...)
		spec.Args = append(spec.Args, "-auto-approve")
	} else {
		spec.Args = append(spec.Args, r.planPath())
	}
	if r.Destroy {
		spec.Description += " (destroy)"
	}
	return spec, nil
}

func (r *Plan) finishApply(reader io.Reader) error {
	out, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if r.applyOnly {
		if err := r.parsePlanOutput(out); err != nil {
			return err
		}
	} else {
		// Plan file can be safely removed
		_ = os.RemoveAll(r.ArtefactsPath)
	}
	report, err := parseApplyReport(string(out))
	if err != nil {
		return err
	}
	r.ApplyReport = &report
	return nil
}
