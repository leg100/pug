package plan

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	workspaces workspaceGetter
}

func cmdArgs(opts CreateOptions, ws *workspace.Workspace) []string {
	args := []string{"-input=false"}
	for _, addr := range opts.TargetAddrs {
		args = append(args, fmt.Sprintf("-target=%s", addr))
	}
	if fname, ok := ws.VarsFile(); ok {
		args = append(args, fmt.Sprintf("-var-file=%s", fname))
	}
	if opts.Destroy {
		args = append(args, "-destroy")
	}
	return args
}

func (f *factory) newPlan(workspaceID resource.ID, opts CreateOptions) (*plan, error) {
	ws, err := f.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	plan := &plan{
		Common:  resource.New(resource.Plan, ws),
		Destroy: opts.Destroy,
	}
	plan.ArtefactsPath = filepath.Join(f.dataDir, fmt.Sprintf("%d", plan.Serial))
	if err := os.MkdirAll(plan.ArtefactsPath, 0o755); err != nil {
		return nil, fmt.Errorf("creating run artefacts directory: %w", err)
	}
	return plan, nil
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

func (r *plan) planTaskSpec() task.Spec {
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
	if r.Destroy {
		spec.Description += " (destroy)"
	}
	return spec
}
