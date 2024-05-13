package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogs(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Go to logs
	tm.Type("l")

	// Expect at a log message indicating modules have been reloaded
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `reloaded modules added=\[modules/[a-c] modules/[a-c] modules/[a-c]\] removed=\[\]`, s)
	})

	// Drill down into log message
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Expect a table of keys and values
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `level\s+INFO`, s) &&
			matchPattern(t, `message\s+reloaded modules`, s) &&
			matchPattern(t, `added\s+\[modules/[a-c] modules/[a-c] modules/[a-c]\]`, s) &&
			matchPattern(t, `removed\s+\[\]`, s)
	})
}
