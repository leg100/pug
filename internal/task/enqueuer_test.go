package task

import (
	"slices"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueuer(t *testing.T) {
	t.Parallel()

	mod1ID := resource.NewID(resource.Module)
	ws1ID := resource.NewID(resource.Workspace)

	mod1Task1 := newTestTask(t, Spec{ModuleID: &mod1ID})
	mod1TaskBlocking1 := newTestTask(t, Spec{ModuleID: &mod1ID, Blocking: true})

	ws1Task1 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID})
	ws1Task2 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID})

	ws1TaskBlocking1 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, Blocking: true})
	ws1TaskBlocking2 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, Blocking: true})
	ws1TaskBlocking3 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, Blocking: true})
	ws1TaskImmediate := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, Immediate: true})
	ws1TaskDependOnTask1 := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, dependsOn: []resource.ID{ws1Task1.ID}})

	ws1TaskCompleted := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID})
	ws1TaskCompleted.updateState(Exited)

	ws1TaskDependOnCompletedTask := newTestTask(t, Spec{ModuleID: &mod1ID, WorkspaceID: &ws1ID, dependsOn: []resource.ID{ws1TaskCompleted.ID}})

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
			want:    nil,
		},
		{
			name:    "don't enqueue workspace task when there is an active blocking module task sharing same module",
			active:  []*Task{mod1TaskBlocking1},
			pending: []*Task{ws1Task1},
			want:    nil,
		},
		{
			name:    "don't enqueue module task when there is an older blocking pending module task",
			active:  []*Task{},
			pending: []*Task{mod1TaskBlocking1, mod1Task1},
			want:    []*Task{mod1TaskBlocking1},
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
			want:    nil,
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

func newTestTask(t *testing.T, spec Spec) *Task {
	f := &factory{counter: internal.Int(0)}
	task, err := f.newTask(spec)
	require.NoError(t, err)
	return task
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
