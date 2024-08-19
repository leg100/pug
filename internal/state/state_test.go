package state

import (
	"bytes"
	"os"
	"testing"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	mod := module.New(module.Options{Path: "a/b/c"})
	ws, err := workspace.New(mod, "dev")
	require.NoError(t, err)

	t.Run("new state", func(t *testing.T) {
		// Mimic response from terraform state pull
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

		wantAttrs := map[string]any{
			"id":        "next-thrush",
			"keepers":   map[string]any{"now": "2024-05-17T17:23:31Z"},
			"length":    float64(2),
			"prefix":    nil,
			"separator": "-",
		}
		assert.Equal(t, wantAttrs, got.Resources["random_pet.pet[3]"].Attributes)

		assert.True(t, got.Resources["random_pet.pet[3]"].Tainted)
	})

	t.Run("empty", func(t *testing.T) {
		// Mimic empty response from terraform state pull
		f := new(bytes.Buffer)

		got, err := newState(ws, f)
		require.NoError(t, err)

		assert.Len(t, got.Resources, 0)
		assert.Equal(t, int64(-1), got.Serial)
	})

	t.Run("state resource with string index key", func(t *testing.T) {
		// Mimic response from terraform state pull
		f, err := os.Open("./testdata/state_with_for_each.json")
		require.NoError(t, err)
		t.Cleanup(func() {
			f.Close()
		})

		got, err := newState(ws, f)
		require.NoError(t, err)

		assert.Len(t, got.Resources, 1)

		assert.Contains(t, got.Resources, ResourceAddress(`time_sleep.wait_three_seconds["duration"]`))
		wantAttrs := map[string]any{
			"id":               "2024-07-15T06:12:24Z",
			"destroy_duration": nil,
			"create_duration":  "3s",
			"triggers":         nil,
		}
		assert.Equal(t, wantAttrs, got.Resources[`time_sleep.wait_three_seconds["duration"]`].Attributes)
	})

}
