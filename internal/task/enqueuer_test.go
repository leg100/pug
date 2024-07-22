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
	run1 := resource.New(resource.Run, ws1)
	run2 := resource.New(resource.Run, ws1)

	mod1Task1 := f.newTask(Spec{Parent: mod1})
	mod1TaskBlocking1 := f.newTask(Spec{Parent: mod1, Blocking: true})

	ws1Task1 := f.newTask(Spec{Parent: ws1})
	ws1Task2 := f.newTask(Spec{Parent: ws1})
	ws1TaskBlocking1 := f.newTask(Spec{Parent: ws1, Blocking: true})
	ws1TaskBlocking2 := f.newTask(Spec{Parent: ws1, Blocking: true})
	ws1TaskBlocking3 := f.newTask(Spec{Parent: ws1, Blocking: true})
	ws1TaskImmediate := f.newTask(Spec{Parent: ws1, Immediate: true})
	ws1TaskDependOnTask1 := f.newTask(Spec{Parent: ws1, DependsOn: []resource.ID{ws1Task1.ID}})

	ws1TaskCompleted := f.newTask(Spec{Parent: ws1})
	ws1TaskCompleted.updateState(Exited)

	ws1TaskDependOnCompletedTask := f.newTask(Spec{Parent: ws1, DependsOn: []resource.ID{ws1TaskCompleted.ID}})

	run1TaskBlocking1 := f.newTask(Spec{Parent: run1, Blocking: true})
	run2Task1 := f.newTask(Spec{Parent: run2})

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
			name:    "enqueue pending workspace task",
			active:  []*Task{},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "enqueue pending workspace task alongside non-blocking active workspace task",
			active:  []*Task{ws1Task2},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "enqueue pending workspace task alongside non-blocking active module task",
			active:  []*Task{mod1Task1},
			pending: []*Task{ws1Task1},
			want:    []*Task{ws1Task1},
		},
		{
			name:    "don't enqueue workspace task when there is an active blocking workspace task sharing same workspace",
			active:  []*Task{ws1TaskBlocking1},
			pending: []*Task{ws1Task1},
			want:    []*Task{},
		},
		{
			name:    "don't enqueue workspace task when there is an active blocking module task sharing same module",
			active:  []*Task{mod1TaskBlocking1},
			pending: []*Task{ws1Task1},
			want:    []*Task{},
		},
		{
			name:    "don't enqueue run task when there is an active blocking run task sharing same workspace",
			active:  []*Task{run1TaskBlocking1},
			pending: []*Task{run2Task1},
			want:    []*Task{},
		},
		{
			name:    "don't enqueue module task when there is an older blocking pending module task",
			active:  []*Task{},
			pending: []*Task{mod1TaskBlocking1, mod1Task1},
			want:    []*Task{mod1TaskBlocking1},
		},
		{
			name:    "don't enqueue run task when there is an older blocking pending run task sharing same workspace",
			active:  []*Task{},
			pending: []*Task{run1TaskBlocking1, run2Task1},
			want:    []*Task{run1TaskBlocking1},
		},
		{
			name:    "only enqueue one of three blocking workspace tasks sharing same module",
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
