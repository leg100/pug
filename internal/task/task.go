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

type Kind int

const (
	GlobalKind Kind = iota
	ModuleKind
	WorkspaceKind
	RunKind
)

type Task struct {
	resource.Resource

	Command  []string
	Args     []string
	Path     string
	Blocking bool
	State    Status
	Env      []string
	Kind     Kind
	Metadata map[string]string
	// Resource ID

	program   string
	exclusive bool

	// Nil until task has started
	proc *os.Process

	created time.Time
	updated time.Time

	// Nil until task finishes with an error
	Err error

	buf *buffer

	// lock to ensure task state is switched atomically.
	mu sync.Mutex

	// this channel is closed once the task is finished
	finished chan struct{}

	// call this whenever state is updated
	afterUpdate func(*Task)
}

type factory struct {
	program     string
	afterUpdate func(*Task)
}

func (f *factory) newTask(opts CreateOptions) (*Task, error) {
	return &Task{
		State:       Pending,
		Resource:    resource.New(&opts.Parent),
		created:     time.Now(),
		updated:     time.Now(),
		finished:    make(chan struct{}),
		buf:         newBuffer(),
		program:     f.program,
		Path:        opts.Path,
		Args:        opts.Args,
		exclusive:   opts.Exclusive,
		afterUpdate: f.afterUpdate,
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
	t.updated = time.Now()
	t.State = state
	if t.afterUpdate != nil {
		t.afterUpdate(t)
	}

	switch state {
	case Exited, Errored, Canceled:
		close(t.finished)
	}
}

func (t *Task) setErrored(err error) {
	t.Err = err
	t.updateState(Errored)
	//if t.AfterError != nil {
	//	t.AfterError(t, err)
	//}
}
