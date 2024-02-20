package module

import "github.com/leg100/pug/internal/task"

const InitTask task.Kind = "init"

type tasks struct {
	pending  []*task.Task
	active   []*task.Task
	finished []*task.Task
}
