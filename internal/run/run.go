package run

import (
	"os"
	"path/filepath"
	"time"

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

type Run struct {
	resource.Resource

	Status    Status
	Created   time.Time
	AutoApply bool

	PlanReport  report
	ApplyReport report

	// Error is non-nil when the run status is Errored
	Error error

	// Call this function after every status update
	afterUpdate func(run *Run)
}

type CreateOptions struct {
	AutoApply bool
}

func newRun(ws *workspace.Workspace, opts CreateOptions) (*Run, error) {
	run := &Run{
		Resource:  resource.New(),
		Status:    Pending,
		AutoApply: opts.AutoApply,
		Created:   time.Now(),
		Workspace: ws.Resource,
	}
	if err := os.MkdirAll(run.PugDirectory(), 0o755); err != nil {
		return nil, err
	}
	return run, nil
}

func (r *Run) PugDirectory() string {
	return filepath.Join(r.Workspace.PugDirectory(), r.String())
}

func (r *Run) PlanPath() string {
	return filepath.Join(r.PugDirectory(), "plan.out")
}

func (r *Run) setErrored(err error) {
	r.Error = err
	r.updateStatus(Errored)
}

func (r *Run) updateStatus(status Status) {
	r.Status = status
	if r.afterUpdate != nil {
		r.afterUpdate(r)
	}
}

func (r *Run) IsFinished() bool {
	switch r.Status {
	case PlannedAndFinished, Applied, Errored, Canceled:
		return true
	default:
		return false
	}
}
