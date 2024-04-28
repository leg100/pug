package run

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/workspace"
)

type Status string

const (
	Pending     Status = "pending"
	Scheduled   Status = "scheduled"
	PlanQueued  Status = "plan queued"
	Planning    Status = "planning"
	Planned     Status = "planned"
	NoChanges   Status = "no changes"
	ApplyQueued Status = "apply queued"
	Applying    Status = "applying"
	Applied     Status = "applied"
	Errored     Status = "errored"
	Canceled    Status = "canceled"
	Discarded   Status = "discarded"

	MaxStatusLen = len(ApplyQueued)
)

type Run struct {
	resource.Resource

	Created time.Time
	Updated time.Time

	Status      Status
	AutoApply   bool
	TargetAddrs []state.ResourceAddress
	Destroy     bool

	PlanReport  Report
	ApplyReport Report

	// Error is non-nil when the run status is Errored
	Error error

	// Run's dedicated directory for artefacts created during its lifetime. The
	// path is relative to its module directory.
	artefactsPath string

	// Call this function after every status update
	afterUpdate func(run *Run)
}

type CreateOptions struct {
	// TargetAddrs creates a plan targeting specific resources.
	TargetAddrs []state.ResourceAddress
	// Destroy creates a plan to destroy all resources.
	Destroy bool

	afterUpdate func(run *Run)
}

func newRun(mod *module.Module, ws *workspace.Workspace, opts CreateOptions) (*Run, error) {
	run := &Run{
		Resource:    resource.New(resource.Run, ws.Resource),
		Status:      Pending,
		AutoApply:   ws.AutoApply,
		TargetAddrs: opts.TargetAddrs,
		Destroy:     opts.Destroy,
		Created:     time.Now(),
		Updated:     time.Now(),
		afterUpdate: opts.afterUpdate,
	}

	// Create directory for artefacts including plan file etc.
	run.artefactsPath = filepath.Join(ws.PugDirectory(), run.String())
	if err := os.MkdirAll(filepath.Join(mod.FullPath(), run.artefactsPath), 0o755); err != nil {
		return nil, err
	}

	return run, nil
}

func (r *Run) WorkspaceID() resource.ID {
	return r.Parent.ID
}

// PlanPath is the path to the run's plan file, relative to the module's
// directory.
func (r *Run) PlanPath() string {
	return filepath.Join(r.artefactsPath, "plan.out")
}

// PlanArgs produces the arguments for terraform plan
func (r *Run) PlanArgs() []string {
	args := []string{"-input=false", "-out", r.PlanPath()}
	for _, addr := range r.TargetAddrs {
		args = append(args, fmt.Sprintf("-target=%s", addr))
	}
	if r.Destroy {
		args = append(args, "-destroy")
	}
	return args
}

func (r *Run) IsFinished() bool {
	switch r.Status {
	case NoChanges, Applied, Errored, Canceled:
		return true
	default:
		return false
	}
}

func (r *Run) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", r.String()),
		slog.String("status", string(r.Status)),
	)
}

func (r *Run) setErrored(err error) {
	r.Error = err
	r.updateStatus(Errored)
}

func (r *Run) updateStatus(status Status) {
	r.Status = status
	r.Updated = time.Now()
	if r.afterUpdate != nil {
		r.afterUpdate(r)
	}
	if r.IsFinished() {
		slog.Info("completed run", "status", r.Status, "run", r)
	}
}
