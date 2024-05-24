package tui

type TabTitle string

const (
	TasksTab     TabTitle = "tasks"
	RunsTab      TabTitle = "runs"
	ResourcesTab TabTitle = "resources"
	PlanTab      TabTitle = "plan"
	ApplyTab     TabTitle = "apply"
)

type SetActiveTabMsg TabTitle
