package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func TestTaskList_Split(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Go to task list
	tm.Type("t")

	// Expect tasks that are automatically triggered when a module is loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-5 of 5") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/a.*workspace list.*exited`, s) &&
			matchPattern(t, `modules/a.*default.*state pull.*exited`, s)
	})

	// Shrink the split until only 3 tasks are visible. By default, the split
	// view shows 12 rows of tasks. Therefore, the pane needs to be decreased in
	// height 9 times.
	tm.Type(strings.Repeat("-", 9))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-3 of 5")
	})

	// Increase the split until all 5 tasks are visible. That means the split
	// needs to be increased 2 times.
	tm.Type(strings.Repeat("+", 2))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-5 of 5")
	})

}

func TestTask_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
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
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
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
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
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
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
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
		return matchPattern(t, "TaskGroup.*retry.*3/3", s)
	})
}
