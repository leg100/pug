package app

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *teatest.TestModel {
	t.Helper()

	// Clean up any leftover artefacts from previous tests (previous tests
	// neglect to clean up artefacts if they end with a panic).
	cleanupArtefacts()

	// Cancel context once test finishes.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Clean up artefacts once test finishes
	t.Cleanup(cleanupArtefacts)

	// Setup provider mirror
	setupProviderMirror(t)

	app, m, err := newApp(
		config{
			FirstPage: "modules",
			Program:   "terraform",
			Workdir:   "./testdata",
			MaxTasks:  3,
			loggingOptions: logging.Options{
				Level: "debug",
				AdditionalWriters: []io.Writer{
					&testLogger{t},
				},
			},
		},
	)
	require.NoError(t, err)

	tm := teatest.NewTestModel(
		t,
		m,
		teatest.WithInitialTermSize(100, 50),
	)
	waitfn := app.start(ctx, tm)
	t.Cleanup(func() {
		err := waitfn()
		assert.NoError(t, err, "waiting for running tasks to complete")
	})
	return tm

}

// cleanupArtefacts removes all the detritus that terraform leaves behind.
func cleanupArtefacts() {
	_ = filepath.WalkDir("./testdata/modules", func(path string, d fs.DirEntry, walkerr error) error {
		if walkerr != nil {
			return walkerr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".terraform", ".pug":
				os.RemoveAll(path)
				return fs.SkipDir
			}
		}
		// TODO: consider leaving lock file; it prevents a warning message cropping
		// up.
		if filepath.Base(path) == ".terraform.lock.hcl" {
			os.Remove(path)
		}
		if filepath.Base(path) == "terraform.tfstate" {
			os.Remove(path)
		}
		if strings.HasSuffix(filepath.Base(path), ".backup") {
			os.Remove(path)
		}
		return nil
	})

}

// setupProviderMirror configures a dedicated provider filesystem mirror for for
// a test.
func setupProviderMirror(t *testing.T) {
	t.Helper()

	abs, err := filepath.Abs("../../mirror/mirror.tfrc")
	require.NoError(t, err)

	t.Setenv("TF_CLI_CONFIG_FILE", abs)
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
		return strings.Contains(s, "Proceed with apply (y/N)?")
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
