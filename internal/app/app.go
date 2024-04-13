// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"fmt"
	"io"

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

type app struct {
	modules    *module.Service
	workspaces *workspace.Service
	states     *state.Service
	runs       *run.Service
	tasks      *task.Service
	logger     *logging.Logger
	cfg        config
}

type sender interface {
	Send(tea.Msg)
}

// Start the app.
func Start(stdout, stderr io.Writer, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse configuration from env vars and flags
	cfg, err := parse(stderr, args)
	if err != nil {
		return err
	}

	if cfg.version {
		fmt.Fprintln(stdout, "pug", version.Version)
		return nil
	}

	app, model, err := newApp(cfg)
	if err != nil {
		return err
	}

	// Log some info useful to the user
	app.logger.Info("loaded config",
		"log_level", cfg.LogLevel,
		"max_tasks", cfg.MaxTasks,
		"plugin_cache", cfg.PluginCache,
		"program", cfg.Program,
		"work_dir", cfg.Workdir,
	)

	p := tea.NewProgram(
		model,
		// Use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)

	// Start daemons and relay events.
	app.start(ctx, p)

	// Blocks until user quits
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// newApp constructs an instance of the app and the top-level TUI model.
func newApp(cfg config) (*app, tea.Model, error) {
	// Setup logging
	logger := logging.NewLogger(cfg.LogLevel)

	// Perform any conversions from the flag parsed primitive types to pug
	// defined types.
	workdir, err := internal.NewWorkdir(cfg.Workdir)
	if err != nil {
		return nil, nil, err
	}

	// Instantiate services
	tasks := task.NewService(task.ServiceOptions{
		Program: cfg.Program,
		Logger:  logger,
		Workdir: workdir,
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
	// TODO: separate auto-state pull code
	states := state.NewService(context.Background(), state.ServiceOptions{
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
		return nil, nil, err
	}

	app := &app{
		modules:    modules,
		workspaces: workspaces,
		runs:       runs,
		states:     states,
		tasks:      tasks,
		cfg:        cfg,
		logger:     logger,
	}
	return app, model, nil
}

// start starts the app daemons and relays events to the TUI.
func (a *app) start(ctx context.Context, s sender) {
	// Start daemons
	//
	// TODO: have the daemons tell us when they're doing setting up.
	go task.StartEnqueuer(ctx, a.tasks)
	go task.StartRunner(ctx, a.logger, a.tasks, a.cfg.MaxTasks)
	go run.StartScheduler(ctx, a.runs, a.workspaces)

	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	logEvents := a.logger.Subscribe(ctx)
	go func() {
		for ev := range logEvents {
			s.Send(ev)
		}
	}()
	modEvents := a.modules.Subscribe(ctx)
	go func() {
		for ev := range modEvents {
			s.Send(ev)
		}
	}()
	wsEvents := a.workspaces.Subscribe(ctx)
	go func() {
		for ev := range wsEvents {
			s.Send(ev)
		}
	}()
	stateEvents := a.states.Subscribe(ctx)
	go func() {
		for ev := range stateEvents {
			s.Send(ev)
		}
	}()
	runEvents := a.runs.Subscribe(ctx)
	go func() {
		for ev := range runEvents {
			s.Send(ev)
		}
	}()
	taskEvents := a.tasks.Subscribe(ctx)
	go func() {
		for ev := range taskEvents {
			s.Send(ev)
		}
	}()
}
