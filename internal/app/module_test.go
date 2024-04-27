package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModule(t *testing.T) {
	tm := setup(t)

	// Expect module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")

	// Expect to see successful init message
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})
}

func TestModule_SetCurrentWorkspace(t *testing.T) {
	tm := setup(t)

	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go to module A's page
	tm.Type("m")

	// Expect two workspaces to be listed, and expect default to be the current
	// workspace
	waitFor(t, tm, func(s string) bool {
		defaultIsCurrent := matchPattern(t, `default.*✓`, s)
		return defaultIsCurrent && strings.Contains(s, "dev")
	})

	// Navigate one row down to the dev workspace (default should be the first
	// row because it is lexicographically before dev).
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `dev.*✓`, s)
	})
}
