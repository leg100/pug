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

	run, err := f.newPlan(ws.ID, CreateOptions{})
	require.NoError(t, err)

	if assert.NotNil(t, run.varsFileArg) {
		assert.Equal(t, *run.varsFileArg, "-var-file=dev.tfvars")
	}
}

func TestRun_MakeArtefactsPath(t *testing.T) {
	f, _, ws := setupTest(t)

	run, err := f.newPlan(ws.ID, CreateOptions{planFile: true})
	require.NoError(t, err)

	assert.DirExists(t, run.ArtefactsPath)
}

func setupTest(t *testing.T) (*factory, *module.Module, *workspace.Workspace) {
	workdir := internal.NewTestWorkdir(t)
	testutils.ChTempDir(t, workdir.String())

	mod := module.New(workdir, module.Options{Path: "a/b/c"})
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)
	factory := factory{
		workspaces: &fakeWorkspaceGetter{ws: ws},
		dataDir:    t.TempDir(),
	}
	return &factory, mod, ws
}

type fakeWorkspaceGetter struct {
	ws *workspace.Workspace
}

func (f *fakeWorkspaceGetter) Get(resource.ID) (*workspace.Workspace, error) {
	return f.ws, nil
}
