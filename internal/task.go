package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type taskstate string

const (
	queued   taskstate = "queued"
	running  taskstate = "running"
	exited   taskstate = "exited"
	errored  taskstate = "errored"
	canceled taskstate = "canceled"
)

type task struct {
	id string

	out    io.Reader
	closer io.Closer
	state  taskstate
	err    error
	proc   *exec.Cmd
	events chan taskstate

	// lock to ensure task state is switched atomically.
	mu sync.Mutex
}

type taskspec struct {
	prog      string
	args      []string
	path      string
	exclusive bool
}

func newTask(spec taskspec) *task {
	cmd := exec.Command(spec.prog, spec.args...)
	cmd.Dir = spec.path
	// pipe stdout and stderr
	rout, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	// task starts life in a queued state
	return &task{
		proc:   cmd,
		out:    rout,
		closer: w,
		state:  queued,
		// there should never be more than 2 events:
		// 2 events: -> running -> {exited,errored,canceled}
		// 1 events: -> {errored,canceled}
		// -> exited
		events: make(chan taskstate, 2),
	}
}

func (t *task) String() string      { return t.id }
func (t *task) Title() string       { return t.id }
func (t *task) Description() string { return string(t.state) }
func (t *task) FilterValue() string { return t.id }

// cancel the task - if it is queued it'll skip the running state and enter the
// exited state
func (t *task) cancel() {
	// lock task state so that cancelation can atomically both inspect current
	// state and update state
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.state {
	case exited, canceled:
		// take no action if already exited
		return
	case queued:
		t.updateState(canceled)
		t.cleanup(nil)
		return
	default: // running
		// ignore any errors from signal; instead take a "best effort" approach
		// to canceling
		_ = t.proc.Process.Signal(os.Interrupt)
		return
	}
}

func (t *task) start() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.proc.Start(); err != nil {
		t.updateState(exited)
		t.cleanup(fmt.Errorf("starting task: %w", err))
		return false
	}
	t.updateState(running)
	return true
}

func (t *task) wait() {
	var err error
	if werr := t.proc.Wait(); werr != nil {
		t.updateState(errored)
		err = fmt.Errorf("task failed: %w", werr)
	}
	t.updateState(exited)
	t.cleanup(err)
}

func (t *task) cleanup(err error) {
	// close stdout/stderr pipe upon completion, letting consumers know there is
	// nothing more to be read
	t.closer.Close()
	t.err = err
	close(t.events)
}

func (t *task) updateState(state taskstate) {
	t.state = state
	t.events <- state
}
