package internal

// runner is the global task runner that provides a couple of invariants:
// (a) no more than MAX tasks run at any given time
// (b) no more than one 'exclusive' task runs at any given time
type runner struct {
	tasks     chan struct{}
	exclusive chan struct{}

	// TODO: maintain queue of tasks for monitoring purposes, i.e. a global
	// tasks view.
}

func NewRunner(max int) *runner {
	return &runner{
		tasks:     make(chan struct{}, max),
		exclusive: make(chan struct{}, 1),
	}
}

// async, enqueues
func (r *runner) run(spec taskspec) (*task, error) {
	task := newTask(spec)
	// asynchronously kick off task
	go func() {
		// Only start task once there is an available slot, and upon completion
		// unoccupy the slot; if this is an exclusive task then use the
		// exclusive slot.
		if spec.exclusive {
			r.exclusive <- struct{}{}
			defer func() { <-r.exclusive }()
		} else {
			r.tasks <- struct{}{}
			defer func() { <-r.tasks }()
		}
		// start task, and if it starts successfully then wait for it to finish
		if task.start() {
			task.wait()
		}
	}()
	return task, nil
}
