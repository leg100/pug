package run

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/testutils"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Scheduler(t *testing.T) {
	// newRun() creates a directory so change into a temp dir first
	testutils.ChTempDir(t, t.TempDir())

	mod1 := module.New(internal.NewTestWorkdir(t), "a/b/c")
	ws1, err := workspace.New(mod1, "dev")
	require.NoError(t, err)
	ws2, err := workspace.New(mod1, "prod")
	require.NoError(t, err)

	ws1Run1, _ := newRun(mod1, ws1, CreateOptions{})
	ws1Run2, _ := newRun(mod1, ws1, CreateOptions{})

	ws2Run1, _ := newRun(mod1, ws2, CreateOptions{})
	ws2Run2, _ := newRun(mod1, ws2, CreateOptions{})

	tests := []struct {
		name    string
		pending []*Run
		active  []*Run
		want    []*Run
	}{
		{
			name: "schedule nothing",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := scheduler{
				runs: &fakeSchedulerLister{
					pending: tt.pending,
					active:  tt.active,
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
	pending, active []*Run
}

func (f *fakeSchedulerLister) List(opts ListOptions) []*Run {
	if len(opts.Status) > 0 && opts.Status[0] == Pending {
		return f.pending
	}
	return f.active
}
