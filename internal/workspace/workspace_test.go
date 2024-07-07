package workspace

import (
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
