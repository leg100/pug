package resource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID_String(t *testing.T) {
	mod := NewMonotonicID(Module)
	ws := NewMonotonicID(Workspace)
	run := NewMonotonicID(Plan)
	task := NewMonotonicID(Task)

	t.Run("string", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(mod.String(), "#"))
		assert.True(t, strings.HasPrefix(ws.String(), "#"))
		assert.True(t, strings.HasPrefix(run.String(), "#"))
		assert.True(t, strings.HasPrefix(task.String(), "#"))
	})
}
