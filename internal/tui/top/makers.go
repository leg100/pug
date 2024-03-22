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
		RunService:  opts.RunService,
		TaskService: opts.TaskService,
	}
	taskListMaker := &tasktui.ListMaker{
		TaskService: opts.TaskService,
		MaxTasks:    opts.MaxTasks,
	}

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			ModuleService: opts.ModuleService,
			RunService:    opts.RunService,
			Spinner:       spinner,
			Workdir:       opts.Workdir,
		},
		tui.ModuleKind: &moduletui.Maker{
			ModuleService:      opts.ModuleService,
			WorkspaceService:   opts.WorkspaceService,
			RunService:         opts.RunService,
			WorkspaceListMaker: workspaceListMaker,
			RunListMaker:       runListMaker,
			TaskListMaker:      taskListMaker,
		},
		tui.WorkspaceListKind: &workspacetui.ListMaker{
			WorkspaceService: opts.WorkspaceService,
			ModuleService:    opts.ModuleService,
			RunService:       opts.RunService,
		},
		tui.WorkspaceKind: &workspacetui.Maker{
			WorkspaceService: opts.WorkspaceService,
			RunService:       opts.RunService,
			TaskService:      opts.TaskService,
			RunListMaker:     runListMaker,
			TaskListMaker:    taskListMaker,
		},
		tui.RunListKind: &runtui.ListMaker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
		},
		tui.RunKind: &runtui.Maker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
			Spinner:     spinner,
		},
		tui.TaskListKind: &tasktui.ListMaker{
			TaskService: opts.TaskService,
			MaxTasks:    opts.MaxTasks,
		},
		tui.TaskKind: &tasktui.Maker{
			TaskService: opts.TaskService,
		},
		tui.LogsKind: &logs.Maker{
			Logger: opts.Logger,
		},
	}
	return makers
}
