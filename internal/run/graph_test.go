package run

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestTopologicalSort(t *testing.T) {
	wd, _ := internal.NewWorkdir("")

	mod1 := module.New(wd, "one")
	// mod2 depends on mod1
	mod2 := module.New(wd, "two", mod1)
	// mod3 depends on mod1 and mod2
	mod3 := module.New(wd, "three", mod1, mod2)

	// Mimic 9 workspaces, with 3 workspaces belonging to each module.
	mod1ws1ID := resource.NewID(resource.Workspace)
	mod1ws2ID := resource.NewID(resource.Workspace)
	mod1ws3ID := resource.NewID(resource.Workspace)
	mod2ws1ID := resource.NewID(resource.Workspace)
	mod2ws2ID := resource.NewID(resource.Workspace)
	mod2ws3ID := resource.NewID(resource.Workspace)
	mod3ws1ID := resource.NewID(resource.Workspace)
	mod3ws2ID := resource.NewID(resource.Workspace)
	mod3ws3ID := resource.NewID(resource.Workspace)

	g := newGraph(
		moduleAndResource{module: mod3, id: mod3ws1ID},
		moduleAndResource{module: mod3, id: mod3ws2ID},
		moduleAndResource{module: mod3, id: mod3ws3ID},
		moduleAndResource{module: mod2, id: mod2ws1ID},
		moduleAndResource{module: mod2, id: mod2ws2ID},
		moduleAndResource{module: mod2, id: mod2ws3ID},
		moduleAndResource{module: mod1, id: mod1ws1ID},
		moduleAndResource{module: mod1, id: mod1ws2ID},
		moduleAndResource{module: mod1, id: mod1ws3ID},
	)
	// Should be one graph vertex per module
	assert.Len(t, g.vertices, 3)
	g.sort()

	// Expect order of module IDs to be reversed.
	want := [][]resource.ID{
		{mod1ws1ID, mod1ws2ID, mod1ws3ID},
		{mod2ws1ID, mod2ws2ID, mod2ws3ID},
		{mod3ws1ID, mod3ws2ID, mod3ws3ID},
	}
	if assert.Len(t, g.results, 3) {
		assert.Equal(t, want, g.results)
	}
}
