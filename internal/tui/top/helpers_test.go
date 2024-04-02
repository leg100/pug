package top

import (
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

type fakeTaskService struct {
	tui.TaskService
}

func (f *fakeTaskService) Counter() int { return 0 }

type fakeModuleService struct {
	tui.ModuleService
}

func (f *fakeModuleService) List() []*module.Module {
	return nil
}

type fakeWorkspaceService struct {
	tui.WorkspaceService
}

func (f *fakeWorkspaceService) ReloadAll() (task.Multi, []error) {
	return nil, nil
}
