package top

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// page cache: not so much for performance but to retain memory of user
// actions, e.g. a user may highlight a particular row in a table, navigate away
// from the page and later return to the page, and they would expect the same
// row still to be hightlighted.
type cache struct {
	cache map[cacheKey]tea.Model
}

type cacheKey struct {
	kind tui.Kind
	id   resource.ID
}

func (c *cache) exists(page tui.Page) bool {
	_, ok := c.cache[pageKey(page)]
	return ok
}

func (c *cache) get(page tui.Page) tea.Model {
	return c.cache[pageKey(page)]
}

func (c *cache) put(page tui.Page, model tea.Model) {
	c.cache[pageKey(page)] = model
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

func pageKey(page tui.Page) cacheKey {
	return cacheKey{kind: page.Kind, id: page.Parent.ID}
}
