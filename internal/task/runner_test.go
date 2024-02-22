package task

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_runnable(t *testing.T) {
	f := &factory{program: "../testdata/task"}

	t1, _ := f.newTask(".")
	t2, _ := f.newTask(".")
	t3, _ := f.newTask(".")
	t1.updateState(Queued)
	t2.updateState(Queued)
	t3.updateState(Queued)

	runner := newRunner(3, &fakeTaskLister{
		queued: []*Task{t1, t2, t3},
	})

	got := runner.runnable()
	assert.Len(t, got, 3)
	assert.Equal(t, t1, got[0])
	assert.Equal(t, t2, got[1])
	assert.Equal(t, t3, got[2])
}

//
//func TestRunner_exclusive(t *testing.T) {
//	runner := NewRunner(0, "../testdata/killme")
//	require.Equal(t, 0, len(runner.exclusive))
//
//	// run an exclusive task
//	task, err := runner.Run(Spec{Exclusive: true})
//	require.NoError(t, err)
//	// wait for it to enter running state
//	assert.Equal(t, Running, <-task.Events, task.Err)
//
//	// should have taken exclusive slot
//	require.Equal(t, 1, len(runner.exclusive))
//
//	// start another exclusive task; should be queued
//	another, err := runner.Run(Spec{Exclusive: true})
//	require.NoError(t, err)
//
//	// kill running task
//	task.cancel()
//	assert.Equal(t, Errored, <-task.Events)
//
//	// other task should now run
//	assert.Equal(t, Running, <-another.Events)
//}

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
