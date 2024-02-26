package module

import "github.com/leg100/pug/internal/task"

type tasks struct {
	pending  []*task.Task
	active   []*task.Task
	finished []*task.Task
}
