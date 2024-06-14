package split

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
)

type Preview interface {
	tea.Model

	SetBorderStyle(lipgloss.Style)
}

// page cache: not so much for performance but to retain memory of user actions,
// e.g. a user may select a particular row in a table, navigate away from the
// page and later return to the page, and they would expect the same row still
// to be selected.
type cache struct {
	cache map[resource.ID]tea.Model
}

func newCache() *cache {
	return &cache{
		cache: make(map[resource.ID]tea.Model),
	}
}

func (c *cache) Get(id resource.ID) tea.Model {
	return c.cache[id]
}

func (c *cache) Put(id resource.ID, model tea.Model) {
	c.cache[id] = model
}

func (c *cache) UpdateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(c.cache))
	var i int
	for k := range c.cache {
		cmds[i] = c.Update(k, msg)
		i++
	}
	return cmds
}

func (c *cache) Update(id resource.ID, msg tea.Msg) tea.Cmd {
	model, ok := c.cache[id]
	if !ok {
		return nil
	}
	updated, cmd := model.Update(msg)
	c.cache[id] = updated
	return cmd
}
