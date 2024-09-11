package task

import "github.com/leg100/pug/internal/resource"

// Spec is a specification for creating a task.
type Spec struct {
	// ModuleID is the ID of the module the task belongs to. If nil, the task
	// does not belong to a module
	ModuleID *resource.ID
	// WorkspaceID is the ID of the workspace the task belongs to. If nil, the
	// task does not belong to a workspace.
	WorkspaceID *resource.ID
	// Execution specifies the execution of a program.
	Execution Execution
	// AdditionalExecution specifies the execution of another program. The
	// program is only executed if the first program exits successfully.
	AdditionalExecution *Execution
	// Identifier uniquely identifies the type of task.
	Identifier Identifier
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
	// Description assigns an optional description to the task to display to the
	// user, overriding the default of displaying the command.
	Description string
	// Call this function before the task has successfully finished. The
	// returned string sets the task summary, and the error, if non-nil, deems
	// the task to have failed and places the task into an errored state.
	BeforeExited func(*Task) (Summary, error)
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
	// Dependencies specifies that the task respect its module's dependencies.
	// Only makes sense when the task is specified as part of a task group. All
	// specs in the task group must set Dependencies to either nil, or to
	// non-nil.
	Dependencies *Dependencies
	// dependsOn are other tasks that all must successfully exit before the
	// task can be enqueued. If any of the other tasks are canceled or error
	// then the task will be canceled.
	dependsOn []resource.ID
}

// SpecFunc is a function that creates a spec.
type SpecFunc func(resource.ID) (Spec, error)

// Execution specifies the program and arguments to execute
type Execution struct {
	// Program to execute. Defaults to the `program` pug config option.
	Program string
	// Terraform command, including sub commands, e.g. plan, state rm, etc.
	// Ignored if Program is non-empty.
	TerraformCommand []string
	// Args to pass to program.
	Args []string
}

// Dependencies specifies that the task respect its module's dependencies: any
// tasks belonging to the its module's dependencies must have finished
// successfully before this task can be started. This only makes sense in the
// context of a task group, in which multiple tasks are created.
type Dependencies struct {
	ModuleIDs []resource.ID
	// InverseDependencyOrder inverts the order of module dependencies, i.e. if
	// module A depends on module B, then a task specified for module B will
	// only be started once any tasks specified on module A have completed. This
	// is useful when carrying out a `terraform destroy`. All specs in the task
	// group must set InverseDependencyOrder to the same value otherwise an
	// error is raised.
	InverseDependencyOrder bool
}
