package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModuleList(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Expect three modules to be listed
	want := []string{
		"modules/a",
		"modules/b",
		"modules/c",
	}
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
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
	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Reload modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Expect message to inform user that reload has finished and no modules
	// have been added nor removed.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "reloaded modules: added 0; removed 0")
	})
}

func TestModuleList_ReloadWorkspaces(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that three modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Wait for each module to be initialized, and to have its current workspace
	// set (should be "default")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "default") == 3
	})

	// Reload workspaces for current module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	// Expect message to inform user that reload has finished.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed reload-workspace task successfully")
	})

	// Reload workspaces for each and every module
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlW})

	// Expect message to inform user that all three reloads have completed
	// successfully.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed reload-workspace tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)")
	})
}

// TestModuleList_CreateRun demonstrates a user selecting multiple modules and
// then attempting to create a run on each. Pug should de-select those
// selections which are not initialized / have no current workspace.
func TestModuleList_CreateRun(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Attempt to create run on uninitialized module
	tm.Type("p")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*module does not have a current workspace", s)
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlA,
	})
	tm.Type("i")

	// Wait for each module to be initialized, and to have its current workspace
	// set (should be "default")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "default") == 3
	})

	// Create a run on two modules, but not on the last one ("c")
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "planned") == 2
	})

}

// TestModuleList_ApplyCurrentRun demonstrates a user selecting
// multiple modules and then attempting to apply the latest run on each. Pug
// should de-select those selections which have no current run in a planned
// state.
func TestModuleList_ApplyCurrentRun(t *testing.T) {
	tm := setup(t, "./testdata/module_list")

	// Expect message to inform user that modules have been loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "loaded 3 modules")
	})

	// Attempt to apply uninitialized module
	tm.Type("a")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*module does not have a current run", s)
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlA,
	})
	tm.Type("i")

	// Wait for each module to be initialized, and to have its current workspace
	// set (should be "default")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "default") == 3
	})

	// Attempt to apply initialized module but has no plan
	tm.Type("a")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*module does not have a current run", s)
	})

	// Attempt to apply multiple modules but none have a plan
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("a")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*no rows are applicable to the given action", s)
	})

	// Create a plan on two modules, but not on the last one ("c")
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.Type("s")
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "planned") == 2
	})

	// Attempt to apply all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Expect one module ("C") to have been de-selected, and the apply to be
	// invoked only on two modules ("A", and "B")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 2 runs (y/N)?")
	})
	tm.Type("y")

	// Expect two modules to be applied
	waitFor(t, tm, func(s string) bool {
		return strings.Count(s, "applied") == 2
	})
}
