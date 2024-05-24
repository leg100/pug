package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// model cache: not so much for performance but to retain memory of user
// actions, e.g. a user may select a particular row in a table, navigate away
// from the model and later return to the model, and they would expect the same
// row still to be selected.
type cache struct {
	cache map[cacheKey]tea.Model
}

type cacheKey struct {
	kind tui.Kind
	id   resource.ID
}

func (c *cache) exists(m page) bool {
	_, ok := c.cache[pageKey(m)]
	return ok
}

func (c *cache) get(m page) tea.Model {
	return c.cache[pageKey(m)]
}

func (c *cache) put(m page, model tea.Model) {
	c.cache[pageKey(m)] = model
}

func (c *cache) updateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, len(c.cache))
	var i int
	for k := range c.cache {
		cmds[i] = c.update(k, msg)
		i++
	}
	return cmds
}

func (c *cache) update(key cacheKey, msg tea.Msg) tea.Cmd {
	updated, cmd := c.cache[key].Update(msg)
	c.cache[key] = updated
	return cmd
}

func pageKey(m page) cacheKey {
	key := cacheKey{kind: m.kind}
	if m.resource != nil {
		key.id = m.resource.GetID()
	}
	return key
}
