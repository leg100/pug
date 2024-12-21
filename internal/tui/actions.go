package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui/keys"
)

// ActionHandler handles actions common to more than one model.
type ActionHandler struct {
	*Helpers
	IDRetriever
}

type IDRetriever interface {
	GetModuleIDs() ([]resource.ID, error)
	GetWorkspaceIDs() ([]resource.ID, error)
}

func (m *ActionHandler) Update(msg tea.Msg) tea.Cmd {
	var (
		createPlanOptions plan.CreateOptions
		// TODO: if only one worskpace is being applied, then customise message
		// to mention name of workspace being applied.
		applyPrompt = "Auto-apply %d workspaces?"
		upgrade     bool
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.InitUpgrade):
			upgrade = true
			fallthrough
		case key.Matches(msg, keys.Common.Init):
			ids, err := m.GetModuleIDs()
			if err != nil {
				return ReportError(err)
			}
			fn := func(moduleID resource.ID) (task.Spec, error) {
				return m.Modules.Init(moduleID, upgrade)
			}
			return m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Execute):
			ids, err := m.GetModuleIDs()
			if err != nil {
				return ReportError(err)
			}
			return CmdHandler(PromptMsg{
				Prompt:      fmt.Sprintf("Execute program in %d module directories: ", len(ids)),
				Placeholder: "terraform version",
				Action: func(v string) tea.Cmd {
					if v == "" {
						return nil
					}
					// split value into program and any args
					parts := strings.Split(v, " ")
					prog := parts[0]
					args := parts[1:]
					fn := func(moduleID resource.ID) (task.Spec, error) {
						return m.Modules.Execute(moduleID, prog, args...)
					}
					return m.CreateTasks(fn, ids...)
				},
				Key:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
				Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			})
		case key.Matches(msg, keys.Common.Validate):
			ids, err := m.GetModuleIDs()
			if err != nil {
				return ReportError(err)
			}
			cmd := m.CreateTasks(m.Modules.Validate, ids...)
			return cmd
		case key.Matches(msg, keys.Common.Format):
			ids, err := m.GetModuleIDs()
			if err != nil {
				return ReportError(err)
			}
			cmd := m.CreateTasks(m.Modules.Format, ids...)
			return cmd
		case key.Matches(msg, keys.Common.PlanDestroy):
			createPlanOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			ids, err := m.GetWorkspaceIDs()
			if err != nil {
				return ReportError(err)
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Plan(workspaceID, createPlanOptions)
			}
			return m.CreateTasks(fn, ids...)
		case key.Matches(msg, keys.Common.Destroy):
			createPlanOptions.Destroy = true
			applyPrompt = "Destroy resources of %d workspaces?"
			fallthrough
		case key.Matches(msg, keys.Common.AutoApply):
			ids, err := m.GetWorkspaceIDs()
			if err != nil {
				return ReportError(err)
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.Plans.Apply(workspaceID, createPlanOptions)
			}
			return YesNoPrompt(
				fmt.Sprintf(applyPrompt, len(ids)),
				m.CreateTasks(fn, ids...),
			)
		case key.Matches(msg, keys.Common.Cost):
			ids, err := m.GetWorkspaceIDs()
			if err != nil {
				return ReportError(err)
			}
			spec, err := m.Workspaces.Cost(ids...)
			if err != nil {
				return ReportError(fmt.Errorf("creating task: %w", err))
			}
			return m.CreateTasksWithSpecs(spec)
		case key.Matches(msg, keys.Common.State):
			ids, err := m.GetWorkspaceIDs()
			if err != nil {
				return ReportError(err)
			}
			if len(ids) == 0 {
				return nil
			}
			return NavigateTo(ResourceListKind, WithParent(ids[0]))
		}
	}
	return nil
}

func (m *ActionHandler) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Init,
		keys.Common.InitUpgrade,
		keys.Common.Format,
		keys.Common.Validate,
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.AutoApply,
		keys.Common.Destroy,
		keys.Common.Execute,
		keys.Common.State,
		keys.Common.Cost,
	}
}
