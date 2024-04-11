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
	helpers := &tui.Helpers{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
		RunService:       opts.RunService,
		StateService:     opts.StateService,
		Logger:           opts.Logger,
	}

	workspaceListMaker := &workspacetui.ListMaker{
		WorkspaceService: opts.WorkspaceService,
		ModuleService:    opts.ModuleService,
		RunService:       opts.RunService,
		Helpers:          helpers,
	}
	runListMaker := &runtui.ListMaker{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
		RunService:       opts.RunService,
		TaskService:      opts.TaskService,
		Helpers:          helpers,
	}
	taskListMaker := &tasktui.ListMaker{
		ModuleService:    opts.ModuleService,
		WorkspaceService: opts.WorkspaceService,
		TaskService:      opts.TaskService,
		MaxTasks:         opts.MaxTasks,
		Helpers:          helpers,
	}

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			ModuleService:    opts.ModuleService,
			WorkspaceService: opts.WorkspaceService,
			RunService:       opts.RunService,
			Spinner:          spinner,
			Workdir:          opts.Workdir.PrettyString(),
			Helpers:          helpers,
		},
		tui.ModuleKind: &moduletui.Maker{
			ModuleService:      opts.ModuleService,
			WorkspaceService:   opts.WorkspaceService,
			RunService:         opts.RunService,
			WorkspaceListMaker: workspaceListMaker,
			RunListMaker:       runListMaker,
			TaskListMaker:      taskListMaker,
			Helpers:            helpers,
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
			Helpers:          helpers,
		},
		tui.RunListKind: runListMaker,
		tui.RunKind: &runtui.Maker{
			RunService:  opts.RunService,
			TaskService: opts.TaskService,
			Spinner:     spinner,
			Helpers:     helpers,
		},
		tui.TaskListKind: taskListMaker,
		tui.TaskKind: &tasktui.Maker{
			TaskService: opts.TaskService,
			Helpers:     helpers,
		},
		tui.LogsKind: &logs.Maker{
			Logger: opts.Logger,
		},
	}
	return makers
}
