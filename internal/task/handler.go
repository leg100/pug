package task

// Handler handles a task, providing the necessary info
type Handler interface {
	Command() []string
	JSON() bool
	Blocking() bool
	Immediate() bool
	OnSuccess(t *Task)
}
