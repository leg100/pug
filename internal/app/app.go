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
	"github.com/hashicorp/terraform/internal/logging"
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
	slog.SetDefault(logger)

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
	workspaces := workspace.NewService(workspace.ServiceOptions{
		TaskService:   tasks,
		ModuleService: modules,
	})
	runs := run.NewService(run.ServiceOptions{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
	})

	// Construct TUI programme.
	model, err := tui.New(tui.Options{
		TaskService:      tasks,
		ModuleService:    modules,
		WorkspaceService: workspaces,
		RunService:       runs,
		Workdir:          workdir,
	})
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		model,
		// use the full size of the terminal in its "alternate screen buffer"
		tea.WithAltScreen(),
		// turn on mouse support so we can track the mouse wheel
		tea.WithMouseCellMotion(),
	)

	// Relay resource events to TUI.
	go func() {
		events, _ := logging.Subscribe(ctx)
		for ev := range events {
			p.Send(ev)
		}
	}()
	go func() {
		events, _ := modules.Subscribe(ctx)
		for ev := range events {
			p.Send(ev)
		}
	}()
	go func() {
		events, _ := workspaces.Subscribe(ctx)
		for ev := range events {
			p.Send(ev)
		}
	}()
	go func() {
		events, _ := runs.Subscribe(ctx)
		for ev := range events {
			p.Send(ev)
		}
	}()
	go func() {
		events, _ := tasks.Subscribe(ctx)
		for ev := range events {
			p.Send(ev)
		}
	}()

	// Blocks until user quits
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
