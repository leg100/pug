package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
	"github.com/stretchr/testify/require"
)

func TestTask_LongRunningTasks(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/modules_with_long_running_terraform_apply/")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Expect init task group with 3 successful tasks.
	// Each module should now be populated with at least one workspace.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 3/3") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s) &&
			strings.Contains(s, "󰠱 3  3")
	})

	// Go to explorer
	tm.Type("0")

	// Auto-apply all modules
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Auto-apply 3 workspaces? (y/N):")
	})
	tm.Type("y")

	// Apply tasks should complete with a non-zero age
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 3/3") &&
			matchPattern(t, `modules/a.*default.*apply.*exited.*\+1~0-0.*[1-9]s ago`, s) &&
			matchPattern(t, `modules/b.*default.*apply.*exited.*\+1~0-0.*[1-9]s ago`, s) &&
			matchPattern(t, `modules/c.*default.*apply.*exited.*\+1~0-0.*[1-9]s ago`, s)
	})
}

func TestTaskList_Split(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Show task list
	tm.Type("t")

	// Expect tasks that are automatically triggered when a module is loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks") &&
			strings.Contains(s, "1-4 of 4") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/a.*workspace list.*exited`, s) &&
			matchPattern(t, `modules/a.*default.*state pull.*exited`, s)
	})

	// Shrink the split until only 1 task is visible. By default, the split
	// view shows 12 rows of tasks. Therefore, the pane needs to be decreased in
	// height 11 times.
	tm.Type(strings.Repeat("-", 11))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks") &&
			strings.Contains(s, "1-1 of 4")
	})

	// Increase the split until all 4 tasks are visible. That means the split
	// needs to be increased three times.
	tm.Type(strings.Repeat("+", 3))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "tasks") &&
			strings.Contains(s, "1-4 of 4")
	})
}

func TestTask_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0") &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Apply plan and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply plan? (y/N):")
	})
	tm.Type("y")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}

func TestTask_MultipleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Create plan on all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 3/3") &&
			matchPattern(t, `modules/a.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*exited.*\+10~0-0`, s)
	})

	// Select all tasks in group and apply
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply 3 plans? (y/N):")
	})
	tm.Type("y")

	// Expected to be taken to the task group page for apply tasks
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "apply 3/3") &&
			matchPattern(t, `modules/a.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*apply.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*apply.*exited.*\+10~0-0`, s)
	})
}

func TestTask_Retry(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule_Explorer(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0") &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Retry task.
	tm.Type("r")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Retry task? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to new task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "plan 󰠱 modules/a  default") &&
			strings.Contains(s, "exited +10~0-0") &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})
}

func TestTask_RetryMultiple(t *testing.T) {
	t.Parallel()

	tm := setupAndInitMultipleModules(t)

	// Create plan for all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 3/3") &&
			matchPattern(t, `modules/a.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*exited.*\+10~0-0`, s)
	})

	// Retry all plan tasks.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("r")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Retry 3 tasks? (y/N):")
	})
	tm.Type("y")

	// Expect to be taken to task group page for plan and wait for it to finish.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 3/3") &&
			matchPattern(t, `modules/a.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*exited.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*exited.*\+10~0-0`, s)
	})
}

func TestTask_Cancel(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/cancel_single/")

	// Expect single module in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "└ 󰠱 a")
	})

	// Cursor should automatically be on module.
	// Initialize module.
	tm.Type("i")

	// Expect to see successful init message.
	// Tree should now have workspace under module.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!") &&
			strings.Contains(s, "init 󰠱 modules/a") &&
			strings.Contains(s, "└ 󰠱 a") &&
			strings.Contains(s, "└  default")
	})

	// Stand up http server to receive request from terraform plan
	setupHTTPServer(t, tm.workdir, "a")

	// Go back to explorer and invoke plan
	tm.Type("0")
	tm.Type("p")

	// Wait for something that never arrives
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "data.http.forever: Reading...")
	})

	// Cancel plan
	tm.Type("c")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Cancel task? (y/N):")
	})
	tm.Type("y")

	// Wait for footer to report signal sent, and for process to receive signal
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "sent cancel signal to task") &&
			strings.Contains(s, "Interrupt received")
	})
}

func TestTask_CancelMultiple(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/cancel_multiple/")

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	// Expect init task group with 3 successful tasks.
	// Each module should now be populated with at least one workspace.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 3/3") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s) &&
			strings.Contains(s, "󰠱 3  3")
	})

	// Stand up http server to receive request from terraform plans
	setupHTTPServer(t, tm.workdir, "a", "b", "c")

	// Go back to explorer and invoke plan on all modules (they should still be
	// selected)
	tm.Type("0")
	tm.Type("p")

	// Wait for plan tasks to enter running state.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "plan 0/3") &&
			matchPattern(t, `modules/a.*default.*plan.*running`, s) &&
			matchPattern(t, `modules/b.*default.*plan.*running`, s) &&
			matchPattern(t, `modules/c.*default.*plan.*running`, s)
	})

	// Cancel plans
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("c")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Cancel 3 tasks? (y/N):")
	})
	tm.Type("y")

	// Wait for footer to report signals sent, and for processes to receive signal
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "sent cancel signal to 3 tasks")
	})
}

// Stand up http server to receive http request, and write out its URL to a
// tfvars file in each of the given modules.
func setupHTTPServer(t *testing.T, workdir string, mods ...string) {
	ch := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block request
		<-ch
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(ch) })

	// Write out a tfvars files with the url variable set to the url of the http
	// server
	for _, mod := range mods {
		path := filepath.Join(workdir, "modules", mod, "default.tfvars")
		contents := fmt.Sprintf(`url = "%s"`, srv.URL)
		err := os.WriteFile(path, []byte(contents), 0o644)
		require.NoError(t, err)
	}
}
