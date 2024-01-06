package task

import "github.com/leg100/pug/internal/resource"

// Multi is a group of tasks, each one carrying out the same underyling task but
// carried out on different resources, e.g. terraform init on multiple modules.
type Multi []*Task

// Func is a function that creates a task.
type Func func(resource.ID) (*Task, error)

// CreateMulti invokes the task creating func for each resource id. If the task
// func fails and returns an error, the error is added to errs.
func CreateMulti(fn Func, ids ...resource.ID) (multi Multi, errs []error) {
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

func (m Multi) Wait() {
	for _, t := range m {
		t.Wait()
	}
}
