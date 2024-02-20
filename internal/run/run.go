package run

import (
	"fmt"
	"time"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

type Status string

const (
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
	resource.ID

	Status      Status
	WorkspaceID workspace.ID
	Created     time.Time
	AutoApply   bool
}

func newRun(ws *workspace.Workspace, mod *module.Module, opts CreateOptions) (*Run, error) {
	if mod.Status != module.Initialized {
		return nil, fmt.Errorf("module must be initalized")
	}
	return &Run{
		ID:        resource.NewID(),
		Status:    PlanQueued,
		AutoApply: opts.AutoApply,
		Created:   time.Now(),
	}, nil
}

type CreateOptions struct {
	AutoApply bool
}
