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

func TestTaskList_Split(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Go to task list
	tm.Type("t")

	// Expect tasks that are automatically triggered when a module is loaded
	waitFor(t, tm, func(s string) bool {
		t.Log(s)
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-4 of 4") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/a.*workspace list.*exited`, s) &&
			matchPattern(t, `modules/a.*default.*state pull.*exited`, s)
	})

	// Shrink the split until only 3 tasks are visible. By default, the split
	// view shows 12 rows of tasks. Therefore, the pane needs to be decreased in
	// height 9 times.
	tm.Type(strings.Repeat("-", 9))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-3 of 4")
	})

	// Increase the split until all 4 tasks are visible. That means the split
	// needs to be increased once.
	tm.Type(strings.Repeat("+", 1))

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Tasks") &&
			strings.Contains(s, "1-4 of 4")
	})

}

func TestTask_SingleApply(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
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

	tm := setupAndInitModuleList(t)

	// Create plan on all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to task group page for plan.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*3/3", s)
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
		return matchPattern(t, "TaskGroup.*apply.*3/3", s) &&
			matchPattern(t, `modules/a.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/b.*default.*\+10~0-0`, s) &&
			matchPattern(t, `modules/c.*default.*\+10~0-0`, s)
	})
}

func TestTask_Retry(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModule(t)

	// Create plan on first module
	tm.Type("p")

	// Expect to be taken to task page for plan.
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
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
		return matchPattern(t, `Task.*plan.*default.*modules/a.*\+10~0-0.*exited`, s) &&
			strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})
}

func TestTask_RetryMultiple(t *testing.T) {
	t.Parallel()

	tm := setupAndInitModuleList(t)

	// Create plan for all modules
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("p")

	// Expect to be taken to task page for plan and wait for it to finish.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*plan.*3/3", s)
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
		return matchPattern(t, "TaskGroup.*plan.*3/3", s)
	})
}

func TestTask_Cancel(t *testing.T) {
	t.Parallel()

	tm := setup(t, "./testdata/cancel_single/")

	// Go to modules listing
	tm.Type("m")

	// Expect single module to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "Task.*init.*modules/a.*exited", s)
	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect single module to be listed, along with its default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s)
	})

	// Stand up http server to receive request from terraform plan
	setupHTTPServer(t, tm.workdir, "a")

	// Invoke plan on module
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

	// Go to modules listing
	tm.Type("m")

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a") &&
			strings.Contains(s, "modules/b") &&
			strings.Contains(s, "modules/c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*init", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	})

	// Go back to modules listing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Expect three modules to be listed, along with their default workspace.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "modules/a.*default", s) &&
			matchPattern(t, "modules/b.*default", s) &&
			matchPattern(t, "modules/c.*default", s)
	})

	// Stand up http server to receive request from terraform plans
	setupHTTPServer(t, tm.workdir, "a", "b", "c")

	// Invoke plans on modules
	tm.Type("p")

	// Wait for plan tasks to enter running state.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*plan.*running`, s) &&
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
