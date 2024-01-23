package internal

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type modulesModel struct {
	list list.Model
}

func newModulesModel(runner *runner) (modulesModel, error) {
	// get modules and convert to items
	modules, err := findModules(".")
	if err != nil {
		return modulesModel{}, err
	}
	items := make([]list.Item, len(modules))
	for i, mod := range modules {
		items[i] = mod
	}
	d := newModuleDelegate(runner)
	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)

	return modulesModel{list: l}, nil
}

func (m modulesModel) Init() tea.Cmd {
	return nil
}

func (m modulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case viewSizeMsg:
		m.list.SetSize(msg.width, msg.height)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modulesModel) View() string {
	return m.list.View()
}
