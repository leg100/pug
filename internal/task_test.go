package internal

import (
	"bufio"
	"testing"

	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_stdout(t *testing.T) {
	task := newTask(taskspec{
		prog: "./testdata/task",
	})
	require.True(t, task.start(), task.err)
	go func() { task.wait() }()

	scanner := bufio.NewScanner(task.out)
	want := []string{"foo", "bar", "baz", "bye"}
	for i := 0; scanner.Scan(); i++ {
		assert.Equal(t, want[i], scanner.Text())
	}
	require.NoError(t, scanner.Err())
	assert.NoError(t, task.err)
	assert.Equal(t, exited, task.state)
}

func TestTask_cancel(t *testing.T) {
	task := newTask(taskspec{
		prog: "./testdata/killme",
	})
	require.True(t, task.start(), task.err)
	done := make(chan struct{})
	go func() {
		task.wait()
		done <- struct{}{}
	}()

	assert.Equal(t, "ok, you can kill me now\n", <-iochan.DelimReader(task.out, '\n'))
	task.cancel()
	<-done
	assert.NoError(t, task.err)
	assert.Equal(t, exited, task.state)
}
