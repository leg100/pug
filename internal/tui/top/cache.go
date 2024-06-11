package top

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// page cache: not so much for performance but to retain memory of user actions,
// e.g. a user may select a particular row in a table, navigate away from the
// page and later return to the page, and they would expect the same row still
// to be selected.
type cache struct {
	m map[cacheKey]tea.Model
}

func newCache() *cache {
	return &cache{
		m: make(map[cacheKey]tea.Model),
	}
}

type cacheKey struct {
	kind tui.Kind
	id   resource.ID
}

func (c *cache) Get(page tui.Page) tea.Model {
	return c.m[newCacheKey(page)]
}

func (c *cache) Put(page tui.Page, model tea.Model) {
	c.m[newCacheKey(page)] = model
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

func (c *cache) Update(key cacheKey, msg tea.Msg) tea.Cmd {
	updated, cmd := c.m[key].Update(msg)
	c.m[key] = updated
	return cmd
}

func newCacheKey(page tui.Page) cacheKey {
	key := cacheKey{kind: page.Kind}
	if page.Resource != nil {
		key.id = page.Resource.GetID()
	}
	return key
}
