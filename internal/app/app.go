// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/workspace"
)

func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse configuration from env vars and flags
	cfg, err := parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Setup logging
	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger.Logger)

	// Log some info useful to the user
	slog.Info(fmt.Sprintf("set max tasks: %d", cfg.MaxTasks))

	// Instantiate services
	tasks := task.NewService(ctx, task.ServiceOptions{
		MaxTasks: cfg.MaxTasks,
		Program:  cfg.Program,
	})
	modules := module.NewService(module.ServiceOptions{
		TaskService: tasks,
		Workdir:     workdir,
		PluginCache: cfg.PluginCache,
	})
	workspaces := workspace.NewService(ctx, workspace.ServiceOptions{
		TaskService:   tasks,
		ModuleService: modules,
	})
	runs := run.NewService(run.ServiceOptions{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
	})

	// Search directory for modules
	if err := modules.Reload(); err != nil {
		return fmt.Errorf("searching for modules: %w", err)
	}

	// Construct TUI programme.
	model, err := tui.New(tui.Options{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
		RunService:       runs,
		Logger:           logger,
		FirstPage:        cfg.FirstPage,
		Workdir:          workdir,
		MaxTasks:         cfg.MaxTasks,
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

	// Relay resource events to TUI.
	logEvents, _ := logger.Subscribe(ctx)
	go func() {
		for ev := range logEvents {
			p.Send(ev)
		}
	}()
	modEvents, _ := modules.Subscribe(ctx)
	go func() {
		for ev := range modEvents {
			p.Send(ev)
		}
	}()
	wsEvents, _ := workspaces.Subscribe(ctx)
	go func() {
		for ev := range wsEvents {
			p.Send(ev)
		}
	}()
	runEvents, _ := runs.Subscribe(ctx)
	go func() {
		for ev := range runEvents {
			p.Send(ev)
		}
	}()
	taskEvents, _ := tasks.Subscribe(ctx)
	go func() {
		for ev := range taskEvents {
			p.Send(ev)
		}
	}()

	// Blocks until user quits
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
