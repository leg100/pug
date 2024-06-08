package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModuleList(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

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
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "TaskGroup{init}") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s)
	})

	// Go back to module listing and format all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.Type("f")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "TaskGroup{format}") &&
			matchPattern(t, `modules/a.*fmt.*exited`, s) &&
			matchPattern(t, `modules/b.*fmt.*exited`, s) &&
			matchPattern(t, `modules/c.*fmt.*exited`, s)
	})

	// Go back to module listing and validate all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.Type("v")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "TaskGroup{validate}") &&
			matchPattern(t, `modules/a.*validate.*exited`, s) &&
			matchPattern(t, `modules/b.*validate.*exited`, s) &&
			matchPattern(t, `modules/c.*validate.*exited`, s)
	})

	// Go back to module listing, and create plan on the current workspace of
	// each module.
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.Type("p")
	// Expect three plan tasks to be created and to reach planned state.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "TaskGroup{plan}") &&
			matchPattern(t, `modules/a.*default.*planned`, s) &&
			matchPattern(t, `modules/b.*default.*planned`, s) &&
			matchPattern(t, `modules/c.*default.*planned`, s)
	})
}

func TestModuleList_Reload(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect all three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `modules/a`) &&
			strings.Contains(s, `modules/b`) &&
			strings.Contains(s, `modules/c`)
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "completed init tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)")
	})

	// Go back to module listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Each module should now have its current workspace set to default.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s) &&
			matchPattern(t, "modules/b.*default", s) &&
			matchPattern(t, "modules/c.*default", s)
	})

	// Reload workspaces for each and every module
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
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
	t.Parallel()

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
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `completed init tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)`)
	})

	// Go back to module listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Each module should now have its current workspace set to default.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s) &&
			matchPattern(t, "modules/b.*default", s) &&
			matchPattern(t, "modules/c.*default", s)
	})

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	// Create a run on two modules, but not on the last one ("/modules/c")
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Runs(all)[2]") &&
			matchPattern(t, "modules/a.*default.*planned", s) &&
			matchPattern(t, "modules/b.*default.*planned", s)
	})

}

// TestModuleList_ApplyCurrentRun demonstrates a user selecting
// multiple modules and then attempting to apply the latest run on each. Pug
// should de-select those selections which have no current run in a planned
// state.
func TestModuleList_ApplyCurrentRun(t *testing.T) {
	t.Parallel()

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
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, `completed init tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)`)
	})

	// Go back to module listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Wait for each module to be initialized, and to have its current workspace
	// set (should be "default")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s) &&
			matchPattern(t, "modules/b.*default", s) &&
			matchPattern(t, "modules/c.*default", s)
	})

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	// Attempt to apply initialized module but has no plan
	tm.Type("a")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*module does not have a current run", s)
	})

	// Attempt to apply all modules but none have a plan
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Expect error message
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Error:.*no rows are applicable to the given action", s)
	})

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	// Create a plan on two modules, but not on the last one ("c")
	tm.Type("s")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Type("s")
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Runs(all)[2]") &&
			matchPattern(t, "modules/a.*default.*planned", s) &&
			matchPattern(t, "modules/b.*default.*planned", s)
	})

	// Go back to module listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

	// Attempt to apply all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")

	// Expect one module ("C") to have been de-selected, and the apply to be
	// invoked only on two modules ("A", and "B")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 2 runs? (y/N):")
	})
	tm.Type("y")

	// Expect two modules to be applied
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Runs(all)[2]") &&
			matchPattern(t, "modules/a.*default.*applied", s) &&
			matchPattern(t, "modules/b.*default.*applied", s)
	})
}

func TestModuleList_Filter(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/module_list")

	// Expect title to show total of 3 modules, and to list the three modules
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Modules(all)[3]") &&
			strings.Contains(s, `modules/a`) &&
			strings.Contains(s, `modules/b`) &&
			strings.Contains(s, `modules/c`)
	})

	// Focus filter widget
	tm.Type("/")

	// Expect filter prompt
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Filter:")
	})

	// Filter to only show modules/a
	tm.Type("modules/a")

	// Expect title to show 1 module filtered out of a total 3.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Modules(all)[1/3]")
	})
}
