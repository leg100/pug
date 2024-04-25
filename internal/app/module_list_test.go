package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModuleList(t *testing.T) {
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

	// Select all modules and init
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlA,
	})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed init tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)")
	})

	// Select all modules and format
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlA,
	})
	tm.Type("f")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed format tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)")
	})

	// Select all modules and validate
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlA,
	})
	tm.Type("v")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed validate tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)")
	})
}

func TestModuleList_Reload(t *testing.T) {
	tm := setup(t)

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Reload modules
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlR,
	})

	// Expect message to inform user that reload has finished and no modules
	// have been added nor removed.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "reloaded modules: added 0; removed 0")
	})
}
