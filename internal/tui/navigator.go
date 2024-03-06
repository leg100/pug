package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui/common"
)

type navigationMsg struct {
	target page
}

func navigate(target page) tea.Cmd {
	return func() tea.Msg {
		return navigationMsg{target: target}
	}
}

type maker interface {
	makeModel(target resource.Resource) (common.Model, error)
}

// navigator traverses the user from page to page, creating and caching the
// corresponding models accordingly.
type navigator struct {
	// history tracks the pages a user has visited, in LIFO order.
	history []page
	// cache each unique page visited
	cache *cache
	// directory of model makers for each kind
	makers map[modelKind]maker
}

func newNavigator(start modelKind, makers map[modelKind]maker) (*navigator, error) {
	n := &navigator{
		makers: makers,
		cache: &cache{
			cache: make(map[cacheKey]common.Model),
		},
	}
	// ignore returned init cmd; instead the main model should invoke it
	_, err := n.setCurrent(navigationMsg{target: page{kind: start}})
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *navigator) currentPage() page {
	return n.history[len(n.history)-1]
}

func (n *navigator) currentModel() common.Model {
	return n.cache.get(n.currentPage())
}

func (n *navigator) setCurrent(msg navigationMsg) (created bool, err error) {
	// Silently ignore the user's request to navigate again to the current page.
	if len(n.history) > 0 && msg.target == n.currentPage() {
		return false, nil
	}

	// Push target page to history
	n.history = append(n.history, msg.target)
	// Check target page model is cached; if not then create and cache it
	if !n.cache.exists(msg.target) {
		model, err := n.makers[n.currentPage().kind].makeModel(n.currentPage().resource)
		if err != nil {
			return false, err
		}
		n.cache.put(n.currentPage(), model)
		created = true
	}
	return
}

func (n *navigator) updateCurrent(msg tea.Msg) tea.Cmd {
	return n.cache.update(n.currentPage().cacheKey(), msg)
}

func (n *navigator) goBack() {
	if len(n.history) == 1 {
		// Silently refuse to go back further than first page.
		return
	}
	// Pop current page from history
	n.history = n.history[:len(n.history)-1]
}
