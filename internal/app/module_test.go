package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModules(t *testing.T) {
	tm := setup(t)

	// Expect three modules to be listed
	want := []string{
		"modules/a",
		"modules/b",
		"modules/c",
	}
	waitFor(t, tm, func(s string) bool {
		for _, w := range want {
			if !strings.Contains(s, w) {
				return false
			}
		}
		return true
	})

	// Initialize first module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "✓")
	})

	// Go to first module's page
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

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
