package top

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/logs"
	moduletui "github.com/leg100/pug/internal/tui/module"
	runtui "github.com/leg100/pug/internal/tui/run"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
)

// makeMakers makes model makers for making models
func makeMakers(opts Options, spinner *spinner.Model) map[tui.Kind]tui.Maker {
	workspaceListMaker := &workspacetui.ListMaker{
		WorkspaceService: opts.WorkspaceService,
		ModuleService:    opts.ModuleService,
		RunService:       opts.RunService,
	}
	runListMaker := &runtui.ListMaker{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
		RunService:       opts.RunService,
		TaskService:      opts.TaskService,
	}
	taskListMaker := &tasktui.ListMaker{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
		TaskService:      opts.TaskService,
		MaxTasks:         opts.MaxTasks,
	}

	breadcrumbs := &tui.Breadcrumbs{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
	}

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			ModuleService:    opts.ModuleService,
			WorkspaceService: opts.WorkspaceService,
			RunService:       opts.RunService,
			Spinner:          spinner,
			Workdir:          opts.Workdir,
		},
		tui.ModuleKind: &moduletui.Maker{
			ModuleService:      opts.ModuleService,
			WorkspaceService:   opts.WorkspaceService,
			RunService:         opts.RunService,
			WorkspaceListMaker: workspaceListMaker,
			RunListMaker:       runListMaker,
			TaskListMaker:      taskListMaker,
			Breadcrumbs:        breadcrumbs,
		},
		tui.WorkspaceListKind: workspaceListMaker,
		tui.WorkspaceKind: &workspacetui.Maker{
			WorkspaceService: opts.WorkspaceService,
			StateService:     opts.StateService,
			RunService:       opts.RunService,
			TaskService:      opts.TaskService,
			RunListMaker:     runListMaker,
			TaskListMaker:    taskListMaker,
			Spinner:          spinner,
			Breadcrumbs:      breadcrumbs,
		},
		tui.RunListKind: runListMaker,
		tui.RunKind: &runtui.Maker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
			Spinner:     spinner,
			Breadcrumbs: breadcrumbs,
		},
		tui.TaskListKind: taskListMaker,
		tui.TaskKind: &tasktui.Maker{
			TaskService: opts.TaskService,
			Breadcrumbs: breadcrumbs,
		},
		tui.LogsKind: &logs.Maker{
			Logger: opts.Logger,
		},
	}
	return makers
}
