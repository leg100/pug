package task

import (
	"slices"
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestRunner_runnable(t *testing.T) {
	t.Parallel()

	mod1 := resource.New(resource.Module, resource.GlobalResource)

	t1 := &Task{Resource: resource.New(resource.Task, mod1)}
	t2 := &Task{Resource: resource.New(resource.Task, mod1)}
	t3 := &Task{Resource: resource.New(resource.Task, mod1)}
	ex1 := &Task{Resource: resource.New(resource.Task, mod1), exclusive: true}
	ex2 := &Task{Resource: resource.New(resource.Task, mod1), exclusive: true}
	immediate := &Task{Resource: resource.New(resource.Task, mod1), Immediate: true}

	tests := []struct {
		name string
		// Max runnable tasks
		max int
		// Queued tasks
		queued []*Task
		// Running tasks
		running []*Task
		// Running exclusive tasks
		exclusive []*Task
		// Want these runnable tasks
		want []*Task
	}{
		{
			name:    "all queued tasks are runnable",
			max:     3,
			queued:  []*Task{t1, t2, t3},
			running: nil,
			want:    []*Task{t1, t2, t3},
		},
		{
			name:    "only one queued task is runnable because max tasks is one",
			max:     1,
			queued:  []*Task{t1, t2, t3},
			running: nil,
			want:    []*Task{t1},
		},
		{
			name:    "no tasks are runnable because max tasks are already running",
			max:     2,
			queued:  []*Task{t3},
			running: []*Task{t1, t2},
			want:    []*Task{},
		},
		{
			name:    "only one task is runnable because there is only one available slot",
			max:     2,
			queued:  []*Task{t2, t3},
			running: []*Task{t1},
			want:    []*Task{t2},
		},
		{
			name:    "only one of two queued exclusive tasks is runnable",
			max:     3,
			queued:  []*Task{ex1, ex2},
			running: nil,
			want:    []*Task{ex1},
		},
		{
			name:      "queued exclusive task is not runnable because an exclusive task is already running",
			max:       3,
			queued:    []*Task{ex2},
			exclusive: []*Task{ex1},
			want:      []*Task{},
		},
		{
			name:    "only one exclusive task is runnable",
			max:     3,
			queued:  []*Task{ex1, ex2},
			running: nil,
			want:    []*Task{ex1},
		},
		{
			name:    "multiple non-exclusive tasks and one exclusive task is runnable",
			max:     4,
			queued:  []*Task{t1, t2, ex1, t3},
			running: nil,
			want:    []*Task{t1, t2, ex1, t3},
		},
		{
			name:    "start immediate task, despite max tasks already running",
			max:     2,
			queued:  []*Task{immediate},
			running: []*Task{t1, t2},
			want:    []*Task{immediate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &runner{
				max: tt.max,
				tasks: &fakeRunnerLister{
					queued:    tt.queued,
					running:   tt.running,
					exclusive: tt.exclusive,
				},
			}
			assert.Equal(t, tt.want, runner.runnable())
		})
	}
}

type fakeRunnerLister struct {
	queued, running, exclusive []*Task
}

func (f *fakeRunnerLister) List(opts ListOptions) []*Task {
	if opts.Exclusive {
		return f.exclusive
	}
	if slices.Equal(opts.Status, []Status{Queued}) {
		return f.queued
	}
	if slices.Equal(opts.Status, []Status{Running}) {
		return f.running
	}
	return nil
}
