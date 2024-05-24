package task

import "github.com/leg100/pug/internal/resource"

type Group []*Task

// CreateGroup creates a task group, a group of tasks
func CreateGroup(fn Func, ids ...resource.ID) (multi Group, errs []error) {
	for _, id := range ids {
		task, err := fn(id)
		if err != nil {
			errs = append(errs, err)
		} else {
			multi = append(multi, task)
		}
	}
	return
}

func (m Group) Wait() {
	for _, t := range m {
		t.Wait()
	}
}
