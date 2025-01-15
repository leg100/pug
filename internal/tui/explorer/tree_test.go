package explorer

import (
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
	"github.com/stretchr/testify/assert"
)

var (
	mod1 = &module.Module{
		ID:   resource.NewMonotonicID(resource.Module),
		Path: "a",
	}
	mod2 = &module.Module{
		ID:   resource.NewMonotonicID(resource.Module),
		Path: "a/b",
	}
	mod3 = &module.Module{
		ID:   resource.NewMonotonicID(resource.Module),
		Path: "a/b/c",
	}
	ws1 = &workspace.Workspace{
		ID:       resource.NewMonotonicID(resource.Workspace),
		ModuleID: mod1.ID,
		Name:     "ws1",
	}
	ws2 = &workspace.Workspace{
		ID:       resource.NewMonotonicID(resource.Workspace),
		ModuleID: mod2.ID,
		Name:     "ws2",
	}
	ws3 = &workspace.Workspace{
		ID:       resource.NewMonotonicID(resource.Workspace),
		ModuleID: mod3.ID,
		Name:     "ws3",
	}
)

func TestTree(t *testing.T) {
	builder := &treeBuilder{
		wd: internal.NewTestWorkdir(t),
		moduleService: &fakeTreeBuilderModuleLister{
			modules: []*module.Module{mod1, mod2, mod3},
		},
		workspaceService: &fakeTreeBuilderWorkspaceLister{
			workspaces: []*workspace.Workspace{ws1, ws2, ws3},
		},
		helpers: &fakeTreeBuilderHelpers{},
	}

	got, _ := builder.newTree("")

	want := &tree{
		value: dirNode{path: builder.wd.String(), root: true},
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

func TestFilter(t *testing.T) {
	unfiltered := &tree{
		value: dirNode{path: "/root", root: true},
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
	want := &tree{
		value: dirNode{path: "/root", root: true},
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
		},
	}
	got := unfiltered.filter("b")
	assert.Equal(t, want, got)
}

func TestSplitDirs(t *testing.T) {
	got := splitDirs("a/b/c/d")
	want := []string{
		"a",
		"a/b",
		"a/b/c",
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

type fakeTreeBuilderModuleLister struct {
	modules []*module.Module
}

func (f *fakeTreeBuilderModuleLister) List() []*module.Module {
	return f.modules
}

type fakeTreeBuilderWorkspaceLister struct {
	workspaces []*workspace.Workspace
}

func (f *fakeTreeBuilderWorkspaceLister) List(workspace.ListOptions) []*workspace.Workspace {
	return f.workspaces
}

type fakeTreeBuilderHelpers struct{}

func (f *fakeTreeBuilderHelpers) WorkspaceResourceCount(*workspace.Workspace) string {
	return ""
}

func (f *fakeTreeBuilderHelpers) WorkspaceCost(*workspace.Workspace) string {
	return ""
}
