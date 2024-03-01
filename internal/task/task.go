package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/leg100/pug/internal/resource"
)

type Task struct {
	resource.Resource

	Command  []string
	Args     []string
	Path     string
	Blocking bool
	State    Status
	Env      []string

	program   string
	exclusive bool

	// Nil until task has started
	proc *os.Process

	Created time.Time
	Updated time.Time

	// Nil until task finishes with an error
	Err error

	buf *buffer

	// lock to ensure task state is switched atomically.
	mu sync.Mutex

	// this channel is closed once the task is finished
	finished chan struct{}

	// call this whenever state is updated
	afterUpdate func(*Task)
	// Call this function after the task has successfully finished
	AfterExited func(*Task)
	// Call this function after the task is enqueued.
	AfterQueued func(*Task)
	// Call this function after the task starts running.
	AfterRunning func(*Task)
	// Call this function after the task fails with an error
	AfterError func(*Task)
	// Call this function after the task is successfully canceled
	AfterCanceled func(*Task)
}

type factory struct {
	program     string
	afterUpdate func(*Task)
}

type CreateOptions struct {
	// Resource that the task belongs to.
	Parent resource.Resource
	// Program command and any sub commands, e.g. plan, state rm, etc.
	Command []string
	// Args to pass to program.
	Args []string
	// Path in which to execute the program - assumed be the terraform module's
	// path.
	Path string
	// Environment variables.
	Env []string
	// A blocking task blocks other tasks from running on the module or
	// workspace.
	Blocking bool
	// Globally exclusive task - at most only one such task can be running
	Exclusive bool
	// Call this function after the task has successfully finished
	AfterExited func(*Task)
	// Call this function after the task is enqueued.
	AfterQueued func(*Task)
	// Call this function after the task starts running.
	AfterRunning func(*Task)
	// Call this function after the task fails with an error
	AfterError func(*Task)
	// Call this function after the task is successfully canceled
	AfterCanceled func(*Task)
	// Call this function after the task is successfully created
	AfterCreate func(*Task)
}

func (f *factory) newTask(opts CreateOptions) (*Task, error) {
	return &Task{
		State:         Pending,
		Resource:      resource.New(resource.Task, "", &opts.Parent),
		Created:       time.Now(),
		Updated:       time.Now(),
		finished:      make(chan struct{}),
		buf:           newBuffer(),
		program:       f.program,
		Path:          opts.Path,
		Args:          opts.Args,
		exclusive:     opts.Exclusive,
		afterUpdate:   f.afterUpdate,
		AfterExited:   opts.AfterExited,
		AfterError:    opts.AfterError,
		AfterCanceled: opts.AfterCanceled,
		AfterRunning:  opts.AfterRunning,
		AfterQueued:   opts.AfterQueued,
	}, nil
}

func (t *Task) String() string      { return t.Resource.String() }
func (t *Task) Title() string       { return t.Resource.String() }
func (t *Task) Description() string { return string(t.State) }
func (t *Task) FilterValue() string { return t.Resource.String() }

// NewReader provides a reader from which to read the task output from start to
// end.
func (t *Task) NewReader() io.Reader {
	return &reader{buf: t.buf}
}

func (t *Task) IsActive() bool {
	switch t.State {
	case Queued, Running:
		return true
	default:
		return false
	}
}

func (t *Task) IsFinished() bool {
	switch t.State {
	case Errored, Exited, Canceled:
		return true
	default:
		return false
	}
}

// Wait for task to complete successfully. If the task completes unsuccessfully
// then the returned error is non-nil.
func (t *Task) Wait() error {
	<-t.finished
	if t.State != Exited {
		return t.Err
	}
	return nil
}

// cancel the task - if it is queued it'll skip the running state and enter the
// exited state
func (t *Task) cancel() {
	// lock task state so that cancelation can atomically both inspect current
	// state and update state
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.State {
	case Exited, Errored, Canceled:
		// silently take no action if already finished
		return
	case Pending, Queued:
		t.updateState(Canceled)
		return
	default: // running
		// ignore any errors from signal; instead take a "best effort" approach
		// to canceling
		_ = t.proc.Signal(os.Interrupt)
		return
	}
}

func (t *Task) start() (func(), error) {
	cmd := exec.Command(t.program, append(t.Command, t.Args...)...)
	cmd.Dir = t.Path
	cmd.Stdout = t.buf
	cmd.Stderr = t.buf
	cmd.Env = t.Env

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != Queued {
		return nil, errors.New("invalid state transition")
	}

	if err := cmd.Start(); err != nil {
		t.updateState(Errored)
		t.Err = fmt.Errorf("starting task: %w", err)
		return nil, err
	}
	t.updateState(Running)
	// save reference to process so that it can be cancelled via cancel()
	t.proc = cmd.Process

	wait := func() {
		state := Exited
		if werr := cmd.Wait(); werr != nil {
			state = Errored
			t.Err = fmt.Errorf("task failed: %w", werr)
		}

		t.mu.Lock()
		t.updateState(state)
		t.mu.Unlock()
	}
	return wait, nil
}

func (t *Task) updateState(state Status) {
	t.Updated = time.Now()
	t.State = state
	if t.afterUpdate != nil {
		t.afterUpdate(t)
	}

	switch state {
	case Queued:
		if t.AfterQueued != nil {
			t.AfterQueued(t)
		}
	case Running:
		if t.AfterRunning != nil {
			t.AfterRunning(t)
		}
	case Canceled:
		if t.AfterCanceled != nil {
			t.AfterCanceled(t)
		}
		close(t.finished)
	case Errored:
		if t.AfterError != nil {
			t.AfterError(t)
		}
		close(t.finished)
	case Exited:
		if t.AfterExited != nil {
			t.AfterExited(t)
		}
		close(t.finished)
	}
}

func (t *Task) setErrored(err error) {
	t.Err = err
	t.updateState(Errored)
}
