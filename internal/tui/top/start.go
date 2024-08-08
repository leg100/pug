package top

import (
	"context"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/require"
)

// Start starts the TUI and blocks until the user exits.
func Start(cfg app.Config) error {
	app, err := app.New(cfg)
	if err != nil {
		return err
	}
	defer app.Cleanup()

	m, err := newModel(cfg, app)
	if err != nil {
		return err
	}

	tp := tea.NewProgram(m,
		// Use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)

	subscribe := setupSubscriptions(app, tp)
	defer subscribe()

	// Blocks until user quits
	_, err = tp.Run()
	return err
}

// StartTest starts the TUI and returns a test model for testing purposes.
func StartTest(t *testing.T, cfg app.Config, width, height int) *teatest.TestModel {
	app, err := app.New(cfg)
	if err != nil {
		return nil
	}
	t.Cleanup(app.Cleanup)

	m, err := newModel(cfg, app)
	require.NoError(t, err)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(width, height))

	unsub := setupSubscriptions(app, tm)
	t.Cleanup(unsub)

	t.Cleanup(func() {
		tm.Quit()
	})
	return tm
}

type sendable interface {
	Send(tea.Msg)
}

func setupSubscriptions(app *app.App, model sendable) func() {
	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	ch := make(chan tea.Msg)
	wg := sync.WaitGroup{} // sync closure of subscriptions

	ctx, cancel := context.WithCancel(context.Background())

	logEvents := app.Logger.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range logEvents {
			ch <- ev
		}
		wg.Done()
	}()

	modEvents := app.Modules.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range modEvents {
			ch <- ev
		}
		wg.Done()
	}()

	wsEvents := app.Workspaces.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range wsEvents {
			ch <- ev
		}
		wg.Done()
	}()

	stateEvents := app.States.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range stateEvents {
			ch <- ev
		}
		wg.Done()
	}()

	runEvents := app.Runs.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range runEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskEvents := app.Tasks.TaskBroker.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range taskEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskGroupEvents := app.Tasks.GroupBroker.Subscribe(ctx)
	wg.Add(1)
	go func() {
		for ev := range taskGroupEvents {
			ch <- ev
		}
		wg.Done()
	}()

	// Relay events to model in background
	go func() {
		for msg := range ch {
			model.Send(msg)
		}
	}()

	// Automatically load workspaces whenever modules are loaded.
	app.Workspaces.LoadWorkspacesUponModuleLoad(app.Modules.Subscribe(ctx))

	// Whenever a workspace is loaded, pull its state
	go func() {
		for event := range app.Workspaces.Subscribe(ctx) {
			if event.Type == resource.CreatedEvent {
				_, _ = app.States.CreateReloadTask(event.Payload.ID)
			}
		}
	}()

	// cleanup function to be invoked when program is terminated.
	return func() {
		cancel()
		// Wait for relays to finish before closing channel, to avoid sends
		// to a closed channel, which would result in a panic.
		wg.Wait()
		close(ch)
	}
}
