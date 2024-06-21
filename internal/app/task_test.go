package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func TestTask_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited.*planned", s) &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Apply plan and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply plan? (y/N):")
	})
	tm.Type("y")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}

func TestTask_MultipleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Create plan on all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to task group page for plan.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*3/3", s)
	})

	// Select all tasks in group and apply
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 3 plans? (y/N):")
	})
	tm.Type("y")

	// Expected to be taken to the task group page for apply tasks
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*apply.*3/3", s) &&
			matchPattern(t, "modules/a.*default.*applied", s) &&
			matchPattern(t, "modules/b.*default.*applied", s) &&
			matchPattern(t, "modules/c.*default.*applied", s)
	})
}

func TestTask_Retry(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited.*planned", s) &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Retry task.
	tm.Type("r")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Retry task? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to new task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, "Task.*plan.*default.*modules/a.*exited.*planned", s) &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})
}

func TestTask_RetryMultiple(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Create plan for all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to task page for plan and wait for it to finish.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*3/3", s)
	})

	// Retry all tasks.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("r")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Retry 3 tasks? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to task page for plan and wait for it to finish.
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		return matchPattern(t, "TaskGroup.*retry.*3/3", s)
	})
}
