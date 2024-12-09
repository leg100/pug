package top

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/explorer"
	"github.com/leg100/pug/internal/tui/logs"
	tasktui "github.com/leg100/pug/internal/tui/task"
	workspacetui "github.com/leg100/pug/internal/tui/workspace"
)

// updateableMaker is a dynamically configurable maker.
type updateableMaker interface {
	Update(tea.Msg) tea.Cmd
}

// makeMakers makes model makers for making models
func makeMakers(
	cfg app.Config,
	app *app.App,
	spinner *spinner.Model,
	helpers *tui.Helpers,
) map[tui.Kind]tui.Maker {
	taskMaker := &tasktui.Maker{
		Plans:   app.Plans,
		Tasks:   app.Tasks,
		Spinner: spinner,
		Helpers: helpers,
		Logger:  app.Logger,
		Program: cfg.Program,
	}
	logMaker := &logs.Maker{
		Logger:  app.Logger,
		Helpers: helpers,
	}
	makers := map[tui.Kind]tui.Maker{
		tui.ExplorerKind: &explorer.Maker{
			ModuleService:    app.Modules,
			WorkspaceService: app.Workspaces,
			PlanService:      app.Plans,
			Workdir:          cfg.Workdir,
			Helpers:          helpers,
		},
		tui.TaskListKind: tasktui.NewListMaker(
			app.Tasks,
			app.Plans,
			taskMaker,
			helpers,
		),
		tui.TaskKind: taskMaker,
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
			Logger:        app.Logger,
			Helpers:       helpers,
			LogModelMaker: logMaker,
		},
		tui.LogKind: logMaker,
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
