package split

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

type cache struct {
	m map[resource.ID]tea.Model
}

func newCache() *cache {
	return &cache{
		m: make(map[resource.ID]tea.Model),
	}
}

func (c *cache) Get(id resource.ID) tea.Model {
	return c.m[id]
}

func (c *cache) Put(id resource.ID, model tea.Model) {
	c.m[id] = model
}

func (c *cache) UpdateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(c.m))
	var i int
	for k := range c.m {
		cmds[i] = c.Update(k, msg)
		i++
	}
	return cmds
}

func (c *cache) Update(key resource.ID, msg tea.Msg) tea.Cmd {
	updated, cmd := c.m[key].Update(msg)
	c.m[key] = updated
	return cmd
}
