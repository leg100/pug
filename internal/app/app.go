// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"fmt"
	"io"
	"os"

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

// Start Pug.
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

	// Construct services and start daemons
	app, err := newApp(cfg, stdout)
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

	// Start the TUI program.
	return top.Start(top.Options{
		Modules:    app.modules,
		Workspaces: app.workspaces,
		Runs:       app.runs,
		States:     app.states,
		Tasks:      app.tasks,
		Logger:     app.logger,
		FirstPage:  cfg.FirstPage,
		Workdir:    app.workdir,
		MaxTasks:   cfg.MaxTasks,
		Debug:      cfg.Debug,
		Program:    cfg.Program,
		Terragrunt: cfg.Terragrunt,
	})
}

type app struct {
	logger  *logging.Logger
	workdir internal.Workdir
	cleanup func()

	modules    *module.Service
	workspaces *workspace.Service
	runs       *run.Service
	states     *state.Service
	tasks      *task.Service
}

// newApp starts the application, constructing services, starting daemons and
// subscribing to events. The returned app is used for constructing the TUI and
// relaying events. The app's cleanup function should be called when finished.
func newApp(cfg config, stdout io.Writer) (*app, error) {
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
		Program:    cfg.Program,
		Logger:     logger,
		Workdir:    workdir,
		UserEnvs:   cfg.Envs,
		UserArgs:   cfg.Args,
		Terragrunt: cfg.Terragrunt,
	})
	modules := module.NewService(module.ServiceOptions{
		Tasks:       tasks,
		Workdir:     workdir,
		PluginCache: cfg.PluginCache,
		Logger:      logger,
		Terragrunt:  cfg.Terragrunt,
	})
	workspaces := workspace.NewService(workspace.ServiceOptions{
		Tasks:   tasks,
		Modules: modules,
		Logger:  logger,
	})
	states := state.NewService(state.ServiceOptions{
		Modules:    modules,
		Workspaces: workspaces,
		Tasks:      tasks,
		Logger:     logger,
	})
	runs := run.NewService(run.ServiceOptions{
		Tasks:                   tasks,
		Modules:                 modules,
		Workspaces:              workspaces,
		States:                  states,
		DisableReloadAfterApply: cfg.DisableReloadAfterApply,
		DataDir:                 cfg.DataDir,
		Logger:                  logger,
		Terragrunt:              cfg.Terragrunt,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start daemons
	task.StartEnqueuer(tasks)
	waitTasks := task.StartRunner(ctx, logger, tasks, cfg.MaxTasks)

	// cleanup function to be invoked when app is terminated.
	cleanup := func() {
		// Cancel context
		cancel()

		// Remove all run artefacts (plan files etc,...)
		for _, run := range runs.List(run.ListOptions{}) {
			_ = os.RemoveAll(run.ArtefactsPath)
		}

		// Wait for running tasks to terminate. Canceling the context (above)
		// sends each task a termination signal so each task's process should
		// shut itself down.
		waitTasks()
	}

	return &app{
		modules:    modules,
		workspaces: workspaces,
		runs:       runs,
		tasks:      tasks,
		states:     states,
		cleanup:    cleanup,
		logger:     logger,
		workdir:    workdir,
	}, nil
}
