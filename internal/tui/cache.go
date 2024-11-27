package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// page Cache: not so much for performance but to retain memory of user actions,
// e.g. a user may select a particular row in a table, navigate away from the
// page and later return to the page, and they would expect the same row still
// to be selected.
type Cache struct {
	cache map[Page]tea.Model
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[Page]tea.Model),
	}
}

func (c *Cache) Get(page Page) tea.Model {
	return c.cache[page]
}

func (c *Cache) Put(page Page, model tea.Model) {
	c.cache[page] = model
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

func (c *Cache) Update(key Page, msg tea.Msg) tea.Cmd {
	updated, cmd := c.cache[key].Update(msg)
	c.cache[key] = updated
	return cmd
}
