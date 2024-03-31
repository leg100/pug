package table

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

// Resource is a wrapper of table.Model specifically for use with pug
// resources.
type Resource[K resource.ID, V ResourceValue] struct {
	Model[K, V]

	parent resource.Resource
}

type ResourceOptions[V ResourceValue] struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
	Columns          []Column
	Renderer         RowRenderer[V]
	Width, Height    int
	Parent           resource.Resource
	SortFunc         SortFunc[V]
}

type ResourceValue interface {
	HasAncestor(id resource.ID) bool
	RowKey() resource.ID
	RowParent() *resource.Resource
}

// NewResource creates a new resource table
func NewResource[K resource.ID, V ResourceValue](opts ResourceOptions[V]) Resource[K, V] {
	table := New[K](opts.Columns, opts.Renderer, opts.Width, opts.Height)
	if opts.SortFunc != nil {
		table = table.WithSortFunc(opts.SortFunc)
	}

	return Resource[K, V]{
		Model:  table,
		parent: opts.Parent,
	}
}

// SetItems populates the table with the given items. Each item must be a
// descendent of the table's parent resource.
func (m *Resource[K, V]) SetItems(items map[K]V) {
	for k, v := range items {
		if !v.HasAncestor(m.parent.ID) {
			delete(items, k)
		}
	}
	m.Model.SetItems(items)
}

func (m Resource[K, V]) Update(msg tea.Msg) (Resource[K, V], tea.Cmd) {
	switch msg := msg.(type) {
	case BulkInsertMsg[V]:
		existing := m.Items()
		for _, ws := range msg {
			existing[K(ws.RowKey())] = ws
		}
		m.SetItems(existing)
	case resource.Event[V]:
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			existing := m.Items()
			existing[K(msg.Payload.RowKey())] = msg.Payload
			m.SetItems(existing)
		case resource.DeletedEvent:
			existing := m.Items()
			delete(existing, K(msg.Payload.RowKey()))
			m.SetItems(existing)
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}
