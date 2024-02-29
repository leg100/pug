package task

import (
	"slices"
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestEnqueuer(t *testing.T) {
	mod1 := resource.New(nil)
	ws1 := resource.New(&mod1)

	mod1Task1 := &Task{Resource: resource.New(&mod1)}
	mod1TaskBlocking1 := &Task{Resource: resource.New(&mod1), Blocking: true}

	ws1Task1 := &Task{Resource: resource.New(&ws1)}
	ws1Task2 := &Task{Resource: resource.New(&ws1)}
	ws1TaskBlocking1 := &Task{Resource: resource.New(&ws1), Blocking: true}

	tests := []struct {
		name string
		// Active tasks
		active []*Task
		// Pending tasks
		pending []*Task
		// Want these tasks enqueued
		want []*Task
	}{
		{
			name:    "enqueue task for parent resource with no active tasks",
			active:  []*Task{},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "enqueue task for parent resource with non-blocking active task",
			active:  []*Task{ws1Task2},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "enqueue task for parent resource with non-blocking active grand-parent task",
			active:  []*Task{mod1Task1},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "don't enqueue tasks for blocked parent resource",
			active:  []*Task{ws1TaskBlocking1},
			pending: []*Task{ws1Task1},
			want:    []*Task{},
		},
		{
			name:    "don't enqueue tasks for blocked grand-parent resource",
			active:  []*Task{mod1TaskBlocking1},
			pending: []*Task{ws1Task1},
			want:    []*Task{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := enqueuer{
				tasks: &fakeEnqueuerLister{
					pending: tt.pending,
					active:  tt.active,
				},
			}
			assert.Equal(t, tt.want, e.enqueue())
		})
	}
}

type fakeEnqueuerLister struct {
	pending, active []*Task
}

func (f *fakeEnqueuerLister) List(opts ListOptions) []*Task {
	if slices.Equal(opts.Status, []Status{Queued, Running}) {
		return f.active
	}
	if slices.Equal(opts.Status, []Status{Pending}) {
		return f.pending
	}
	return nil
}
