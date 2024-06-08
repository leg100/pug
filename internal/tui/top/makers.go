package top

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/logs"
	moduletui "github.com/leg100/pug/internal/tui/module"
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
		TaskService:      opts.TaskService,
		Logger:           opts.Logger,
	}

	workspaceListMaker := &workspacetui.ListMaker{
		WorkspaceService: opts.WorkspaceService,
		ModuleService:    opts.ModuleService,
		RunService:       opts.RunService,
		Helpers:          helpers,
	}
	taskMaker := &tasktui.Maker{
		MakerID:     tasktui.TaskMakerID,
		RunService:  opts.RunService,
		TaskService: opts.TaskService,
		Helpers:     helpers,
		Logger:      opts.Logger,
	}
	taskListMaker := &tasktui.ListMaker{
		RunService:  opts.RunService,
		TaskService: opts.TaskService,
		Helpers:     helpers,
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
		tui.WorkspaceListKind: workspaceListMaker,
		tui.TaskListKind:      taskListMaker,
		tui.TaskKind:          taskMaker,
		tui.TaskGroupListKind: &tasktui.GroupListMaker{
			TaskService: opts.TaskService,
			Helpers:     helpers,
		},
		tui.TaskGroupKind: &tasktui.GroupMaker{
			RunService:    opts.RunService,
			TaskService:   opts.TaskService,
			TaskListMaker: taskListMaker,
			Helpers:       helpers,
		},
		tui.LogListKind: &logs.ListMaker{
			Logger: opts.Logger,
		},
		tui.LogKind: &logs.Maker{
			Logger: opts.Logger,
		},
	}
	return makers
}
