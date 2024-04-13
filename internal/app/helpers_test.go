package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *teatest.TestModel {
	t.Helper()

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
		teatest.WithInitialTermSize(300, 100),
	)
	app.start(ctx, tm)
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
			case ".terraform", ".pug", "terraform.tfstate.d":
				os.RemoveAll(path)
				return fs.SkipDir
			}
		}
		// TODO: consider leaving this; it prevents a warning message cropping
		// up.
		if filepath.Base(path) == ".terraform.lock.hcl" {
			os.Remove(path)
		}
		if strings.HasPrefix(filepath.Base(path), "terraform.tfstate") {
			os.Remove(path)
		}
		return nil
	})

}

const cliConfigTemplate = `
provider_installation {
  filesystem_mirror {
    path = "%s"
  }
}
`

// setupProviderMirror creates and configures a dedicated provider filesystem
// mirror for for a test.
func setupProviderMirror(t *testing.T) {
	t.Helper()

	// Copy providers into a temp dir
	mirrorPath := t.TempDir()
	err := CopyDirectory("../../testdata/providers", mirrorPath)
	require.NoError(t, err)

	// Write terraform CLI config file pointing at filesystem mirror
	config := fmt.Sprintf(cliConfigTemplate, mirrorPath)
	configPath := filepath.Join(t.TempDir(), "_terraform.tfrc")
	err = os.WriteFile(configPath, []byte(config), 0o644)
	require.NoError(t, err)

	// Configure test to use mirror
	t.Setenv("TF_CLI_CONFIG_FILE", configPath)
}

// testLogger relays pug log records to the go test logger
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Write(b []byte) (int, error) {
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
		teatest.WithDuration(time.Second*5),
	)
}
