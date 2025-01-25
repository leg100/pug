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

	p := tea.NewProgram(m,
		// Use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)

	ch, unsub := setupSubscriptions(app, cfg)
	defer unsub()

	// Relay events to model in background
	go func() {
		for msg := range ch {
			p.Send(msg)
		}
	}()

	// Blocks until user quits
	_, err = p.Run()
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

	ch, unsub := setupSubscriptions(app, cfg)
	t.Cleanup(unsub)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(width, height))

	// Relay events to model in background
	go func() {
		for msg := range ch {
			tm.Send(msg)
		}
	}()

	t.Cleanup(func() {
		tm.Quit()
	})
	return tm
}

func setupSubscriptions(app *app.App, cfg app.Config) (chan tea.Msg, func()) {
	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	ch := make(chan tea.Msg)
	wg := sync.WaitGroup{} // sync closure of subscriptions

	ctx, cancel := context.WithCancel(context.Background())

	{
		sub := app.Logger.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()
	}
	{
		sub := app.Modules.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()
	}
	{
		sub := app.Workspaces.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()

	}
	{
		sub := app.States.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()

	}
	{
		sub := app.Plans.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()

	}
	{
		sub := app.Tasks.TaskBroker.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()

	}
	{
		sub := app.Tasks.GroupBroker.Subscribe(ctx)
		wg.Add(1)
		go func() {
			for ev := range sub {
				ch <- ev
			}
			wg.Done()
		}()
	}
	// Automatically load workspaces whenever modules are loaded.
	{
		sub := app.Modules.Subscribe(ctx)
		go app.Workspaces.LoadWorkspacesUponModuleLoad(sub)
	}
	// Automatically load workspaces whenever init is run and workspaces have
	// not yet been loaded.
	{
		sub := app.Tasks.TaskBroker.Subscribe(ctx)
		go app.Workspaces.LoadWorkspacesUponInit(sub)
	}
	// Whenever a workspace is loaded, pull its state
	{
		sub := app.Workspaces.Subscribe(ctx)
		go func() {
			for event := range sub {
				if event.Type == resource.CreatedEvent {
					_, err := app.States.CreateReloadTask(event.Payload.ID)
					if err != nil {
						app.Logger.Error("loading state after loading workspace", "error", err)
					}
				}
			}
		}()
	}
	// Whenever an apply is successful, pull workspace state
	if !cfg.DisableReloadAfterApply {
		sub := app.Tasks.TaskBroker.Subscribe(ctx)
		go app.Plans.ReloadAfterApply(sub)
	}
	// Whenever a plan file is created, create a task to convert it into JSON
	{
		sub := app.Tasks.TaskBroker.Subscribe(ctx)
		go app.Plans.GenerateJSONAfterPlan(sub)
	}
	// cleanup function to be invoked when program is terminated.
	return ch, func() {
		cancel()
		// Wait for relays to finish before closing channel, to avoid sends
		// to a closed channel, which would result in a panic.
		wg.Wait()
		close(ch)
	}
}
