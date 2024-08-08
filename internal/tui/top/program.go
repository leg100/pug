package top

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/require"
)

// Options for starting the TUI.
type Options struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	States     *state.Service
	Runs       *run.Service
	Tasks      *task.Service
	Logger     *logging.Logger
	Workdir    internal.Workdir
	FirstPage  string
	MaxTasks   int
	Debug      bool
	Program    string
	Terragrunt bool
}

// Start starts the TUI and blocks until the user exits.
func Start(cfg app.Config) error {
	p, err := newProgram(cfg)
	if err != nil {
		return err
	}
	defer p.cleanup()

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

// StartTest starts the TUI and returns a test model for testing purposes.
func StartTest(t *testing.T, cfg app.Config, width, height int) *teatest.TestModel {
	p, err := newProgram(cfg)
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

func newProgram(cfg app.Config) (*program, error) {
	app, err := app.New(cfg)
	if err != nil {
		return nil, err
	}
	m, err := newModel(cfg, app)
	if err != nil {
		app.Cleanup()
		return nil, err
	}
	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	ch := make(chan tea.Msg)
	wg := sync.WaitGroup{} // sync closure of subscriptions

	logEvents := app.Logger.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range logEvents {
			ch <- ev
		}
		wg.Done()
	}()

	modEvents := app.Modules.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range modEvents {
			ch <- ev
		}
		wg.Done()
	}()

	wsEvents := app.Workspaces.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range wsEvents {
			ch <- ev
		}
		wg.Done()
	}()

	stateEvents := app.States.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range stateEvents {
			ch <- ev
		}
		wg.Done()
	}()

	runEvents := app.Runs.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range runEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskEvents := app.Tasks.TaskBroker.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range taskEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskGroupEvents := app.Tasks.GroupBroker.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range taskGroupEvents {
			ch <- ev
		}
		wg.Done()
	}()

	// Automatically load workspaces whenever modules are loaded.
	app.Workspaces.LoadWorkspacesUponModuleLoad(app.Modules.Subscribe())

	// Whenever a workspace is added, pull its state
	go func() {
		for event := range app.Workspaces.Subscribe() {
			if event.Type == resource.CreatedEvent {
				_, _ = app.States.CreateReloadTask(event.Payload.ID)
			}
		}
	}()

	// cleanup function to be invoked when program is terminated.
	cleanup := func() {
		// Gracefully shutdown remaining tasks.
		//
		// TODO: check this is the right place to do this, and not *after*
		// shutting down pub/sub below.
		app.Cleanup()

		// Close subscriptions
		app.Logger.Shutdown()
		app.Tasks.TaskBroker.Shutdown()
		app.Tasks.GroupBroker.Shutdown()
		app.Modules.Shutdown()
		app.Workspaces.Shutdown()
		app.Runs.Shutdown()
		app.States.Shutdown()

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
