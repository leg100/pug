// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	toptui "github.com/leg100/pug/internal/tui/top"
	"github.com/leg100/pug/internal/version"
	"github.com/leg100/pug/internal/workspace"
)

func Start(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse configuration from env vars and flags
	cfg, err := parse(args)
	if err != nil {
		return err
	}

	if cfg.version {
		fmt.Println("pug", version.Version)
		return nil
	}

	// Setup logging
	logger := logging.NewLogger(cfg.LogLevel)

	// Log some info useful to the user
	logger.Info("loaded config",
		"log_level", cfg.LogLevel,
		"max_tasks", cfg.MaxTasks,
		"plugin_cache", cfg.PluginCache,
		"program", cfg.Program,
		"work_dir", cfg.Workdir,
	)

	// Instantiate services
	tasks := task.NewService(task.ServiceOptions{
		Program: cfg.Program,
		Logger:  logger,
		Workdir: cfg.Workdir,
	})
	modules := module.NewService(module.ServiceOptions{
		TaskService: tasks,
		Workdir:     cfg.Workdir,
		PluginCache: cfg.PluginCache,
		Logger:      logger,
	})
	workspaces := workspace.NewService(ctx, workspace.ServiceOptions{
		TaskService:   tasks,
		ModuleService: modules,
		Logger:        logger,
	})
	states := state.NewService(ctx, state.ServiceOptions{
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

	// Construct TUI programme.
	model, err := toptui.New(toptui.Options{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
		StateService:     states,
		RunService:       runs,
		Logger:           logger,
		FirstPage:        cfg.FirstPage,
		Workdir:          cfg.Workdir,
		MaxTasks:         cfg.MaxTasks,
		Debug:            cfg.Debug,
	})
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		model,
		// use the full size of the terminal with its "alternate screen buffer"
		tea.WithAltScreen(),
		// Enabling mouse cell motion removes the ability to "blackboard" text
		// with the mouse, which is useful for then copying text into the
		// clipboard. Therefore we've decided to disable it and leave it
		// commented out for posterity.
		//
		// tea.WithMouseCellMotion(),
	)

	// Relay resource events to TUI. Deliberately set up subscriptions *before*
	// any events are triggered, to ensure the TUI receives all messages.
	logEvents := logger.Subscribe(ctx)
	go func() {
		for ev := range logEvents {
			p.Send(ev)
		}
	}()
	modEvents := modules.Subscribe(ctx)
	go func() {
		for ev := range modEvents {
			p.Send(ev)
		}
	}()
	wsEvents := workspaces.Subscribe(ctx)
	go func() {
		for ev := range wsEvents {
			p.Send(ev)
		}
	}()
	stateEvents := states.Subscribe(ctx)
	go func() {
		for ev := range stateEvents {
			p.Send(ev)
		}
	}()
	runEvents := runs.Subscribe(ctx)
	go func() {
		for ev := range runEvents {
			p.Send(ev)
		}
	}()
	taskEvents := tasks.Subscribe(ctx)
	go func() {
		for ev := range taskEvents {
			p.Send(ev)
		}
	}()

	// Start daemons
	go task.StartEnqueuer(ctx, tasks)
	go task.StartRunner(ctx, logger, tasks, cfg.MaxTasks)
	go run.StartScheduler(ctx, runs, workspaces)

	// Blocks until user quits
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
