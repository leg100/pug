package workspace

import (
	"strings"
	"testing"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_parseList(t *testing.T) {
	r := strings.NewReader(`
  default
* foo
  bar
`)

	got, current, err := parseList(r)
	require.NoError(t, err)

	assert.Len(t, got, 3)
	assert.Contains(t, got, "default")
	assert.Contains(t, got, "foo")
	assert.Contains(t, got, "bar")

	assert.Equal(t, "foo", current)
}

func TestWorkspace_resetWorkspaces(t *testing.T) {
	mod := &module.Module{Path: "a/b/c"}

	dev, err := New(mod, "dev")
	require.NoError(t, err)

	staging, err := New(mod, "staging")
	require.NoError(t, err)

	table := &fakeWorkspaceTable{
		existing: []*Workspace{dev, staging},
	}
	moduleService := &fakeModuleService{}
	reloader := &reloader{
		&Service{
			modules: moduleService,
			table:   table,
		},
	}
	_, _, err = reloader.resetWorkspaces(mod, []string{"dev", "prod"}, "dev")
	require.NoError(t, err)

	// expect staging to be dropped
	assert.Equal(t, []resource.Identity{staging.ID}, table.deleted)

	// expect prod to be added
	assert.Len(t, table.added, 1)
	assert.Equal(t, "prod", table.added[0].Name)

	// expect dev to have been made the current workspace
	assert.Equal(t, dev.ID, moduleService.current)
}

type fakeModuleService struct {
	current resource.Identity

	modules
}

func (f *fakeModuleService) SetCurrent(moduleID, workspaceID resource.Identity) error {
	f.current = workspaceID
	return nil
}

type fakeWorkspaceTable struct {
	existing []*Workspace
	added    []*Workspace
	deleted  []resource.Identity

	workspaceTable
}

func (f *fakeWorkspaceTable) Add(id resource.Identity, row *Workspace) {
	f.added = append(f.added, row)
}

func (f *fakeWorkspaceTable) Delete(id resource.Identity) {
	f.deleted = append(f.deleted, id)
}

func (f *fakeWorkspaceTable) List() []*Workspace {
	return f.existing
}
