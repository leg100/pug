package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_TerraformEnv(t *testing.T) {
	ws, err := New(&module.Module{}, "dev")
	require.NoError(t, err)

	assert.Equal(t, "TF_WORKSPACE=dev", ws.TerraformEnv())
}

func TestWorkspace_VarsFile(t *testing.T) {
	workdir := internal.NewTestWorkdir(t)
	mod := module.NewTestModule(t, module.Options{Path: "a/b/c"})
	ws, err := New(mod, "dev")
	require.NoError(t, err)

	// Create a workspace tfvars file for dev
	path := workdir.Join(mod.Path, "dev.tfvars")
	os.MkdirAll(filepath.Dir(path), 0o755)
	_, err = os.Create(path)
	require.NoError(t, err)

	got, ok := ws.VarsFile(workdir)
	require.True(t, ok)
	assert.Equal(t, "dev.tfvars", got)
}
