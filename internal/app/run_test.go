package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRuns(t *testing.T) {
	tm := setup(t)

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "✓")
	})

	// Invoke plan
	tm.Type("p")
	// Expect to be taken automatically to the run page
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Run[default](demos/modules/a)")
	})

	// Go to tasks tab (should only need to press tab twice, but some reason
	// test only passes with three presses?)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	// Expect to see finished init task
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "exited")
	})

	// Go to init task page
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})
}
