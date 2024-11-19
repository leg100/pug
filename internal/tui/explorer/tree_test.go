package explorer

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	wd := internal.NewTestWorkdir(t)
	mod1 := &module.Module{
		ID:   resource.NewID(resource.Module),
		Path: "a",
	}
	mod2 := &module.Module{
		ID:   resource.NewID(resource.Module),
		Path: "a/b",
	}
	mod3 := &module.Module{
		ID:   resource.NewID(resource.Module),
		Path: "a/b/c",
	}
	ws1 := &workspace.Workspace{
		ID:       resource.NewID(resource.Workspace),
		ModuleID: mod1.ID,
		Name:     "ws1",
	}
	ws2 := &workspace.Workspace{
		ID:       resource.NewID(resource.Workspace),
		ModuleID: mod1.ID,
		Name:     "ws2",
	}
	ws3 := &workspace.Workspace{
		ID:       resource.NewID(resource.Workspace),
		ModuleID: mod1.ID,
		Name:     "ws3",
	}
	got := newTree(
		wd,
		[]*module.Module{mod1, mod2, mod3},
		[]*workspace.Workspace{ws1, ws2, ws3},
	)
	want := &tree{
		value: dirNode{path: wd.String()},
		children: []*tree{
			{
				value: moduleNode{path: "a"},
			},
			{
				value: dirNode{path: "a"},
				children: []*tree{
					{
						value: moduleNode{path: "a/b"},
					},
					{
						value: dirNode{path: "a/b"},
					},
				},
			},
		},
	}
	assert.Equal(t, want, got)
}

func TestSplitDirs(t *testing.T) {
	got := splitDirs("a/b/c/d")
	want := []string{
		"a",
		"a/b",
		"a/b/c",
		"a/b/c/d",
	}
	assert.Equal(t, want, got)
}
