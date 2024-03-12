package task

import (
	"bufio"
	"io"
	"testing"

	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_stdout(t *testing.T) {
	f := factory{program: "../testdata/task"}
	task, err := f.newTask(CreateOptions{})
	require.NoError(t, err)
	task.updateState(Queued)
	waitfn, err := task.start()
	require.NoError(t, err)
	waitfn()

	scanner := bufio.NewScanner(task.NewReader())
	want := []string{"foo", "bar", "baz", "bye"}
	for i := 0; scanner.Scan(); i++ {
		assert.Equal(t, want[i], scanner.Text())
	}
	require.NoError(t, scanner.Err())
	assert.NoError(t, task.Err)
	assert.Equal(t, Exited, task.State)

	// create another reader, to demonstrate that reading resets to the
	// beginning
	got, err := io.ReadAll(task.NewReader())
	require.NoError(t, err)
	assert.Equal(t, "foo\nbar\nbaz\nbye\n", string(got))
}

func TestTask_cancel(t *testing.T) {
	f := factory{program: "../testdata/killme"}
	task, err := f.newTask(CreateOptions{})
	require.NoError(t, err)
	task.updateState(Queued)

	done := make(chan struct{})
	go func() {
		waitfn, err := task.start()
		require.NoError(t, err)
		waitfn()
		done <- struct{}{}
	}()

	assert.Equal(t, "ok, you can kill me now\n", <-iochan.DelimReader(task.NewReader(), '\n'))
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
