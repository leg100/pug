// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/top"
	"github.com/leg100/pug/internal/version"
	"github.com/leg100/pug/internal/workspace"
)

// Start the app.
func Start(stdout, stderr io.Writer, args []string) error {
	// Parse configuration from env vars and flags
	cfg, err := parse(stderr, args)
	if err != nil {
		return err
	}

	if cfg.version {
		fmt.Fprintln(stdout, "pug", version.Version)
		return nil
	}

	// Start daemons and create event subscriptions.
	app, err := startApp(cfg, stdout)
	if err != nil {
		return err
	}
	defer app.cleanup()

	// Log some info useful to the user
	app.logger.Info("loaded config",
		"log_level", cfg.loggingOptions.Level,
		"max_tasks", cfg.MaxTasks,
		"plugin_cache", cfg.PluginCache,
		"program", cfg.Program,
		"work_dir", cfg.WorkDir,
	)

	p := tea.NewProgram(
		app.model,
		// Use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)

	// Relay events to TUI
	app.relay(p)

	// Blocks until user quits
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

type app struct {
	model   tea.Model
	ch      chan tea.Msg
	logger  *logging.Logger
	cleanup func()
}

// startApp starts the application, constructing services, starting daemons and
// subscribing to events. The returned app is used for constructing the TUI and
// relaying events. The app's cleanup function should be called when finished.
func startApp(cfg config, stdout io.Writer) (*app, error) {
	// Setup logging
	logger := logging.NewLogger(cfg.loggingOptions)

	// Perform any conversions from the flag parsed primitive types to pug
	// defined types.
	workdir, err := internal.NewWorkdir(cfg.WorkDir)
	if err != nil {
		return nil, err
	}

	// Instantiate services
	tasks := task.NewService(task.ServiceOptions{
		Program:  cfg.Program,
		Logger:   logger,
		Workdir:  workdir,
		UserEnvs: cfg.Envs,
	})
	modules := module.NewService(module.ServiceOptions{
		TaskService: tasks,
		Workdir:     workdir,
		PluginCache: cfg.PluginCache,
		Logger:      logger,
	})
	workspaces := workspace.NewService(workspace.ServiceOptions{
		TaskService:   tasks,
		ModuleService: modules,
		Logger:        logger,
	})
	states := state.NewService(state.ServiceOptions{
		ModuleService:    modules,
		WorkspaceService: workspaces,
		TaskService:      tasks,
		Logger:           logger,
	})
	runs := run.NewService(run.ServiceOptions{
		TaskService:             tasks,
		ModuleService:           modules,
		WorkspaceService:        workspaces,
		StateService:            states,
		DisableReloadAfterApply: cfg.DisableReloadAfterApply,
		DataDir:                 cfg.DataDir,
		Logger:                  logger,
	})

	// Construct top-level TUI model.
	model, err := top.New(top.Options{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
		StateService:     states,
		RunService:       runs,
		Logger:           logger,
		FirstPage:        cfg.FirstPage,
		Workdir:          workdir,
		MaxTasks:         cfg.MaxTasks,
		Debug:            cfg.Debug,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start daemons
	task.StartEnqueuer(tasks)
	run.StartScheduler(runs, workspaces)
	waitTasks := task.StartRunner(ctx, logger, tasks, cfg.MaxTasks)

	// Automatically load workspaces whenever modules are loaded.
	workspaces.LoadWorkspacesUponModuleLoad(modules)

	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	ch := make(chan tea.Msg)
	wg := sync.WaitGroup{} // sync closure of subscriptions

	logEvents := logger.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range logEvents {
			ch <- ev
		}
		wg.Done()
	}()

	modEvents := modules.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range modEvents {
			ch <- ev
		}
		wg.Done()
	}()

	wsEvents := workspaces.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range wsEvents {
			ch <- ev
		}
		wg.Done()
	}()

	stateEvents := states.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range stateEvents {
			ch <- ev
		}
		wg.Done()
	}()

	runEvents := runs.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range runEvents {
			ch <- ev
		}
		wg.Done()
	}()

	taskEvents := tasks.Subscribe()
	wg.Add(1)
	go func() {
		for ev := range taskEvents {
			ch <- ev
		}
		wg.Done()
	}()

	// cleanup function to be invoked when app is terminated.
	cleanup := func() {
		// Cancel context
		cancel()

		// Close subscriptions
		logger.Shutdown()
		tasks.Shutdown()
		modules.Shutdown()
		workspaces.Shutdown()
		states.Shutdown()
		runs.Shutdown()

		// Wait for relays to finish before closing channel, to avoid sends
		// to a closed channel, which would result in a panic.
		wg.Wait()
		close(ch)

		// Remove all run artefacts (plan files etc,...)
		for _, run := range runs.List(run.ListOptions{}) {
			_ = os.RemoveAll(run.ArtefactsPath())
		}

		// Wait for running tasks to terminate. Canceling the context (above)
		// sends each task a termination signal so each task's process should
		// shut itself down.
		waitTasks()
	}
	return &app{
		model:   model,
		ch:      ch,
		cleanup: cleanup,
		logger:  logger,
	}, nil
}

type tui interface {
	Send(tea.Msg)
}

// relay events to TUI.
func (a *app) relay(s tui) {
	go func() {
		for msg := range a.ch {
			s.Send(msg)
		}
	}()
}
