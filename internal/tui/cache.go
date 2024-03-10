package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// page cache: not so much for performance but to retain memory of user
// actions, e.g. a user may highlight a particular row in a table, navigate away
// from the page and later return to the page, and they would expect the same
// row still to be hightlighted.
type cache struct {
	cache map[cacheKey]Model
}

type cacheKey struct {
	kind modelKind
	id   resource.ID
}

func (c *cache) exists(page page) bool {
	_, ok := c.cache[page.cacheKey()]
	return ok
}

func (c *cache) get(page page) Model {
	return c.cache[page.cacheKey()]
}

func (c *cache) put(page page, model Model) {
	c.cache[page.cacheKey()] = model
}

func (c *cache) updateAll(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	for k := range c.cache {
		cmds = append(cmds, c.update(k, msg))
	}
	return cmds
}

func (c *cache) update(key cacheKey, msg tea.Msg) tea.Cmd {
	updated, cmd := c.cache[key].Update(msg)
	c.cache[key] = updated
	return cmd
}
