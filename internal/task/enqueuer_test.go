package task

import (
	"slices"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestEnqueuer(t *testing.T) {
	t.Parallel()

	f := &factory{counter: internal.Int(0)}

	mod1 := resource.New(resource.Module, resource.GlobalResource)
	ws1 := resource.New(resource.Workspace, mod1)

	mod1Task1 := f.newTask(CreateOptions{Parent: mod1})
	mod1TaskBlocking1 := f.newTask(CreateOptions{Parent: mod1, Blocking: true})

	ws1Task1 := f.newTask(CreateOptions{Parent: ws1})
	ws1Task2 := f.newTask(CreateOptions{Parent: ws1})
	ws1TaskBlocking1 := f.newTask(CreateOptions{Parent: ws1, Blocking: true})
	ws1TaskBlocking2 := f.newTask(CreateOptions{Parent: ws1, Blocking: true})
	ws1TaskBlocking3 := f.newTask(CreateOptions{Parent: ws1, Blocking: true})
	ws1TaskImmediate := f.newTask(CreateOptions{Parent: ws1, Immediate: true})
	ws1TaskDependOnTask1 := f.newTask(CreateOptions{Parent: ws1, DependsOn: []resource.ID{ws1Task1.ID}})

	ws1TaskCompleted := f.newTask(CreateOptions{Parent: ws1})
	ws1TaskCompleted.updateState(Exited)

	ws1TaskDependOnCompletedTask := f.newTask(CreateOptions{Parent: ws1, DependsOn: []resource.ID{ws1TaskCompleted.ID}})

	tests := []struct {
		name string
		// Active tasks
		active []*Task
		// Pending tasks
		pending []*Task
		// Other tasks for retrieval via their ID
		other []*Task
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
		{
			name:    "only enqueue one of three tasks which block same parent",
			active:  []*Task{},
			pending: []*Task{ws1TaskBlocking1, ws1TaskBlocking2, ws1TaskBlocking3},
			want:    []*Task{ws1TaskBlocking1},
		},
		{
			name:    "enqueue immediate task despite being blocked",
			active:  []*Task{ws1TaskBlocking1},
			pending: []*Task{ws1TaskImmediate},
			want:    []*Task{ws1TaskImmediate},
		},
		{
			name:    "don't enqueue task with a dependency on an incomplete task",
			active:  []*Task{ws1Task1},
			pending: []*Task{ws1TaskDependOnTask1},
			want:    []*Task{},
		},
		{
			name:    "enqueue task with a dependency on a completed task",
			other:   []*Task{ws1TaskCompleted},
			pending: []*Task{ws1TaskDependOnCompletedTask},
			want:    []*Task{ws1TaskDependOnCompletedTask},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := enqueuer{
				tasks: &fakeEnqueuerTaskService{
					pending: tt.pending,
					active:  tt.active,
					other:   tt.other,
				},
			}
			assert.Equal(t, tt.want, e.enqueuable())
		})
	}
}

type fakeEnqueuerTaskService struct {
	pending, active, other []*Task
}

func (f *fakeEnqueuerTaskService) List(opts ListOptions) []*Task {
	if slices.Equal(opts.Status, []Status{Queued, Running}) {
		return f.active
	}
	if slices.Equal(opts.Status, []Status{Pending}) {
		return f.pending
	}
	return nil
}

func (f *fakeEnqueuerTaskService) Get(id resource.ID) (*Task, error) {
	for _, task := range append(append(f.pending, f.active...), f.other...) {
		if id == task.ID {
			return task, nil
		}
	}
	return nil, resource.ErrNotFound
}
