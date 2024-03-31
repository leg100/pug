package resource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	mod := New(Module, GlobalResource)
	ws := New(Workspace, mod)
	run := New(Run, ws)
	task := New(Task, run)

	t.Run("string", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(mod.String(), "mod-"))
		assert.True(t, strings.HasPrefix(ws.String(), "ws-"))
		assert.True(t, strings.HasPrefix(run.String(), "run-"))
		assert.True(t, strings.HasPrefix(task.String(), "task-"))
	})

	t.Run("has ancestor", func(t *testing.T) {
		assert.True(t, task.HasAncestor(mod.ID))
		assert.False(t, mod.HasAncestor(task.ID))
	})

	t.Run("get ancestor of specific kind", func(t *testing.T) {
		assert.Equal(t, &mod, task.Module())
		assert.Equal(t, &ws, task.Workspace())
		assert.Equal(t, &run, task.Run())
	})
}
