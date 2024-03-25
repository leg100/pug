package task

import "github.com/leg100/pug/internal/resource"

// Multi is a group of one or more tasks.
type Multi []*Task

func CreateMulti(fn Func, ids ...resource.ID) (multi Multi, errs []error) {
	for _, id := range ids {
		task, err := fn(id)
		if err != nil {
			errs = append(errs, err)
		}
		multi = append(multi, task)
	}
	return
}

func (m Multi) Wait() {
	for _, t := range m {
		t.Wait()
	}
}
