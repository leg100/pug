package app

import (
	"io"

	cp "github.com/otiai10/copy"

	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/require"
)

type setupOption func(*setupOptions)

type setupOptions struct {
	keepState bool
}

func keepState() setupOption {
	return func(opts *setupOptions) {
		opts.keepState = true
	}
}

func setup(t *testing.T, workdir string, sopts ...setupOption) *teatest.TestModel {
	t.Helper()

	var opts setupOptions
	for _, fn := range sopts {
		fn(&opts)
	}

	// Copy workdir to a dedicated directory for this test, to ensure any
	// artefacts created in workdir are done so in isolation from other
	// tests that are run in parallel, and to ensure artefacts don't persist to
	// future invocations of this test.
	target := t.TempDir()
	err := cp.Copy(workdir, target)
	require.NoError(t, err)
	workdir = target

	// Get absolute path to terraform cli config. The config sets up terraform
	// to use a provider filesystem mirror, which ensures tests avoid any
	// network roundtrips to retrieve or query providers.
	mirrorConfigPath, err := filepath.Abs("../../mirror/mirror.tfrc")
	require.NoError(t, err)

	app, err := startApp(
		config{
			FirstPage: "modules",
			Program:   "terraform",
			WorkDir:   workdir,
			MaxTasks:  3,
			DataDir:   t.TempDir(),
			Envs:      []string{"TF_CLI_CONFIG_FILE", mirrorConfigPath},
			loggingOptions: logging.Options{
				Level: "debug",
				AdditionalWriters: []io.Writer{
					&testLogger{t},
				},
			},
		},
		io.Discard,
	)
	require.NoError(t, err)
	t.Cleanup(app.cleanup)

	tm := teatest.NewTestModel(
		t,
		app.model,
		teatest.WithInitialTermSize(100, 50),
	)
	t.Cleanup(func() {
		tm.Quit()
	})

	// Relay events to TUI
	app.relay(tm)

	return tm
}

// testLogger relays pug log records to the go test logger
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Write(b []byte) (int, error) {
	l.t.Helper()

	l.t.Log(string(b))
	return len(b), nil
}

func waitFor(t *testing.T, tm *teatest.TestModel, cond func(s string) bool) {
	t.Helper()

	teatest.WaitFor(
		t,
		tm.Output(),
		func(b []byte) bool {
			return cond(string(b))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*10),
	)
}

func initAndApplyModuleA(t *testing.T, tm *teatest.TestModel) {
	// Wait for module to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a")
	})

	// Initialize module
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Terraform has been successfully initialized!")
	})

	// Go to workspaces
	tm.Type("W")

	// Wait for workspace to be loaded
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "default")
	})

	// Create plan for first workspace
	tm.Type("p")

	// User should now be taken to the run page...

	// Expect to see summary of changes
	waitFor(t, tm, func(s string) bool {
		// Remove bold formatting
		s = internal.StripAnsi(s)
		return strings.Contains(s, "Plan: 10 to add, 0 to change, 0 to destroy.")
	})

	// Apply plan and provide confirmation
	tm.Type("a")
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Proceed with apply? (y/N):")
	})
	tm.Type("y")

	// Wait for apply to complete
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "Apply complete! Resources: 10 added, 0 changed, 0 destroyed.")
	})
}

func matchPattern(t *testing.T, pattern string, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	require.NoError(t, err)
	return matched
}
