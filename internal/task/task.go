package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
)

// Task is an execution of a CLI program.
type Task struct {
	resource.Common

	Command       []string
	Args          []string
	Path          string
	Blocking      bool
	State         Status
	JSON          bool
	Immediate     bool
	AdditionalEnv []string
	DependsOn     []resource.ID
	description   string

	program   string
	exclusive bool

	// Nil until task has started
	proc *os.Process

	Created time.Time
	Updated time.Time

	// Nil until task finishes with an error
	Err error

	// stdout contains only the stdout stream
	stdout *buffer
	// combined contains both the stderr and stdout streams
	combined *buffer

	// lock to ensure task state is switched atomically.
	mu sync.Mutex

	// this channel is closed once the task is finished
	finished chan struct{}

	// timestamps records the time at which the task transitioned into a status
	// and out of a status.
	timestamps map[Status]statusTimestamps

	// Retain a copy of the Spec used to originally create the task so that
	// the task can be retried.
	Spec Spec

	// Call this function after the CLI program has finished successfully but
	// before the task has exited. If a non-nil error is returned then the task
	// is deemed to have errored.
	BeforeExited func(*Task) error
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
	// Call this function after the task terminates for whatever reason.
	AfterFinish func(*Task)

	// call this whenever state is updated
	afterUpdate func(*Task)

	// call this once the task has terminated
	afterFinish func(*Task)
}

type factory struct {
	counter   *int
	program   string
	publisher resource.Publisher[*Task]
	workdir   internal.Workdir
	// Additional user-supplied environment variables.
	userEnvs []string
	// Additional user-supplied CLI args.
	userArgs []string
	// Terragrunt mode
	terragrunt bool
}

// TODO: check presence of mandatory options
func (f *factory) newTask(spec Spec) *Task {
	// In terragrunt mode add default terragrunt flags
	args := append(f.userArgs, spec.Args...)
	if f.terragrunt {
		args = append(args, "--terragrunt-non-interactive")
	}

	return &Task{
		Common:        resource.New(resource.Task, spec.Parent),
		State:         Pending,
		Created:       time.Now(),
		Updated:       time.Now(),
		finished:      make(chan struct{}),
		stdout:        newBuffer(),
		combined:      newBuffer(),
		program:       f.program,
		Command:       spec.Command,
		Path:          filepath.Join(f.workdir.String(), spec.Path),
		Args:          args,
		AdditionalEnv: append(f.userEnvs, spec.Env...),
		JSON:          spec.JSON,
		Blocking:      spec.Blocking,
		DependsOn:     spec.DependsOn,
		Immediate:     spec.Immediate,
		exclusive:     spec.Exclusive,
		description:   spec.Description,
		Spec:          spec,
		BeforeExited:  spec.BeforeExited,
		AfterExited:   spec.AfterExited,
		AfterError:    spec.AfterError,
		AfterCanceled: spec.AfterCanceled,
		AfterRunning:  spec.AfterRunning,
		AfterQueued:   spec.AfterQueued,
		AfterFinish:   spec.AfterFinish,
		// Publish an event whenever task state is updated
		afterUpdate: func(t *Task) {
			if f.publisher != nil {
				f.publisher.Publish(resource.UpdatedEvent, t)
			}
		},
		// Decrement live task counter whenever task terminates
		afterFinish: func(t *Task) {
			*f.counter--
		},
		timestamps: map[Status]statusTimestamps{
			Pending: {
				started: time.Now(),
			},
		},
	}
}

func (t *Task) String() string {
	if t.description != "" {
		return t.description
	}
	return strings.Join(t.Command, " ")
}

// NewReader provides a reader from which to read the task output from start to
// end. Set combined to true to receieve stderr as well as stdout.
func (t *Task) NewReader(combined bool) io.Reader {
	if combined {
		return &reader{buf: t.combined}
	}
	return &reader{buf: t.stdout}
}

func (t *Task) IsActive() bool {
	switch t.State {
	case Queued, Running:
		return true
	default:
		return false
	}
}

// Elapsed returns the length of time the task has been in the given status.
func (t *Task) Elapsed(s Status) time.Duration {
	st, ok := t.timestamps[s]
	if !ok {
		return 0
	}
	return st.Elapsed()
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

func (t *Task) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID.String()),
		slog.Any("command", t.Command),
		slog.Any("args", t.Args),
		slog.Any("deps", t.DependsOn),
	)
}

// cancel the task - if it is queued it'll skip the running state and enter the
// exited state
func (t *Task) cancel() error {
	// lock task state so that cancelation can atomically both inspect current
	// state and update state
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.State {
	case Exited, Errored, Canceled:
		return errors.New("task has already finished")
	case Pending, Queued:
		t.updateState(Canceled)
		return nil
	default: // running
		return t.proc.Signal(os.Interrupt)
	}
}

func (t *Task) start(ctx context.Context) (func(), error) {
	// Use the provided context to kill the program if the context becomes done,
	// but also to prevent the program from starting if the context becomes done.
	cmd := exec.CommandContext(ctx, t.program, append(t.Command, t.Args...)...)
	cmd.Cancel = func() error {
		// Kill program gracefully
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.Dir = t.Path
	cmd.Stdout = io.MultiWriter(t.stdout, t.combined)
	cmd.Stderr = t.combined
	cmd.Env = append(t.AdditionalEnv, os.Environ()...)

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

// record time at which current status finished
func (t *Task) recordStatusEndTime(now time.Time) {
	currentStateTimestamps := t.timestamps[t.State]
	currentStateTimestamps.ended = now
	t.timestamps[t.State] = currentStateTimestamps
}

func (t *Task) updateState(state Status) {
	now := time.Now()
	t.Updated = now

	// record times at which old status ended, and new status started
	t.recordStatusEndTime(now)
	t.timestamps[state] = statusTimestamps{
		started: now,
	}

	t.State = state

	if t.afterUpdate != nil {
		t.afterUpdate(t)
	}

	if t.IsFinished() {
		// Close log streams
		t.stdout.Close()
		t.combined.Close()

		if t.State == Exited && t.BeforeExited != nil {
			if err := t.BeforeExited(t); err != nil {
				t.State = Errored
			}
		}

		t.recordStatusEndTime(now)

		close(t.finished)
		if t.afterFinish != nil {
			t.afterFinish(t)
		}
		if t.AfterFinish != nil {
			t.AfterFinish(t)
		}
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
	case Errored:
		if t.AfterError != nil {
			t.AfterError(t)
		}
	case Exited:
		if t.AfterExited != nil {
			t.AfterExited(t)
		}
	}
}
