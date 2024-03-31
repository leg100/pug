package run

import (
	"testing"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/testutils"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func Test_Scheduler(t *testing.T) {
	// newRun() creates a directory so change into a temp dir first
	testutils.ChTempDir(t)

	mod1 := module.New("a/b/c")
	ws1 := workspace.New(mod1, "dev")
	ws2 := workspace.New(mod1, "prod")

	ws1Run1, _ := newRun(mod1, ws1, CreateOptions{})
	ws1Run2, _ := newRun(mod1, ws1, CreateOptions{})
	ws1RunPlanOnly, _ := newRun(mod1, ws1, CreateOptions{PlanOnly: true})

	ws2Run1, _ := newRun(mod1, ws2, CreateOptions{})
	ws2Run2, _ := newRun(mod1, ws2, CreateOptions{})
	ws2RunPlanOnly, _ := newRun(mod1, ws2, CreateOptions{PlanOnly: true})

	tests := []struct {
		name            string
		pending         []*Run
		pendingPlanOnly []*Run
		active          []*Run
		want            []*Run
	}{
		{
			name: "schedule nothing",
		},
		{
			name:            "auto schedule pending plan-only runs",
			pendingPlanOnly: []*Run{ws1RunPlanOnly, ws2RunPlanOnly},
			active:          []*Run{},
			want:            []*Run{ws1RunPlanOnly, ws2RunPlanOnly},
		},
		{
			name:    "schedule oldest pending runs for each workspace",
			pending: []*Run{ws1Run1, ws1Run2, ws2Run1, ws2Run2},
			active:  []*Run{},
			want:    []*Run{ws1Run1, ws2Run1},
		},
		{
			name:    "dont schedule run on blocked workspace",
			pending: []*Run{ws1Run2},
			active:  []*Run{ws1Run1},
		},
		{
			name:    "schedule one unblocked pending run, and dont schedule a blocked pending run",
			pending: []*Run{ws1Run2, ws2Run1},
			active:  []*Run{ws1Run1},
			want:    []*Run{ws2Run1},
		},
		{
			name:            "schedule both pending plan-only run and a run that is not plan-only",
			pending:         []*Run{ws1Run1},
			pendingPlanOnly: []*Run{ws1RunPlanOnly},
			want:            []*Run{ws1RunPlanOnly, ws1Run1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := scheduler{
				runs: &fakeSchedulerLister{
					pending:         tt.pending,
					pendingPlanOnly: tt.pendingPlanOnly,
					active:          tt.active,
				},
			}
			got := e.schedule()
			assert.Equal(t, len(tt.want), len(got))
			for _, want := range tt.want {
				assert.Contains(t, got, want)
			}
		})
	}
}

type fakeSchedulerLister struct {
	pending, pendingPlanOnly, active []*Run
}

func (f *fakeSchedulerLister) List(opts ListOptions) []*Run {
	if opts.PlanOnly != nil {
		if *opts.PlanOnly {
			return f.pendingPlanOnly
		} else {
			return f.pending
		}
	}
	// Lazily assume lack of plan-only filter means return active runs.
	return f.active
}
