package run

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/leg100/pug/internal/pubsub"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
)

type Status string

const (
	Pending     Status = "pending"
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
	resource.Common

	Created time.Time
	Updated time.Time

	Status Status

	PlanReport    *Report
	ApplyReport   *Report
	Changes       bool
	ArtefactsPath string
	Destroy       bool

	applyOnly          bool
	defaultArgs        []string
	createPlanFileArgs []string

	// Call this function after every status update
	afterUpdate func(run *Run)
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
	modules    moduleGetter
	workspaces workspaceGetter
	broker     *pubsub.Broker[*Run]
	terragrunt bool
}

func (f *factory) newRun(workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	ws, err := f.workspaces.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	mod, err := f.modules.Get(ws.ModuleID())
	if err != nil {
		return nil, fmt.Errorf("workspace module: %w", err)
	}

	var args []string
	if opts.Destroy {
		args = append(args, "-destroy")
	}
	for _, addr := range opts.TargetAddrs {
		args = append(args, fmt.Sprintf("-target=%s", addr))
	}

	// Check if a tfvars file exists for the workspace. If so then use it for
	// the run.
	varsFilename := fmt.Sprintf("%s.tfvars", ws.Name)
	tfvars := filepath.Join(mod.FullPath(), varsFilename)
	if _, err := os.Stat(tfvars); err == nil {
		args = append(args, fmt.Sprintf("-var-file=%s", varsFilename))
	}

	run := &Run{
		Common:             resource.New(resource.Run, ws),
		Status:             Pending,
		Destroy:            opts.Destroy,
		applyOnly:          opts.applyOnly,
		defaultArgs:        []string{"-input=false"},
		createPlanFileArgs: args,
		afterUpdate: func(run *Run) {
			f.broker.Publish(resource.UpdatedEvent, run)
		},
	}

	if !opts.applyOnly {
		run.ArtefactsPath = filepath.Join(f.dataDir, fmt.Sprintf("%d", run.Serial))
		if err := os.MkdirAll(run.ArtefactsPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating run artefacts directory: %w", err)
		}
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

func (r *Run) updateStatus(status Status) {
	r.Status = status
	r.Updated = time.Now()
	r.afterUpdate(r)
}

func (r *Run) planPath() string {
	return filepath.Join(r.ArtefactsPath, "plan")
}

func (r *Run) planArgs() []string {
	args := append(r.defaultArgs, r.createPlanFileArgs...)
	if r.applyOnly {
		return args
	}
	return append(args, "-out", r.planPath())
}

func (r *Run) finishPlan(reader io.Reader) (err error) {
	defer func() {
		if err != nil {
			r.updateStatus(Errored)
		}
	}()

	out, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if err := r.parsePlanOutput(out); err != nil {
		return err
	}
	if !r.Changes {
		r.updateStatus(NoChanges)
		return nil
	}
	r.updateStatus(Planned)
	return nil
}

func (r *Run) parsePlanOutput(out []byte) error {
	changes, report, err := parsePlanReport(string(out))
	if err != nil {
		return err
	}
	r.PlanReport = &report
	r.Changes = changes
	return nil
}

func (r *Run) applyArgs() []string {
	if r.applyOnly {
		args := append(r.defaultArgs, r.createPlanFileArgs...)
		return append(args, "-auto-approve")
	}
	return append(r.defaultArgs, r.planPath())
}

func (r *Run) finishApply(reader io.Reader) (err error) {
	defer func() {
		if err != nil {
			r.updateStatus(Errored)
		}
	}()

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
	r.updateStatus(Applied)

	return nil
}
