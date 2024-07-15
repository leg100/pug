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
		Modules:    opts.Modules,
		Workspaces: opts.Workspaces,
		Runs:       opts.Runs,
		States:     opts.States,
		Tasks:      opts.Tasks,
		Logger:     opts.Logger,
	}

	workspaceListMaker := &workspacetui.ListMaker{
		Workspaces: opts.Workspaces,
		Modules:    opts.Modules,
		Runs:       opts.Runs,
		Helpers:    helpers,
	}
	taskMaker := &tasktui.Maker{
		Runs:    opts.Runs,
		Tasks:   opts.Tasks,
		Spinner: spinner,
		Helpers: helpers,
		Logger:  opts.Logger,
		Program: opts.Program,
	}
	taskListMaker := tasktui.NewListMaker(
		opts.Tasks,
		opts.Runs,
		taskMaker,
		helpers,
	)

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			Modules:    opts.Modules,
			Workspaces: opts.Workspaces,
			Runs:       opts.Runs,
			Spinner:    spinner,
			Workdir:    opts.Workdir.PrettyString(),
			Helpers:    helpers,
			Terragrunt: opts.Terragrunt,
		},
		tui.WorkspaceListKind: workspaceListMaker,
		tui.TaskListKind:      taskListMaker,
		tui.TaskKind:          taskMaker,
		tui.TaskGroupListKind: &tasktui.GroupListMaker{
			Tasks:   opts.Tasks,
			Helpers: helpers,
		},
		tui.TaskGroupKind: tasktui.NewGroupMaker(
			opts.Tasks,
			opts.Runs,
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
			Workspaces: opts.Workspaces,
			States:     opts.States,
			Runs:       opts.Runs,
			Spinner:    spinner,
			Helpers:    helpers,
		},
		tui.ResourceKind: &workspacetui.ResourceMaker{
			States:  opts.States,
			Runs:    opts.Runs,
			Helpers: helpers,
		},
	}
	return makers
}
