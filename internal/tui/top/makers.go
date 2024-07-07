package top

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/logs"
	moduletui "github.com/leg100/pug/internal/tui/module"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
)

// updateableMaker is a dynamically configurable maker.
type updateableMaker interface {
	Update(tea.Msg) tea.Cmd
}

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
		RunService:  opts.RunService,
		TaskService: opts.TaskService,
		Spinner:     spinner,
		Helpers:     helpers,
		Logger:      opts.Logger,
		Program:     opts.Program,
	}
	taskListMaker := tasktui.NewListMaker(
		opts.TaskService,
		opts.RunService,
		taskMaker,
		helpers,
	)

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
		tui.TaskGroupKind: tasktui.NewGroupMaker(
			opts.TaskService,
			opts.RunService,
			taskMaker,
			helpers,
		),
		tui.LogListKind: &logs.ListMaker{
			Logger: opts.Logger,
		},
		tui.LogKind: &logs.Maker{
			Logger: opts.Logger,
		},
		tui.ResourceListKind: &workspacetui.ResourceListMaker{
			WorkspaceService: opts.WorkspaceService,
			StateService:     opts.StateService,
			RunService:       opts.RunService,
			Spinner:          spinner,
			Helpers:          helpers,
		},
		tui.ResourceKind: &workspacetui.ResourceMaker{
			Helpers: helpers,
		},
	}
	return makers
}
