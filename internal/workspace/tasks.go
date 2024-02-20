package workspace

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// tasks is a cache of workspace tasks
type tasks struct {
	cache map[ID]task.Categories
}

func (t *tasks) HandleEvent(event resource.Event[*task.Task]) {
}
