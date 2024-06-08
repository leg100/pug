package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/testutils"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_VarsFile tests creating a run with a workspace tfvars file.
func TestRun_VarsFile(t *testing.T) {
	f, mod, ws := setupTest(t)

	// Create a workspace tfvars file for dev
	os.MkdirAll(mod.FullPath(), 0o755)
	_, err := os.Create(filepath.Join(mod.FullPath(), "dev.tfvars"))
	require.NoError(t, err)

	run, err := f.newRun(ws.ID, CreateOptions{})
	require.NoError(t, err)

	assert.Contains(t, run.runArgs, "-var-file=dev.tfvars")
}

func TestRun_MakeArtefactsPath(t *testing.T) {
	f, _, ws := setupTest(t)

	run, err := f.newRun(ws.ID, CreateOptions{})
	require.NoError(t, err)

	assert.DirExists(t, run.ArtefactsPath)
}

func setupTest(t *testing.T) (*factory, *module.Module, *workspace.Workspace) {
	workdir := internal.NewTestWorkdir(t)
	testutils.ChTempDir(t, workdir.String())

	mod := module.New(workdir, "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)
	factory := factory{
		modules:    &fakeModuleGetter{module: mod},
		workspaces: &fakeWorkspaceGetter{ws: ws},
		dataDir:    t.TempDir(),
	}
	return &factory, mod, ws
}

type fakeModuleGetter struct {
	module *module.Module
}

func (f *fakeModuleGetter) Get(resource.ID) (*module.Module, error) {
	return f.module, nil
}

type fakeWorkspaceGetter struct {
	ws *workspace.Workspace
}

func (f *fakeWorkspaceGetter) Get(resource.ID) (*workspace.Workspace, error) {
	return f.ws, nil
}
