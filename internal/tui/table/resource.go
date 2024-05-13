package table

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// Resource is a wrapper of table.Model specifically for use with pug
// resources.
type Resource[K resource.ID, V ResourceValue] struct {
	Model[K, V]

	parent resource.Resource
}

type ResourceOptions[V ResourceValue] struct {
	Columns       []Column
	Renderer      RowRenderer[V]
	Width, Height int
	Parent        resource.Resource
	SortFunc      SortFunc[V]
}

type ResourceValue interface {
	HasAncestor(id resource.ID) bool
	RowKey() resource.ID
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

func (m Resource[K, V]) Update(msg tea.Msg) (Resource[K, V], tea.Cmd) {
	switch msg := msg.(type) {
	case BulkInsertMsg[V]:
		existing := m.Items()
		for _, ws := range msg {
			existing[K(ws.RowKey())] = ws
		}
		m.setItems(existing)
	case resource.Event[V]:
		switch msg.Type {
		case resource.CreatedEvent, resource.UpdatedEvent:
			existing := m.Items()
			existing[K(msg.Payload.RowKey())] = msg.Payload
			m.setItems(existing)
		case resource.DeletedEvent:
			existing := m.Items()
			delete(existing, K(msg.Payload.RowKey()))
			m.setItems(existing)
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

// setItems populates the table with the given items. Items that are not a
// descendent of the table's parent resource are skipped.
func (m *Resource[K, V]) setItems(items map[K]V) {
	for k, v := range items {
		if !v.HasAncestor(m.parent.ID) {
			delete(items, k)
		}
	}
	m.Model.SetItems(items)
}

// Prune passes each value from the selected rows (or if there are no
// selections, from the highlighted row) to the provided func. If the func
// returns an error the row is de-selected (or if there are no selections, then
// an error is returned). The resulting IDs from the provided func are returned.
// If all selections are de-selected then an error is returned.
func (m *Resource[K, V]) Prune(fn func(value V) (resource.ID, error)) ([]resource.ID, error) {
	rows := m.HighlightedOrSelected()
	switch len(rows) {
	case 0:
		return nil, errors.New("no rows in table")
	case 1:
		// highlighted, no selections
		id, err := fn(rows[0].Value)
		if err != nil {
			// the single highlighted value is to be pruned, so report this as an
			// error
			return nil, err
		}
		return []resource.ID{id}, nil
	default:
		// one or more selections: iterate thru and prune accordingly.
		var ids []resource.ID
		for k, v := range m.Selected {
			id, err := fn(v)
			if err != nil {
				// De-select
				m.ToggleSelectionByKey(k)
				continue
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			// no rows survived pruning, so report error
			return nil, errors.New("no rows are applicable to the given action")
		}
		return ids, nil
	}
}
