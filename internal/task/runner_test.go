package task

import (
	"slices"
	"testing"
)

func TestRunner_queue(t *testing.T) {
	// create runner with max number of tasks set to 3, and fire off 3 tasks
	runner := newRunner(3, &fakeTaskLister{})

	runner.process()
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
