package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHistory(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/single_module")

	// Show task list
	tm.Type("t")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks")
	})

	// Try go back but get error
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "already at first page")
	})

	// Show logs
	tm.Type("l")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "logs")
	})

	// Go back, expect task list
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks")
	})

	// Try go back but get error
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "already at first page")
	})

	// Focus explorer
	tm.Type("0")

	// Start init task
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 󰠱 modules/a")
	})

	// Go back, expect task list
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks")
	})
}
