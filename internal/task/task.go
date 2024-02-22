package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	"github.com/leg100/pug/internal/resource"
)

// Kind differentiates tasks, i.e. Init, Plan, etc.
type Kind string

type Spec struct {
	Kind      Kind
	Parent    resource.Resource
	Args      []string
	Path      string
	Exclusive bool
}

type Scheduler interface {
	Handle(event resource.Event[*Task]) []*Task
}

type Task struct {
	Path          string
	WorkspaceName *string
	Scheduler     Scheduler

	program   string
	args      []string
	exclusive bool

	ID uuid.UUID

	// Nil until task has started
	proc *os.Process

	Kind  Kind
	State Status

	created time.Time
	updated time.Time

	Err error

	buf *buffer

	// lock to ensure task state is switched atomically.
	mu sync.Mutex

	// this channel is closed once the task is finished
	finished chan struct{}

	// call this whenever state is updated
	callback func(*Task)
}

type factory struct {
	program  string
	callback func(*Task)
}

func (f *factory) newTask(path string, args ...string) (*Task, error) {
	return &Task{
		State:    Pending,
		ID:       uuid.New(),
		created:  time.Now(),
		updated:  time.Now(),
		finished: make(chan struct{}),
		buf:      newBuffer(),
		program:  f.program,
		Path:     path,
		args:     args,
		callback: f.callback,
	}, nil
}

func (t *Task) id() string {
	return base58.Encode(t.ID[:])
}

func (t *Task) String() string      { return t.id() }
func (t *Task) Title() string       { return t.id() }
func (t *Task) Description() string { return string(t.State) }
func (t *Task) FilterValue() string { return t.id() }

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
	cmd := exec.Command(t.program, t.args...)
	cmd.Dir = t.Path
	cmd.Stdout = t.buf
	cmd.Stderr = t.buf

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
	if t.callback != nil {
		t.callback(t)
	}

	switch state {
	case Exited, Errored, Canceled:
		close(t.finished)
	}
}
