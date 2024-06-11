package top

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
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

func newNavigator(opts Options, spinner *spinner.Model) (*navigator, error) {
	n := &navigator{
		makers: makeMakers(opts, spinner),
		cache:  newCache(),
	}

	firstKind, err := tui.FirstPageKind(opts.FirstPage)
	if err != nil {
		return nil, err
	}

	// ignore returned init cmd; instead the main model should invoke it
	if _, err = n.setCurrent(tui.Page{Kind: firstKind, Resource: resource.GlobalResource}); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *navigator) currentPage() (tui.Page, bool) {
	if len(n.history) == 0 {
		return tui.Page{}, false
	}
	return n.history[len(n.history)-1], true
}

func (n *navigator) currentModel() tea.Model {
	page, ok := n.currentPage()
	if !ok {
		return nil
	}
	return n.cache.Get(page)
}

func (n *navigator) setCurrent(page tui.Page) (tea.Cmd, error) {
	// Silently ignore the user's request to navigate again to the current page.
	if current, ok := n.currentPage(); ok && page == current {
		return nil, nil
	}

	// Check if target page model is cached; if not then create and cache it
	model := n.cache.Get(page)
	if model == nil {
		maker, ok := n.makers[page.Kind]
		if !ok {
			return nil, fmt.Errorf("no maker could be found for %s", page.Kind)
		}
		var err error
		model, err = maker.Make(page.Resource, n.width, n.height)
		if err != nil {
			return nil, fmt.Errorf("making page: %w", err)
		}
		n.cache.Put(page, model)
	}
	// Push new current page to history
	n.history = append(n.history, page)
	return model.Init(), nil
}

func (n *navigator) updateCurrent(msg tea.Msg) tea.Cmd {
	page, ok := n.currentPage()
	if !ok {
		return nil
	}
	return n.cache.Update(newCacheKey(page), msg)
}

func (n *navigator) goBack() tea.Cmd {
	if len(n.history) == 1 {
		// Silently refuse to go back further than first page.
		return nil
	}
	// Pop current page from history
	n.history = n.history[:len(n.history)-1]
	// If either previous current model or new current model are a viewport then
	// we must explicitly switch from one to the other.
	return n.currentModel().Init()
}
