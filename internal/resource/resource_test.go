package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	mod := New(Module, "a/b/c", nil)
	ws := New(Workspace, "dev", &mod)
	run := New(Run, "", &ws)
	task := New(Task, "", &run)

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, task.id.String(), task.String())
		assert.Equal(t, "dev", ws.String())
	})

	t.Run("ancestors", func(t *testing.T) {
		got := task.Ancestors()

		assert.Equal(t, 3, len(got))
		assert.Equal(t, Run, got[0].Kind)
		assert.Equal(t, Workspace, got[1].Kind)
		assert.Equal(t, Module, got[2].Kind)
	})

	t.Run("has ancestor", func(t *testing.T) {
		assert.True(t, task.HasAncestor(mod.ID()()))
		assert.False(t, mod.HasAncestor(task.ID()()))
	})

	t.Run("module", func(t *testing.T) {
		assert.Equal(t, &mod, task.Module())
	})

	t.Run("workspace", func(t *testing.T) {
		assert.Equal(t, &ws, task.Workspace())
	})
}
