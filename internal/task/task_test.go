package task

import (
	"context"
	"io"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_NewReader(t *testing.T) {
	t.Parallel()

	f := factory{
		counter:   internal.Int(0),
		program:   "./testdata/task",
		publisher: &fakePublisher[*Task]{},
	}
	task, err := f.newTask(Spec{})
	require.NoError(t, err)
	task.updateState(Queued)
	waitfn, err := task.start(context.Background())
	require.NoError(t, err)
	waitfn()

	got, err := io.ReadAll(task.NewReader(false))
	require.NoError(t, err)
	assert.Equal(t, "foo\nbar\nbaz\nbye\n", string(got))

	// create another reader, to demonstrate that reading resets to the
	// beginning
	got, err = io.ReadAll(task.NewReader(false))
	require.NoError(t, err)
	assert.Equal(t, "foo\nbar\nbaz\nbye\n", string(got))

	// this time, create a combined reader, to demonstrate reading both stdout
	// and stderr
	got, err = io.ReadAll(task.NewReader(true))
	require.NoError(t, err)
	// stderr and stdout streams can be interwoven, so we use assert.Contains
	// rather than assert.Equal
	assert.Contains(t, string(got), "foo\nbar\nbaz\nbye\n")
	assert.Contains(t, string(got), "err")
}

func TestTask_cancel(t *testing.T) {
	t.Parallel()

	f := factory{
		counter:   internal.Int(0),
		program:   "./testdata/killme",
		publisher: &fakePublisher[*Task]{},
	}
	task, err := f.newTask(Spec{})
	require.NoError(t, err)

	task.updateState(Queued)

	done := make(chan struct{})
	go func() {
		waitfn, err := task.start(context.Background())
		require.NoError(t, err)
		waitfn()
		done <- struct{}{}
	}()

	assert.Equal(t, []byte("ok, you can kill me now\n"), <-task.NewStreamer())
	task.cancel()
	<-done
	assert.NoError(t, task.Err)
	assert.Equal(t, Exited, task.State)
}

// func TestTask_WaitFor_immediateExit(t *testing.T) {
// 	f := factory{program: "../testdata/task"}
// 	task, err := f.newTask(".")
// 	require.NoError(t, err)
// 	task.run()()
//
// 	require.True(t, task.WaitFor(Exited))
// }
//
// func TestTask_WaitFor(t *testing.T) {
// 	f := factory{program: "../testdata/killme"}
// 	task, err := f.newTask(".")
// 	require.NoError(t, err)
//
// 	// wait for task to exit in background
// 	got := make(chan bool)
// 	go func() {
// 		got <- task.WaitFor(Exited)
// 	}()
//
// 	// start task in background
// 	go func() {
// 		task.run()()
// 	}()
//
// 	// wait for task to start
// 	assert.Equal(t, "ok, you can kill me now\n", <-iochan.DelimReader(task.NewReader(), '\n'))
// 	// then cancel
// 	task.cancel()
//
// 	// verify task exits
// 	require.True(t, <-got)
// }
