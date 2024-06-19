package tui

import (
	"testing"

	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

func Test_TaskWorkspace(t *testing.T) {
	h := &Helpers{Logger: logging.Discard}

	mod := &module.Module{
		Common: resource.New(resource.Module, resource.GlobalResource),
	}
	task := &task.Task{Common: resource.New(resource.Task, mod)}

	var got resource.Resource = h.TaskWorkspace(task)
	if got != nil {
		t.Fatal("expected TaskWorkspace() to return nil")
	}
}
