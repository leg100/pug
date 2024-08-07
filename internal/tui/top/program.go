package top

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
	"github.com/stretchr/testify/require"
)

type Options struct {
	Modules    tui.ModuleService
	Workspaces tui.WorkspaceService
	States     tui.StateService
	Runs       tui.RunService
	Tasks      tui.TaskService
	Logger     *logging.Logger
	Workdir    internal.Workdir
	FirstPage  string
	MaxTasks   int
	Debug      bool
	Program    string
	Terragrunt bool
}

// New constructs and runs the TUI program.
func New(opts Options) error {
	p, err := newProgram(opts)
	if err != nil {
		return err
	}
	tp := tea.NewProgram(p.model,
		// Use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)
	// Relay events in background
	go func() {
		for msg := range p.ch {
			tp.Send(msg)
		}
	}()
	// Blocks until user quits
	_, err = tp.Run()
	return err
}

func NewTest(t *testing.T, opts Options, width, height int) *teatest.TestModel {
	p, err := newProgram(opts)
	require.NoError(t, err)

	tm := teatest.NewTestModel(t, p.model, teatest.WithInitialTermSize(width, height))
	t.Cleanup(func() {
		p.cleanup()
		tm.Quit()
	})

	// Relay events in background
	go func() {
		for msg := range p.ch {
			tm.Send(msg)
		}
	}()
	return tm
}

type program struct {
	model   tea.Model
	ch      chan tea.Msg
	cleanup func()
}

func newProgram(opts Options) (*program, error) {
	m, err := newModel(opts)
	if err != nil {
		return nil, err
	}
	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	ch := make(chan tea.Msg)
	wg := sync.WaitGroup{} // sync closure of subscriptions

	logEvents := opts.Logger.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range logEvents {
			ch <- ev
		}
		wg.Done()
	}()

	modEvents := opts.Modules.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range modEvents {
			ch <- ev
		}
		wg.Done()
	}()

	wsEvents := opts.Workspaces.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range wsEvents {
			ch <- ev
		}
		wg.Done()
	}()

	stateEvents := opts.States.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range stateEvents {
			ch <- ev
		}
		wg.Done()
	}()

	runEvents := opts.Runs.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range runEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskEvents := opts.Tasks.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range taskEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskGroupEvents := opts.Tasks.SubscribeGroups()
	wg.Add(1)
	go func() {
		for ev := range taskGroupEvents {
			ch <- ev
		}
		wg.Done()
	}()

	// Automatically load workspaces whenever modules are loaded.
	opts.Workspaces.LoadWorkspacesUponModuleLoad(opts.Modules.Subscribe())

	// Whenever a workspace is added, pull its state
	go func() {
		for event := range opts.Workspaces.Subscribe() {
			if event.Type == resource.CreatedEvent {
				_, _ = opts.States.CreateReloadTask(event.Payload.ID)
			}
		}
	}()

	// cleanup function to be invoked when program is terminated.
	cleanup := func() {
		// Close subscriptions
		opts.Logger.Shutdown()
		opts.Tasks.Shutdown()
		opts.Tasks.ShutdownGroups()
		opts.Modules.Shutdown()
		opts.Workspaces.Shutdown()
		opts.Runs.Shutdown()
		opts.States.Shutdown()

		// Wait for relays to finish before closing channel, to avoid sends
		// to a closed channel, which would result in a panic.
		wg.Wait()
		close(ch)
	}

	return &program{
		cleanup: cleanup,
		ch:      ch,
		model:   m,
	}, nil
}
