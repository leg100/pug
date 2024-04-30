package run

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/testutils"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_VarsFile tests creating a run with a workspace tfvars file.
func TestRun_VarsFile(t *testing.T) {
	workdir := internal.NewTestWorkdir(t)
	testutils.ChTempDir(t, workdir.String())

	mod := module.New(workdir, "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	// Create a workspace tfvars file for dev
	os.MkdirAll(mod.FullPath(), 0o755)
	_, err = os.Create(filepath.Join(mod.FullPath(), "dev.tfvars"))
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	assert.Equal(t, "dev.tfvars", run.varsFilename)
	assert.Contains(t, run.PlanArgs(), "-var-file=dev.tfvars")
}

func TestRun_MakePugDirectory(t *testing.T) {
	workdir := internal.NewTestWorkdir(t)
	testutils.ChTempDir(t, workdir.String())

	mod := module.New(workdir, "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf("a/b/c/.pug/dev/%s", run.ID)
	assert.DirExists(t, want)
}

func TestRun_PugDirectory(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf(".pug/dev/%s", run.ID)
	assert.Equal(t, want, run.artefactsPath)
}

func TestRun_PlanPath(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	run, err := newRun(mod, ws, CreateOptions{})
	require.NoError(t, err)

	want := fmt.Sprintf(".pug/dev/%s/plan.out", run.ID)
	assert.Equal(t, want, run.PlanPath())
}
