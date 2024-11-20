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
		ModuleID: mod2.ID,
		Name:     "ws2",
	}
	ws3 := &workspace.Workspace{
		ID:       resource.NewID(resource.Workspace),
		ModuleID: mod3.ID,
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
				value: dirNode{path: "a"},
				children: []*tree{
					{
						value: dirNode{path: "a/b"},
						children: []*tree{
							{
								value: moduleNode{id: mod3.ID, path: "a/b/c"},
								children: []*tree{
									{
										value: workspaceNode{id: ws3.ID, name: "ws3"},
									},
								},
							},
						},
					},
					{
						value: moduleNode{id: mod2.ID, path: "a/b"},
						children: []*tree{
							{
								value: workspaceNode{id: ws2.ID, name: "ws2"},
							},
						},
					},
				},
			},
			{
				value: moduleNode{id: mod1.ID, path: "a"},
				children: []*tree{
					{
						value: workspaceNode{id: ws1.ID, name: "ws1"},
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

func TestSplitDirs_OneDirectory(t *testing.T) {
	got := splitDirs("a")
	assert.Equal(t, []string(nil), got)
}

func TestSplitDirs_OneSubdirectory(t *testing.T) {
	got := splitDirs("a/b")
	assert.Equal(t, []string{"a"}, got)
}
