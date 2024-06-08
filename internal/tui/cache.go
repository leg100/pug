package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
)

// page Cache: not so much for performance but to retain memory of user actions,
// e.g. a user may select a particular row in a table, navigate away from the
// page and later return to the page, and they would expect the same row still
// to be selected.
type Cache struct {
	cache map[cacheKey]tea.Model
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[cacheKey]tea.Model),
	}
}

type cacheKey struct {
	kind Kind
	id   resource.ID
}

func (c *Cache) Exists(page Page) bool {
	_, ok := c.cache[NewCacheKey(page)]
	return ok
}

func (c *Cache) Entries() []resource.ID {
	ids := make([]resource.ID, len(c.cache))
	var i int
	for k := range c.cache {
		ids[i] = k.id
		i++
	}
	return ids
}

func (c *Cache) Get(page Page) tea.Model {
	return c.cache[NewCacheKey(page)]
}

func (c *Cache) Put(page Page, model tea.Model) {
	c.cache[NewCacheKey(page)] = model
}

func (c *Cache) UpdateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(c.cache))
	var i int
	for k := range c.cache {
		cmds[i] = c.Update(k, msg)
		i++
	}
	return cmds
}

func (c *Cache) Update(key cacheKey, msg tea.Msg) tea.Cmd {
	updated, cmd := c.cache[key].Update(msg)
	c.cache[key] = updated
	return cmd
}

func NewCacheKey(page Page) cacheKey {
	key := cacheKey{kind: page.Kind}
	if page.Resource != nil {
		key.id = page.Resource.GetID()
	}
	return key
}
