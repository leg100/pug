package top

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/tui"
)

// navigator navigates the user from page to page, creating and caching
// corresponding models accordingly.
type navigator struct {
	// history tracks the pages a user has visited, in LIFO order.
	history []tui.Page
	// cache each unique page visited
	cache *tui.Cache
	// directory of model makers for each kind
	makers map[tui.Kind]tui.Maker

	// navigator needs to know width and height when making a model
	width  int
	height int
}

func newNavigator(opts Options, makers map[tui.Kind]tui.Maker) (*navigator, error) {
	n := &navigator{
		makers: makers,
		cache:  tui.NewCache(),
	}

	firstKind, err := tui.FirstPageKind(opts.FirstPage)
	if err != nil {
		return nil, err
	}

	// ignore returned init cmd; instead the main model should invoke it
	if _, err = n.setCurrent(tui.Page{Kind: firstKind}); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *navigator) currentPage() tui.Page {
	return n.history[len(n.history)-1]
}

func (n *navigator) currentModel() tea.Model {
	return n.cache.Get(n.currentPage())
}

func (n *navigator) setCurrent(page tui.Page) (created bool, err error) {
	// Silently ignore the user's request to navigate again to the current page.
	if len(n.history) > 0 && page == n.currentPage() {
		return false, nil
	}

	// Check target page model is cached; if not then create and cache it
	if !n.cache.Exists(page) {
		maker, ok := n.makers[page.Kind]
		if !ok {
			return false, fmt.Errorf("no maker could be found for %s", page.Kind)
		}
		model, err := maker.Make(page.ID, n.width, n.height)
		if err != nil {
			return false, fmt.Errorf("making page: %w", err)
		}
		n.cache.Put(page, model)
		created = true
	}
	// Push new current page to history
	n.history = append(n.history, page)
	return
}

func (n *navigator) updateCurrent(msg tea.Msg) tea.Cmd {
	return n.cache.Update(n.currentPage(), msg)
}

func (n *navigator) goBack() {
	if len(n.history) == 1 {
		// Silently refuse to go back further than first page.
		return
	}
	// Pop current page from history
	n.history = n.history[:len(n.history)-1]
}
