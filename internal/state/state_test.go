package state

import (
	"os"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getResourcesFromFile(t *testing.T) {
	mod := module.New(internal.NewTestWorkdir(t), "a/b/c")
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	f, err := os.Open("./testdata/with_mods/terraform.tfstate.d/dev/terraform.tfstate")
	require.NoError(t, err)
	t.Cleanup(func() {
		f.Close()
	})

	got, err := newState(ws, f)
	require.NoError(t, err)

	assert.Len(t, got.Resources, 17)

	assert.Contains(t, got.Resources, ResourceAddress("random_pet.pet[0]"))
	assert.Contains(t, got.Resources, ResourceAddress("random_integer.suffix"))
	assert.Contains(t, got.Resources, ResourceAddress("module.child1.random_pet.pet"))
	assert.Contains(t, got.Resources, ResourceAddress("module.child1.random_integer.suffix"))
	assert.Contains(t, got.Resources, ResourceAddress("module.child2.random_integer.suffix"))
	assert.Contains(t, got.Resources, ResourceAddress("module.child2.module.child3.random_integer.suffix"))
	assert.Contains(t, got.Resources, ResourceAddress("module.child2.module.child3.random_pet.pet"))

	assert.True(t, got.Resources["random_pet.pet[3]"].Tainted)
}
