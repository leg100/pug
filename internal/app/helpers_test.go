package app

import (
	"fmt"
	"io"
	"os"

	cp "github.com/otiai10/copy"

	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal/logging"
	"github.com/stretchr/testify/require"
)

const mirrorConfigPath = "../../mirror/mirror.tfrc"

type configOption func(cfg *config)

func withTerragrunt() configOption {
	return func(cfg *config) {
		cfg.Program = "terragrunt"
		// Use terraform to ensure the provider mirror is valid (terraform and
		// tofu use different registry hosts, which forms part of the provider
		// mirror path).
		// And disable auto-init otherwise multiple init's may be running
		// concurrently, which causes issues when writing the provider lock
		// file.
		cfg.Args = []string{"--terragrunt-tfpath", "terraform", "--terragrunt-no-auto-init"}
		cfg.Terragrunt = true
	}
}

func setup(t *testing.T, workdir string, opts ...configOption) *testModel {
	t.Helper()

	if _, err := os.Stat(mirrorConfigPath); err != nil {
		t.Skip("integration tests require mirror to be setup with ./hacks/setup_mirror.sh")
	}

	// Get absolute path to terraform cli config. The config sets up terraform
	// to use a provider filesystem mirror, which ensures tests avoid any
	// network roundtrips to retrieve or query providers.
	mirrorConfigPath, err := filepath.Abs(mirrorConfigPath)
	require.NoError(t, err)
	// Copy workdir to a dedicated directory for this test, to ensure any
	// artefacts created in workdir are done so in isolation from other
	// tests that are run in parallel, and to ensure artefacts don't persist to
	// future invocations of this test.
	target := t.TempDir()
	err = cp.Copy(workdir, target)
	require.NoError(t, err)
	workdir = target

	cfg := config{
		FirstPage: "modules",
		Program:   "terraform",
		WorkDir:   workdir,
		MaxTasks:  3,
		DataDir:   t.TempDir(),
		Debug:     true,
		Envs:      []string{fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", mirrorConfigPath)},
		loggingOptions: logging.Options{
			Level: "debug",
			AdditionalWriters: []io.Writer{
				&testLogger{t},
			},
		},
	}
	for _, fn := range opts {
		fn(&cfg)
	}

	app, err := startApp(
		cfg,
		io.Discard,
	)
	require.NoError(t, err)
	t.Cleanup(app.cleanup)

	tm := teatest.NewTestModel(
		t,
		app.model,
		teatest.WithInitialTermSize(120, 50),
	)
	t.Cleanup(func() {
		tm.Quit()
	})

	// Relay events to TUI
	app.relay(tm)

	return &testModel{
		TestModel: tm,
		workdir:   workdir,
	}
}

type testModel struct {
	*teatest.TestModel

	workdir string
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

func waitFor(t *testing.T, tm *testModel, cond func(s string) bool) {
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

func matchPattern(t *testing.T, pattern string, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	require.NoError(t, err)
	return matched
}
