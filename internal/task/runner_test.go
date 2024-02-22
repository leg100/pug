package task

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRunner_runnable(t *testing.T) {
	t1 := &Task{ID: uuid.New()}
	t2 := &Task{ID: uuid.New()}
	t3 := &Task{ID: uuid.New()}
	ex1 := &Task{ID: uuid.New(), exclusive: true}
	ex2 := &Task{ID: uuid.New(), exclusive: true}

	tests := []struct {
		name string
		// Max runnable tasks
		max int
		// Queued tasks
		queued []*Task
		// Running tasks
		running []*Task
		// Want these runnable tasks
		want []*Task
	}{
		{
			"all queued tasks are runnable",
			3,
			[]*Task{t1, t2, t3},
			nil,
			[]*Task{t1, t2, t3},
		},
		{
			"only one queued task is runnable because max tasks is one",
			1,
			[]*Task{t1, t2, t3},
			nil,
			[]*Task{t1},
		},
		{
			"no tasks are runnable because max tasks are already running",
			2,
			[]*Task{t3},
			[]*Task{t1, t2},
			nil,
		},
		{
			"only one task is runnable because there is only one available slot",
			2,
			[]*Task{t2, t3},
			[]*Task{t1},
			[]*Task{t2},
		},
		{
			"only one exclusive task is runnable",
			3,
			[]*Task{ex1, ex2},
			nil,
			[]*Task{ex1},
		},
		{
			"multiple non-exclusive tasks and one exclusive task is runnable",
			4,
			[]*Task{t1, t2, ex1, t3},
			nil,
			[]*Task{t1, t2, ex1, t3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := newRunner(tt.max, &fakeTaskLister{
				queued:  tt.queued,
				running: tt.running,
			})
			assert.Equal(t, tt.want, runner.runnable())
		})
	}
}

type fakeTaskLister struct {
	queued, running []*Task
}

func (f *fakeTaskLister) List(opts ListOptions) []*Task {
	if slices.Equal(opts.Status, []Status{Queued}) {
		return f.queued
	}
	if slices.Equal(opts.Status, []Status{Running}) {
		return f.running
	}
	return nil
}
