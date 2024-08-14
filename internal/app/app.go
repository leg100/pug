// package app is the main entrypoint into the application, responsible for
// configuring and starting the application, services, dependency injection,
// etc.
package app

import (
	"context"
	"os"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type App struct {
	Logger  *logging.Logger
	Cleanup func()

	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	States     *state.Service
	Tasks      *task.Service
}

// New starts the application, constructing services, starting daemons and
// subscribing to events. The returned app is used for constructing the TUI and
// relaying events. The app's cleanup function should be called when finished.
func New(cfg Config) (*App, error) {
	// Setup logging
	logger := logging.NewLogger(cfg.Logging)

	// Log some info useful to the user
	logger.Info("loaded config",
		"log_level", cfg.Logging.Level,
		"max_tasks", cfg.MaxTasks,
		"plugin_cache", cfg.PluginCache,
		"program", cfg.Program,
		"work_dir", cfg.Workdir,
	)

	// Instantiate services
	tasks := task.NewService(task.ServiceOptions{
		Program:    cfg.Program,
		Logger:     logger,
		Workdir:    cfg.Workdir,
		UserEnvs:   cfg.Envs,
		UserArgs:   cfg.Args,
		Terragrunt: cfg.Terragrunt,
	})
	modules := module.NewService(module.ServiceOptions{
		Tasks:       tasks,
		Workdir:     cfg.Workdir,
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
	plans := plan.NewService(plan.ServiceOptions{
		Tasks:      tasks,
		Modules:    modules,
		Workspaces: workspaces,
		States:     states,
		DataDir:    cfg.DataDir,
		Logger:     logger,
		Terragrunt: cfg.Terragrunt,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start daemons
	task.StartEnqueuer(tasks)
	waitTasks := task.StartRunner(ctx, logger, tasks, cfg.MaxTasks)

	// cleanup function to be invoked when app is terminated.
	cleanup := func() {
		// Cancel context
		cancel()

		// Close subscriptions
		logger.Shutdown()
		tasks.TaskBroker.Shutdown()
		tasks.GroupBroker.Shutdown()
		modules.Shutdown()
		workspaces.Shutdown()
		plans.Shutdown()
		states.Shutdown()

		// Wait for running tasks to terminate. Canceling the context (above)
		// sends each task a termination signal so each task's process should
		// shut itself down.
		waitTasks()

		// Remove all run artefacts (plan files etc,...)
		for _, plan := range plans.List() {
			_ = os.RemoveAll(plan.ArtefactsPath)
		}
	}

	return &App{
		Modules:    modules,
		Workspaces: workspaces,
		Plans:      plans,
		Tasks:      tasks,
		States:     states,
		Cleanup:    cleanup,
		Logger:     logger,
	}, nil
}
