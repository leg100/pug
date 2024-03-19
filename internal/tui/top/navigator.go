package top

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
)

// navigator navigates the user from page to page, creating and caching
// corresponding models accordingly.
type navigator struct {
	// history tracks the pages a user has visited, in LIFO order.
	history []tui.Page
	// cache each unique page visited
	cache *cache
	// directory of model makers for each kind
	makers map[tui.Kind]tui.Maker
	// navigator needs to know width and height when making a model
	width  int
	height int
}

func newNavigator(start tui.Kind, makers map[tui.Kind]tui.Maker) (*navigator, error) {
	n := &navigator{
		makers: makers,
		cache: &cache{
			cache: make(map[cacheKey]tui.Model),
		},
	}
	// ignore returned init cmd; instead the main model should invoke it
	_, err := n.setCurrent(tui.Page{Kind: start})
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *navigator) currentPage() tui.Page {
	return n.history[len(n.history)-1]
}

func (n *navigator) currentModel() tui.Model {
	return n.cache.get(n.currentPage())
}

func (n *navigator) setCurrent(page tui.Page) (created bool, err error) {
	// Silently ignore the user's request to navigate again to the current page.
	if len(n.history) > 0 && page == n.currentPage() {
		return false, nil
	}

	// Push target page to history
	n.history = append(n.history, page)
	// Check target page model is cached; if not then create and cache it
	if !n.cache.exists(page) {
		model, err := n.makers[n.currentPage().Kind].Make(n.currentPage().Resource, n.width, n.height)
		if err != nil {
			return false, err
		}
		n.cache.put(n.currentPage(), model)
		created = true
	}
	return
}

func (n *navigator) updateCurrent(msg tea.Msg) tea.Cmd {
	return n.cache.update(pageKey(n.currentPage()), msg)
}

func (n *navigator) goBack() {
	if len(n.history) == 1 {
		// Silently refuse to go back further than first page.
		return
	}
	// Pop current page from history
	n.history = n.history[:len(n.history)-1]
}
