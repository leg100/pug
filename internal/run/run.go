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
	Stale       Status = "stale"
	Errored     Status = "errored"
	Canceled    Status = "canceled"
	Discarded   Status = "discarded"

	MaxStatusLen = len(ApplyQueued)
)

type Run struct {
	resource.Mixin

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

	// Path to pug's data directory
	dataDir string

	// Call this function after every status update
	afterUpdate func(run *Run)

	// Name of tfvars file to pass to terraform plan. An empty string means
	// there is no vars file.
	varsFilename string
}

type CreateOptions struct {
	// TargetAddrs creates a plan targeting specific resources.
	TargetAddrs []state.ResourceAddress
	// Destroy creates a plan to destroy all resources.
	Destroy bool

	afterUpdate func(run *Run)
	dataDir     string
}

func newRun(mod *module.Module, ws *workspace.Workspace, opts CreateOptions) (*Run, error) {
	run := &Run{
		Mixin:       resource.New(resource.Run, ws),
		Status:      Pending,
		AutoApply:   ws.AutoApply,
		TargetAddrs: opts.TargetAddrs,
		Destroy:     opts.Destroy,
		Created:     time.Now(),
		Updated:     time.Now(),
		afterUpdate: opts.afterUpdate,
		dataDir:     opts.dataDir,
	}

	// Create directory for run artefacts including plan file etc.
	if err := os.MkdirAll(run.ArtefactsPath(), 0o755); err != nil {
		return nil, fmt.Errorf("creating run artefacts directory: %w", err)
	}

	// Check if a tfvars file exists for the workspace. If so then use it for
	// the plan.
	varsFilename := fmt.Sprintf("%s.tfvars", ws.Name)
	tfvars := filepath.Join(mod.FullPath(), varsFilename)
	if _, err := os.Stat(tfvars); err == nil {
		run.varsFilename = varsFilename
	}

	return run, nil
}

func (r *Run) WorkspaceID() resource.ID {
	return r.Parent.GetID()
}

func (r *Run) WorkspaceName() string {
	return r.Parent.String()
}

func (r *Run) ModulePath() string {
	return r.Parent.GetParent().String()
}

// PlanPath is the path to the run's plan file, relative to the module's
// directory.
func (r *Run) PlanPath() string {
	return filepath.Join(r.ArtefactsPath(), "plan")
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
	if r.varsFilename != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", r.varsFilename))
	}
	return args
}

func (r *Run) IsFinished() bool {
	switch r.Status {
	case NoChanges, Applied, Errored, Canceled, Stale:
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

// Run's dedicated directory for artefacts created during its lifetime.
func (r *Run) ArtefactsPath() string {
	return filepath.Join(r.dataDir, r.String())
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

		// Once a run is finished remove its artefacts
		_ = os.RemoveAll(r.ArtefactsPath())
	}
}
