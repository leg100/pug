package run

import (
	"os"
	"path/filepath"
	"time"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

type Status string

const (
	Pending            Status = "pending"
	PlanQueued         Status = "plan_queued"
	Planning           Status = "planning"
	Planned            Status = "planned"
	PlannedAndFinished Status = "planned_and_finished"
	ApplyQueued        Status = "apply_queued"
	Applying           Status = "applying"
	Applied            Status = "applied"
	Errored            Status = "errored"
	Canceled           Status = "canceled"
)

func PugDirectory(module *module.Module, ws *workspace.Workspace, run *Run) string {
	return filepath.Join(workspace.PugDirectory(module.Path, ws.String()), run.String())
}

func PlanPath(module *module.Module, ws *workspace.Workspace, run *Run) string {
	return filepath.Join(PugDirectory(module, ws, run), "plan.out")
}

type Run struct {
	resource.Resource

	Status    Status
	Created   time.Time
	AutoApply bool
	PlanOnly  bool

	PlanReport  report
	ApplyReport report

	PlanTask  *resource.Resource
	ApplyTask *resource.Resource

	// Error is non-nil when the run status is Errored
	Error error

	// Call this function after every status update
	afterUpdate func(run *Run)
}

type CreateOptions struct {
	AutoApply bool
	PlanOnly  bool

	afterUpdate func(run *Run)
}

func newRun(mod *module.Module, ws *workspace.Workspace, opts CreateOptions) (*Run, error) {
	run := &Run{
		Resource:    resource.New(resource.Run, "", &ws.Resource),
		Status:      Pending,
		AutoApply:   opts.AutoApply,
		PlanOnly:    opts.PlanOnly,
		Created:     time.Now(),
		afterUpdate: opts.afterUpdate,
	}
	if err := os.MkdirAll(PugDirectory(mod, ws, run), 0o755); err != nil {
		return nil, err
	}
	return run, nil
}

func (r *Run) Workspace() resource.Resource {
	return *r.Parent
}

func (r *Run) IsFinished() bool {
	switch r.Status {
	case PlannedAndFinished, Applied, Errored, Canceled:
		return true
	default:
		return false
	}
}

func (r *Run) CurrentTask() resource.Resource {
	if r.ApplyTask != nil {
		return *r.ApplyTask
	}
	return *r.PlanTask
}

func (r *Run) setErrored(err error) {
	r.Error = err
	r.updateStatus(Errored)
}

func (r *Run) addPlan(pfile planFile) (apply bool) {
	r.PlanReport = pfile.resourceChanges()
	if !r.PlanReport.HasChanges() {
		r.updateStatus(PlannedAndFinished)
		return false
	}
	r.updateStatus(Planned)
	return r.AutoApply
}

func (r *Run) updateStatus(status Status) {
	r.Status = status
	if r.afterUpdate != nil {
		r.afterUpdate(r)
	}
}
