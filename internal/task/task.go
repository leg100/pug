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

// Identifier uniquely identifies the type of task.
type Identifier string

// Task is an execution of a CLI program.
type Task struct {
	resource.ID

	ModuleID            *resource.ID
	WorkspaceID         *resource.ID
	Identifier          Identifier
	Program             string
	Args                []string
	AdditionalExecution *Execution
	Path                string
	Blocking            bool
	State               Status
	JSON                bool
	Immediate           bool
	AdditionalEnv       []string
	DependsOn           []resource.ID
	// Summary summarises the outcome of a task to the end-user.
	Summary     Summary
	Description string

	exclusive bool
	// terragrunt is true if terragrunt is in use.
	terragrunt bool

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

	AfterCreate   func(*Task)
	AfterQueued   func(*Task)
	AfterRunning  func(*Task)
	BeforeExited  func(*Task) (Summary, error)
	AfterExited   func(*Task)
	AfterError    func(*Task)
	AfterCanceled func(*Task)
	AfterFinish   func(*Task)
	afterUpdate   func(*Task)
	afterFinish   func(*Task)
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

// Summary summarises the outcome of a task.
type Summary interface {
	String() string
}

// TODO: check presence of mandatory options
// TODO: use values from spec directly - embed spec in task and use those values
func (f *factory) newTask(spec Spec) (*Task, error) {
	if spec.WorkspaceID != nil && spec.ModuleID == nil {
		return nil, errors.New("workspace ID cannot be provided without module ID")
	}
	task := &Task{
		ID:                  resource.NewID(resource.Task),
		ModuleID:            spec.ModuleID,
		WorkspaceID:         spec.WorkspaceID,
		Identifier:          spec.Identifier,
		State:               Pending,
		Created:             time.Now(),
		Updated:             time.Now(),
		finished:            make(chan struct{}),
		stdout:              newBuffer(),
		combined:            newBuffer(),
		terragrunt:          f.terragrunt,
		Path:                filepath.Join(f.workdir.String(), spec.Path),
		AdditionalExecution: spec.AdditionalExecution,
		AdditionalEnv:       append(f.userEnvs, spec.Env...),
		JSON:                spec.JSON,
		Blocking:            spec.Blocking,
		DependsOn:           spec.dependsOn,
		Immediate:           spec.Immediate,
		exclusive:           spec.Exclusive,
		Description:         spec.Description,
		Spec:                spec,
		AfterCreate:         spec.AfterCreate,
		AfterRunning:        spec.AfterRunning,
		AfterQueued:         spec.AfterQueued,
		BeforeExited:        spec.BeforeExited,
		AfterExited:         spec.AfterExited,
		AfterError:          spec.AfterError,
		AfterCanceled:       spec.AfterCanceled,
		AfterFinish:         spec.AfterFinish,
		// Publish an event whenever task state is updated
		afterUpdate: func(t *Task) {
			// TODO: remove nil-check that is only here to ensure tests don't
			// have to mock publisher...
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
	// Determine the program and the args to pass to program.
	if spec.Execution.Program == "" {
		// Is terraform task
		task.Program = f.program
		task.Args = spec.Execution.TerraformCommand
	} else {
		// Non-terraform task
		task.Program = spec.Execution.Program
	}
	task.Args = append(task.Args, f.userArgs...)
	task.Args = append(task.Args, spec.Execution.Args...)

	// If description is not explicitly set then set it using provided terraform
	// commands or - if this is not a terraform execution - then using the
	// provided program.
	if spec.Description == "" {
		if len(spec.Execution.TerraformCommand) > 0 {
			task.Description = strings.Join(spec.Execution.TerraformCommand, " ")
		} else {
			task.Description = spec.Execution.Program
		}
	}

	// In terragrunt mode add default terragrunt flags
	//
	// TODO: introduce a better way to determine whether terrarunt is in use.
	// Perhaps use constants for terraform, tofu, and terragrunt.
	if task.Program == "terragrunt" && f.terragrunt {
		task.AdditionalEnv = append(task.AdditionalEnv, "TERRAGRUNT_FORWARD_TF_STDOUT=1")
		task.Args = append(task.Args, "--terragrunt-non-interactive")
	}
	return task, nil
}

func (t *Task) String() string {
	return t.Description
}

// NewReader returns a reader which contains what has been written thus far to
// the task buffer.
// Set combined to true to receieve stderr as well as stdout.
func (t *Task) NewReader(combined bool) io.Reader {
	if combined {
		return t.combined.NewReader()
	}
	return t.stdout.NewReader()
}

// NewStreamer returns a stream of output from the task; the channel is closed
// when the task has finished.
func (t *Task) NewStreamer() <-chan []byte {
	return t.combined.Stream()
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
	attrs := []slog.Attr{
		slog.String("id", t.ID.String()),
		slog.Any("program", t.Program),
		slog.Any("args", t.Args),
	}
	if t.terragrunt {
		attrs = append(attrs, slog.Any("deps", t.DependsOn))
	}
	if t.Summary != nil {
		// If task summary implements LogValuer and returns a group of
		// attributes then "flatten" them into the rest of the task's
		// attributes.
		// Otherwise add its value to an attribute with the "summary" key.
		if lg, ok := t.Summary.(slog.LogValuer); ok {
			if lg.LogValue().Kind() == slog.KindGroup {
				attrs = append(attrs, lg.LogValue().Group()...)
			} else {
				attrs = append(attrs, slog.Any("summary", t.Summary))
			}
		} else {
			attrs = append(attrs, slog.String("summary", t.Summary.String()))
		}
	}
	return slog.GroupValue(attrs...)
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
	cmd := t.execute(ctx, t.Program, t.Args)

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
		if err := cmd.Wait(); err != nil {
			state = Errored
			t.Err = fmt.Errorf("task failed: %w", err)
		} else if t.AdditionalExecution != nil {
			// Execute additional program.
			cmd = t.execute(ctx, t.AdditionalExecution.Program, t.AdditionalExecution.Args)
			if err := cmd.Run(); err != nil {
				state = Errored
				t.Err = fmt.Errorf("task failed: %w", err)
			}
		}

		t.mu.Lock()
		t.updateState(state)
		t.mu.Unlock()
	}
	return wait, nil
}

func (t *Task) execute(ctx context.Context, program string, args []string) *exec.Cmd {
	// Use the provided context to kill the program if the context becomes done,
	// but also to prevent the program from starting if the context becomes done.
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Cancel = func() error {
		// Kill program gracefully
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.Dir = t.Path
	cmd.Stdout = io.MultiWriter(t.stdout, t.combined)
	cmd.Stderr = t.combined
	cmd.Env = append(t.AdditionalEnv, os.Environ()...)
	return cmd
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

	// Close output streams. It's important this is done before BeforeExited is
	// called because it may want to consume the output streams until EOF.
	if state.IsFinal() {
		t.stdout.Close()
		t.combined.Close()
	}

	// Before task exits trigger callback and if it fails set task's status to
	// errored. Otherwise the returned summary summarises the task's outcome.
	if state == Exited && t.BeforeExited != nil {
		summary, err := t.BeforeExited(t)
		if err != nil {
			state = Errored
			t.Err = err
		}
		t.Summary = summary
	}

	t.State = state
	if t.afterUpdate != nil {
		t.afterUpdate(t)
	}

	if t.State.IsFinal() {
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
