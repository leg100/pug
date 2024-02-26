package task

import (
	"slices"
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestEnqueuer(t *testing.T) {
	tests := []struct {
		name string
		// Module status
		status module.Status
		// Path of module in task event parameter
		path string
		// Active tasks
		active []*task.Task
		// Pending tasks
		pending []*task.Task
		// Want these tasks enqueued
		want []*task.Task
	}{
		{
			"don't enqueue tasks for uninitialized module",
			module.Uninitialized,
			"a/b/c",
			nil,
			nil,
			nil,
		},
		{
			"don't enqueue tasks when there is an active init task",
			module.Initialized,
			"a/b/c",
			[]*task.Task{{Kind: module.InitTask}},
			nil,
			nil,
		},
		{
			"no tasks to enqueue",
			module.Initialized,
			"a/b/c",
			nil,
			nil,
			nil,
		},
		{
			"enqueue plan",
			module.Initialized,
			"a/b/c",
			nil,
			[]*task.Task{
				{Kind: PlanTask},
			},
			[]*task.Task{{Kind: PlanTask}},
		},
		{
			"enqueue init but not newer plan",
			module.Initialized,
			"a/b/c",
			nil,
			[]*task.Task{
				{Kind: module.InitTask},
				{Kind: PlanTask},
			},
			[]*task.Task{{Kind: module.InitTask}},
		},
		{
			"enqueue plan but not newer init",
			module.Initialized,
			"a/b/c",
			nil,
			[]*task.Task{
				{Kind: PlanTask},
				{Kind: module.InitTask},
			},
			[]*task.Task{{Kind: PlanTask}},
		},
		{
			"enqueue multiple tasks but not newer init nor tasks following init",
			module.Initialized,
			"a/b/c",
			nil,
			[]*task.Task{
				{Kind: PlanTask},
				{Kind: ApplyTask},
				{Kind: PlanTask},
				{Kind: module.InitTask},
				{Kind: PlanTask},
				{Kind: PlanTask},
			},
			[]*task.Task{
				{Kind: PlanTask},
				{Kind: ApplyTask},
				{Kind: PlanTask},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := enqueuer{
				modules: &fakeModuleGetter{
					m: &module.Module{Status: tt.status},
				},
				tasks: &fakeTaskLister{
					pending: tt.pending,
					active:  tt.active,
				},
			}
			event := resource.Event[*task.Task]{
				Payload: &task.Task{
					Path: tt.path,
				},
			}
			assert.Equal(t, tt.want, e.Enqueue(event))
		})
	}
}

type fakeModuleGetter struct {
	m *module.Module
}

func (f *fakeModuleGetter) Get(path string) (*module.Module, error) {
	return f.m, nil
}

type fakeTaskLister struct {
	pending, active []*task.Task
}

func (f *fakeTaskLister) List(opts task.ListOptions) []*task.Task {
	if slices.Equal(opts.Status, []task.Status{task.Queued, task.Running}) {
		return f.active
	}
	if slices.Equal(opts.Status, []task.Status{task.Pending}) {
		return f.pending
	}
	return nil
}
