package navigator

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// Navigator navigates the user from page to page, creating and caching
// models accordingly.
type Navigator struct {
	// history tracks the models a user has visited, in LIFO order.
	history []page
	// cache each model visited
	cache *cache
	// directory of model makers for each kind
	makers map[tui.Kind]tui.Maker
	// navigator needs to know width and height when making a model
	width  int
	height int
}

func New(firstPage string, makers map[tui.Kind]tui.Maker) (*Navigator, error) {
	n := &Navigator{
		makers: makers,
		cache: &cache{
			cache: make(map[cacheKey]tea.Model),
		},
	}

	firstKind, err := firstPageKind(firstPage)
	if err != nil {
		return nil, err
	}
	m := page{kind: firstKind, resource: resource.GlobalResource}

	// ignore returned init cmd; instead the main model should invoke it
	if _, err = n.update(GoMsg(m)); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Navigator) SetHeight(h int) {
	n.height = h
}

func (n *Navigator) SetWidth(w int) {
	n.width = w
}

func (n *Navigator) current() page {
	return n.history[len(n.history)-1]
}

func (n *Navigator) CurrentModel() tea.Model {
	return n.cache.get(n.current())
}

func (n *Navigator) Update(msg tea.Msg) tea.Cmd {
	init, err := n.update(msg)
	if err != nil {
		return tui.ReportError(err, "updating navigator")
	}
	return init
}

func (n *Navigator) update(msg tea.Msg) (tea.Cmd, error) {
	switch msg := msg.(type) {
	case SwitchTabMsg:
		// Switch tab on the current page.
		if len(n.history) == 0 {
			// no current page
			return nil, nil
		}
		// Get current page and set tab and add to history
		current := n.current()
		current.tab = tui.TabTitle(msg)
		n.history = append(n.history, current)
		// Instruct current page to switch tabs.
		n.UpdateCurrent(tui.SetActiveTabMsg(current.tab))
	case GoMsg:
		page := page(msg)

		// Silently ignore the user's request to navigate again to the current page.
		if len(n.history) > 0 && page == n.current() {
			return nil, nil
		}

		// Check target page model is cached; if not then create and cache it
		var created bool
		if !n.cache.exists(page) {
			maker, ok := n.makers[page.kind]
			if !ok {
				return nil, fmt.Errorf("no maker could be found for %s", page.kind)
			}
			model, err := maker.Make(page.resource, n.width, n.height)
			if err != nil {
				return nil, fmt.Errorf("making page: %w", err)
			}
			n.cache.put(page, model)
			created = true
		}
		if page.tab != "" {
			n.UpdateCurrent(tui.SetActiveTabMsg(msg.tab))
		}
		// Push new current page to history
		n.history = append(n.history, page)

		if created {
			return n.CurrentModel().Init(), nil
		}
	}
	return nil, nil
}

func (n *Navigator) UpdateCurrent(msg tea.Msg) tea.Cmd {
	return n.cache.update(pageKey(n.current()), msg)
}

func (n *Navigator) UpdateAll(msg tea.Msg) []tea.Cmd {
	return n.cache.updateAll(msg)
}

func (n *Navigator) GoBack() {
	if len(n.history) == 1 {
		// Silently refuse to go back further than first page.
		return
	}
	// Pop current page from history
	n.history = n.history[:len(n.history)-1]
	// Instruct new current page's model to switch tab. If its model doesn't
	// have tabs, it'll simply ignore the message. And if the model does have
	// tabs, but the page didn't name a tab, then it'll instruct the model to
	// switch to the first tab.
	n.UpdateCurrent(tui.SetActiveTabMsg(n.current().tab))
}
