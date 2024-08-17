package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogs(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Go to logs
	tm.Type("l")

	// Wait for log message to appear telling us modules have been reloaded.
	// Note we only test for the first part of the log message because the test
	// terminal is only of a limited width, which means the message is
	// truncated.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `reloaded modules`)
	})

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only the reloaded modules message, so that we can be sure the
	// cursor is on that message.
	tm.Type("reloaded modules")

	// Exit filter prompt
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

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
