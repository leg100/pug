package top

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/app"
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
func makeMakers(cfg app.Config, app *app.App, spinner *spinner.Model) map[tui.Kind]tui.Maker {
	helpers := &tui.Helpers{
		Modules:    app.Modules,
		Workspaces: app.Workspaces,
		Plans:      app.Plans,
		States:     app.States,
		Tasks:      app.Tasks,
		Logger:     app.Logger,
	}

	workspaceListMaker := &workspacetui.ListMaker{
		Workspaces: app.Workspaces,
		Modules:    app.Modules,
		Plans:      app.Plans,
		Helpers:    helpers,
	}
	taskMaker := &tasktui.Maker{
		Plans:   app.Plans,
		Tasks:   app.Tasks,
		Spinner: spinner,
		Helpers: helpers,
		Logger:  app.Logger,
		Program: cfg.Program,
	}
	taskListMaker := tasktui.NewListMaker(
		app.Tasks,
		app.Plans,
		taskMaker,
		helpers,
	)

	makers := map[tui.Kind]tui.Maker{
		tui.ModuleListKind: &moduletui.ListMaker{
			Modules:    app.Modules,
			Workspaces: app.Workspaces,
			Plans:      app.Plans,
			Spinner:    spinner,
			Workdir:    cfg.Workdir,
			Helpers:    helpers,
			Terragrunt: cfg.Terragrunt,
		},
		tui.WorkspaceListKind: workspaceListMaker,
		tui.TaskListKind:      taskListMaker,
		tui.TaskKind:          taskMaker,
		tui.TaskGroupListKind: &tasktui.GroupListMaker{
			Tasks:   app.Tasks,
			Helpers: helpers,
		},
		tui.TaskGroupKind: tasktui.NewGroupMaker(
			app.Tasks,
			app.Plans,
			taskMaker,
			helpers,
		),
		tui.LogListKind: &logs.ListMaker{
			Logger:  app.Logger,
			Helpers: helpers,
		},
		tui.LogKind: &logs.Maker{
			Logger:  app.Logger,
			Helpers: helpers,
		},
		tui.ResourceListKind: &workspacetui.ResourceListMaker{
			Workspaces: app.Workspaces,
			States:     app.States,
			Plans:      app.Plans,
			Spinner:    spinner,
			Helpers:    helpers,
		},
		tui.ResourceKind: &workspacetui.ResourceMaker{
			States:  app.States,
			Plans:   app.Plans,
			Helpers: helpers,
		},
	}
	return makers
}
