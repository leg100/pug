package app

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList_SetCurrentWorkspace(t *testing.T) {
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

	// Go to global workspaces page
	tm.Type("W")

	// Expect two workspaces to be listed, and expect default to be the current
	// workspace
	waitFor(t, tm, func(s string) bool {
		defaultIsCurrent, err := regexp.MatchString(`default.*✓`, s)
		require.NoError(t, err)
		return defaultIsCurrent && strings.Contains(s, "dev")
	})

	// Navigate one row down to the dev workspace (default should be the first
	// row because it is lexicographically before dev).
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Make dev the current workspace
	tm.Type("C")

	// Expect dev to be the new current workspace
	waitFor(t, tm, func(s string) bool {
		devIsCurrent, err := regexp.MatchString(`dev.*✓`, s)
		require.NoError(t, err)
		return devIsCurrent
	})
}
