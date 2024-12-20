package tui

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

// Helper methods for easily surfacing info in the TUI.
//
// TODO: leverage a cache to enhance performance, particularly if we introduce
// sqlite at some stage. These helpers are invoked on every render, which for a
// table with, say 40 visible rows, means they are invoked 40 times a render,
// which is 40 lookups.
type Helpers struct {
	Modules    *module.Service
	Workspaces *workspace.Service
	Plans      *plan.Service
	Tasks      *task.Service
	States     *state.Service
	Logger     logging.Interface
}

func (h *Helpers) ModuleCurrentWorkspace(mod *module.Module) *workspace.Workspace {
	if mod.CurrentWorkspaceID == nil {
		return nil
	}
	ws, err := h.Workspaces.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("retrieving current workspace for module", "error", err, "module", mod)
		return nil
	}
	return ws
}

func (h *Helpers) CurrentWorkspaceName(workspaceID *resource.ID) string {
	if workspaceID == nil {
		return "-"
	}
	ws, err := h.Workspaces.Get(*workspaceID)
	if err != nil {
		h.Logger.Error("rendering current workspace name", "error", err)
		return ""
	}
	return ws.Name
}

func (h *Helpers) ModuleCurrentResourceCount(mod *module.Module) string {
	if mod.CurrentWorkspaceID == nil {
		return ""
	}
	ws, err := h.Workspaces.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("rendering module current workspace resource count", "error", err)
		return ""
	}
	return h.WorkspaceResourceCount(ws)
}

// WorkspaceCurrentCheckmark returns a check mark if the workspace is the
// current workspace for its module.
func (h *Helpers) WorkspaceCurrentCheckmark(ws *workspace.Workspace) string {
	mod, err := h.Modules.Get(ws.ModuleID)
	if err != nil {
		h.Logger.Error("rendering current workspace checkmark", "error", err)
		return ""
	}
	if mod.CurrentWorkspaceID != nil && *mod.CurrentWorkspaceID == ws.ID {
		return "✓"
	}
	return ""
}

// ModuleCost renders the cost of the module's current workspace, if it has one.
func (h *Helpers) ModuleCost(mod *module.Module) string {
	if ws := h.ModuleCurrentWorkspace(mod); ws != nil {
		return h.WorkspaceCost(ws)
	}
	return ""
}

// WorkspaceCost renders the cost of the given workspace.
func (h *Helpers) WorkspaceCost(ws *workspace.Workspace) string {
	if ws.Cost != nil {
		return fmt.Sprintf("$%.2f", *ws.Cost)
	}
	return ""
}

func (h *Helpers) WorkspaceResourceCount(ws *workspace.Workspace) string {
	state, err := h.States.Get(ws.ID)
	if errors.Is(err, resource.ErrNotFound) {
		// not found most likely means state not loaded yet
		return ""
	} else if err != nil {
		h.Logger.Error("rendering workspace resource count", "error", err)
		return ""
	}
	return strconv.Itoa(len(state.Resources))
}

func (h *Helpers) TaskModule(t *task.Task) *module.Module {
	moduleID := t.ModuleID
	if moduleID == nil {
		return nil
	}
	mod, err := h.Modules.Get(*moduleID)
	if err != nil {
		return nil
	}
	return mod
}

func (h *Helpers) TaskModulePath(t *task.Task) string {
	if mod := h.TaskModule(t); mod != nil {
		return ModuleStyle.Render(mod.Path)
	}
	return ""
}

func (h *Helpers) TaskModulePathWithIcon(t *task.Task) string {
	if mod := h.TaskModule(t); mod != nil {
		return ModulePathWithIcon(mod.Path, true)
	}
	return ""
}

// TaskWorkspace retrieves the task's workspace if it belongs to one.
func (h *Helpers) TaskWorkspace(t *task.Task) *workspace.Workspace {
	workspaceID := t.WorkspaceID
	if workspaceID == nil {
		return nil
	}
	ws, err := h.Workspaces.Get(*workspaceID)
	if err != nil {
		return nil
	}
	return ws
}

func (h *Helpers) TaskWorkspaceName(t *task.Task) string {
	if ws := h.TaskWorkspace(t); ws != nil {
		return WorkspaceName(ws.Name)
	}
	return ""
}

func (h *Helpers) TaskWorkspaceNameWithIcon(t *task.Task) string {
	if ws := h.TaskWorkspace(t); ws != nil {
		return WorkspaceNameWithIcon(ws.Name, true)
	}
	return ""
}

func ModulePath(modulePath string) string {
	return ModuleStyle.Render(modulePath)
}

func ModuleIcon() string {
	return ModuleStyle.Render(fmt.Sprintf("%s ", moduleIcon))
}

func ModulePathWithIcon(modulePath string, squareBrackets bool) string {
	s := fmt.Sprintf("%s%s",
		ModuleIcon(),
		ModulePath(modulePath),
	)
	return s
}

func WorkspaceName(name string) string {
	return WorkspaceStyle.Render(name)
}

func WorkspaceIcon() string {
	return WorkspaceStyle.Render(fmt.Sprintf("%s ", workspaceIcon))
}

func WorkspaceNameWithIcon(name string, squareBrackets bool) string {
	s := fmt.Sprintf("%s%s",
		WorkspaceIcon(),
		WorkspaceName(name),
	)
	return s
}

// TaskWorkspaceOrCurrentWorkspace retrieves either the task's workspace if it belongs to a
// workspace, or if it belongs to a module, then it retrieves the module's
// current workspace
func (h *Helpers) TaskWorkspaceOrCurrentWorkspace(t *task.Task) *workspace.Workspace {
	if ws := h.TaskWorkspace(t); ws != nil {
		return ws
	}
	if mod := h.TaskModule(t); mod != nil {
		if ws := h.ModuleCurrentWorkspace(mod); ws != nil {
			return ws
		}
		return nil
	}
	return nil
}

// TaskStatus provides a rendered colored task status.
func (h *Helpers) TaskStatus(t *task.Task, table bool) string {
	var color lipgloss.Color

	switch t.State {
	case task.Pending:
		color = Grey
	case task.Queued:
		color = Orange
	case task.Running:
		color = Blue
	case task.Exited:
		color = GreenBlue
	case task.Errored:
		color = Red
	}

	return Regular.Foreground(color).Render(string(t.State))
}

// TaskSummary renders a summary of the task's outcome.
func (h *Helpers) TaskSummary(t *task.Task, table bool) string {
	if t.Summary == nil {
		return ""
	}
	style := lipgloss.NewStyle()
	// Render special resource report
	var content string
	switch summary := t.Summary.(type) {
	case plan.Report:
		content = h.ResourceReport(summary, style)
	case workspace.ReloadSummary:
		content = h.WorkspaceReloadReport(summary, style)
	case workspace.CostSummary:
		content = h.CostSummary(summary, style)
	case state.ReloadSummary:
		content = h.StateReloadReport(summary, style)
	default:
		content = t.Summary.String()
	}
	return content
}

// ResourceReport renders a colored summary of resource changes as a result of a
// plan or apply.
func (h *Helpers) ResourceReport(report plan.Report, inherit lipgloss.Style) string {
	additions := Regular.Foreground(Green).Inherit(inherit).Render(fmt.Sprintf("+%d", report.Additions))
	changes := Regular.Foreground(Blue).Inherit(inherit).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := Regular.Foreground(Red).Inherit(inherit).Render(fmt.Sprintf("-%d", report.Destructions))

	return fmt.Sprintf("%s%s%s", additions, changes, destructions)
}

// WorkspaceReloadReport renders a colored summary of workspaces added or
// removed as a result of a workspace reload.
func (h *Helpers) WorkspaceReloadReport(report workspace.ReloadSummary, inherit lipgloss.Style) string {
	added := Regular.Foreground(Green).Inherit(inherit).Render(fmt.Sprintf("+%d", len(report.Added)))
	removed := Regular.Foreground(Red).Inherit(inherit).Render(fmt.Sprintf("-%d", len(report.Removed)))

	return fmt.Sprintf("%s%s", added, removed)
}

// StateReloadReport renders a colored summary of changes resulting from a
// workspace reload.
func (h *Helpers) StateReloadReport(report state.ReloadSummary, inherit lipgloss.Style) string {
	var foreground lipgloss.TerminalColor
	switch report {
	case state.Empty:
		foreground = Red
	case state.Unchanged:
		foreground = Grey
	case state.Updated:
		foreground = Green
	}
	return Regular.Foreground(foreground).Inherit(inherit).Render(report.String())
}

// CostSummary renders a summary of the costs for a workspace.
func (h *Helpers) CostSummary(report workspace.CostSummary, inherit lipgloss.Style) string {
	return Regular.Foreground(Green).Inherit(inherit).Render(report.String())
}

// GroupReport renders a colored summary of a task group's task statuses.
func (h *Helpers) GroupReport(group *task.Group, table bool) string {
	var inherit lipgloss.Style
	if !table {
		inherit = Padded.Background(GroupReportBackgroundColor)
	}
	slash := Regular.Inherit(inherit).Foreground(Grey).Render("/")
	exited := Regular.Inherit(inherit).Foreground(Green).Render(fmt.Sprintf("%d", group.Exited()))
	total := Regular.Inherit(inherit).Foreground(Blue).Render(fmt.Sprintf("%d", len(group.Tasks)))

	s := fmt.Sprintf("%s%s%s", exited, slash, total)
	if errored := group.Errored(); errored > 0 {
		erroredString := Regular.Foreground(Red).Render(fmt.Sprintf("%d", errored))
		s = fmt.Sprintf("%s%s%s", erroredString, slash, s)
	}
	if table {
		return s
	}
	return Padded.Background(GroupReportBackgroundColor).Render(s)
}

// CreateTasks repeatedly invokes fn with each id in ids, creating a task for
// each invocation. If there is more than one id then a task group is created
// and the user sent to the task group's page; otherwise if only id is provided,
// the user is sent to the task's page.
func (h *Helpers) CreateTasks(fn task.SpecFunc, ids ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		switch len(ids) {
		case 0:
			return nil
		case 1:
			spec, err := fn(ids[0])
			if err != nil {
				return ErrorMsg(fmt.Errorf("creating task: %w", err))
			}
			task, err := h.Tasks.Create(spec)
			if err != nil {
				return ErrorMsg(fmt.Errorf("creating task: %w", err))
			}
			return NewNavigationMsg(TaskKind, WithParent(task.ID))
		default:
			specs := make([]task.Spec, 0, len(ids))
			for _, id := range ids {
				spec, err := fn(id)
				if err != nil {
					h.Logger.Error("creating task spec", "error", err, "id", id)
					continue
				}
				specs = append(specs, spec)
			}
			return h.createTaskGroup(specs...)
		}
	}
}

func (h *Helpers) CreateTasksWithSpecs(specs ...task.Spec) tea.Cmd {
	return func() tea.Msg {
		switch len(specs) {
		case 0:
			return nil
		case 1:
			task, err := h.Tasks.Create(specs[0])
			if err != nil {
				return ErrorMsg(fmt.Errorf("creating task: %w", err))
			}
			return NewNavigationMsg(TaskKind, WithParent(task.ID))
		default:
			return h.createTaskGroup(specs...)
		}
	}
}

func (h *Helpers) createTaskGroup(specs ...task.Spec) tea.Msg {
	group, err := h.Tasks.CreateGroup(specs...)
	if err != nil {
		return ReportError(fmt.Errorf("creating task group: %w", err))
	}
	return NewNavigationMsg(TaskGroupKind, WithParent(group.ID))
}

func (h *Helpers) Move(workspaceID resource.ID, from state.ResourceAddress) tea.Cmd {
	return CmdHandler(PromptMsg{
		Prompt:       "Enter destination address: ",
		InitialValue: string(from),
		Action: func(v string) tea.Cmd {
			if v == "" {
				return nil
			}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return h.States.Move(workspaceID, from, state.ResourceAddress(v))
			}
			return h.CreateTasks(fn, workspaceID)
		},
		Key:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	})
}
