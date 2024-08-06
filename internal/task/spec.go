package task

import "github.com/leg100/pug/internal/resource"

// Spec is a specification for creating a task.
type Spec struct {
	// Resource that the task belongs to.
	Parent resource.Resource
	// Program command and any sub commands, e.g. plan, state rm, etc.
	Command []string
	// Args to pass to program.
	Args []string
	// Path relative to the pug working directory in which to run the command.
	Path string
	// Environment variables.
	Env []string
	// A blocking task blocks other tasks from running on the module or
	// workspace.
	Blocking bool
	// Globally exclusive task - at most only one such task can be running
	Exclusive bool
	// Set to true to indicate that the task produces JSON output
	JSON bool
	// Skip queue and immediately start task
	Immediate bool
	// Wait blocks until the task has finished
	Wait bool
	// DependsOn are other tasks that all must successfully exit before the
	// task can be enqueued. If any of the other tasks are canceled or error
	// then the task will be canceled.
	DependsOn []resource.ID
	// Description assigns an optional description to the task to display to the
	// user, overriding the default of displaying the command.
	Description string
	// RespectModuleDependencies when true ensures the task respects its
	// module's dependencies. i.e. if module A depends on module B,
	// and a task is specified for both modules then the task for module A is
	// only started once the task for module B has completed. This option
	// only makes sense in the context of a task group, which constructs tasks
	// from multiple specs. All specs must set RespectModuleDependencies to
	// the same value otherwise an error is raised.
	RespectModuleDependencies bool
	// InverseDependencyOrder inverts the order of module dependencies, i.e. if
	// module A depends on module B, then a task specified for module B will
	// only be started once any tasks specified on module A have completed. This
	// is useful when carrying out a `terraform destroy`. This option only takes
	// effect when RespectModuleDependencies is true, and the spec is specified
	// as part of a task group. All specs in the task group must set
	// InverseDependencyOrder to the same value otherwise an error is raised.
	InverseDependencyOrder bool
	// Call this function after the CLI program has successfully finished
	AfterCLISuccess func(*Task) error
	// Call this function after the task has successfully finished
	AfterExited func(*Task)
	// Call this function after the task is enqueued.
	AfterQueued func(*Task)
	// Call this function after the task starts running.
	AfterRunning func(*Task)
	// Call this function after the task fails with an error
	AfterError func(*Task)
	// Call this function after the task is successfully canceled
	AfterCanceled func(*Task)
	// Call this function after the task is successfully created
	AfterCreate func(*Task)
	// Call this function after the task terminates for whatever reason.
	AfterFinish func(*Task)
}

// SpecFunc is a function that creates a spec.
type SpecFunc func(resource.ID) (Spec, error)
