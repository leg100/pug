package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_queue(t *testing.T) {
	// create runner with max number of tasks set to 3, and fire off 3 tasks
	runner := NewRunner(3)
	batch1 := make([]*task, 3)
	for i := 0; i < 3; i++ {
		task, err := runner.run(taskspec{
			prog: "./testdata/killme",
		})
		require.NoError(t, err)
		// wait for it to enter running state
		assert.Equal(t, running, <-task.events)
		batch1[i] = task
	}
	// start further batch of 3 tasks, which will be queued
	batch2 := make([]*task, 3)
	for i := 0; i < 3; i++ {
		task, err := runner.run(taskspec{
			prog: "./testdata/killme",
		})
		require.NoError(t, err)
		batch2[i] = task
	}
	// kill first batch
	for i := 0; i < 3; i++ {
		batch1[i].cancel()
		assert.Equal(t, errored, <-batch1[i].events)
	}
	// second batch should now run
	for i := 0; i < 3; i++ {
		// wait for it to enter running state
		assert.Equal(t, running, <-batch2[i].events)
	}
}

func TestRunner_exclusive(t *testing.T) {
	runner := NewRunner(0)
	require.Equal(t, 0, len(runner.exclusive))

	// run an exclusive task
	task, err := runner.run(taskspec{
		prog:      "./testdata/killme",
		exclusive: true,
	})
	require.NoError(t, err)
	// wait for it to enter running state
	assert.Equal(t, running, <-task.events)

	// should have taken exclusive slot
	require.Equal(t, 1, len(runner.exclusive))

	// start another exclusive task; should be queued
	another, err := runner.run(taskspec{
		prog:      "./testdata/killme",
		exclusive: true,
	})
	require.NoError(t, err)

	// kill running task
	task.cancel()
	assert.Equal(t, errored, <-task.events)

	// other task should now run
	assert.Equal(t, running, <-another.events)
}
