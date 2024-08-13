package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	mod := New(Module, GlobalResource)
	ws := New(Workspace, mod)
	run := New(Plan, ws)
	task := New(Task, run)

	t.Run("has ancestor", func(t *testing.T) {
		assert.True(t, task.HasAncestor(mod.ID))
		assert.False(t, mod.HasAncestor(task.ID))
	})

	t.Run("get ancestor of specific kind", func(t *testing.T) {
		assert.Equal(t, mod, task.Module())
		assert.Equal(t, ws, task.Workspace())
		assert.Equal(t, run, task.Run())
	})
}
