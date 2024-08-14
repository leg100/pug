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
	mod := module.New(internal.NewTestWorkdir(t), module.Options{Path: "a/b/c"})
	ws, err := New(mod, "dev")
	require.NoError(t, err)

	assert.Equal(t, "TF_WORKSPACE=dev", ws.TerraformEnv())
}

func TestWorkspace_VarsFile(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), module.Options{Path: "a/b/c"})
	ws, err := New(mod, "dev")
	require.NoError(t, err)

	// Create a workspace tfvars file for dev
	os.MkdirAll(mod.FullPath(), 0o755)
	_, err = os.Create(filepath.Join(mod.FullPath(), "dev.tfvars"))
	require.NoError(t, err)

	got, ok := ws.VarsFile()
	require.True(t, ok)
	assert.Equal(t, "dev.tfvars", got)
}
