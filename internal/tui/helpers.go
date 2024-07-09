package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
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
	ModuleService    ModuleService
	WorkspaceService WorkspaceService
	RunService       RunService
	TaskService      TaskService
	StateService     StateService
	Logger           logging.Interface
}

func (h *Helpers) ModulePath(res resource.Resource) string {
	if mod := res.Module(); mod != nil {
		return mod.String()
	}
	return ""
}

func (h *Helpers) WorkspaceName(res resource.Resource) string {
	if ws := res.Workspace(); ws != nil {
		return ws.String()
	}
	return ""
}

func (h *Helpers) ModuleCurrentWorkspace(mod *module.Module) *workspace.Workspace {
	if mod.CurrentWorkspaceID == nil {
		h.Logger.Error("module does not have a current workspace", "module", mod)
		return nil
	}
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("retrieving current workspace for module", "error", err, "module", mod)
		return nil
	}
	return ws
}

func (h *Helpers) Module(res resource.Resource) *module.Module {
	if res.Module() == nil {
		return nil
	}
	mod, ok := res.Module().(*module.Module)
	if !ok {
		h.Logger.Error("unable to unwrap module from resource interface", "resource", res)
		return nil
	}
	return mod
}

func (h *Helpers) CurrentWorkspaceName(workspaceID *resource.ID) string {
	if workspaceID == nil {
		return "-"
	}
	ws, err := h.WorkspaceService.Get(*workspaceID)
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
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("rendering module current workspace resource count", "error", err)
		return ""
	}
	return h.WorkspaceResourceCount(ws)
}

// WorkspaceCurrentCheckmark returns a check mark if the workspace is the
// current workspace for its module.
func (h *Helpers) WorkspaceCurrentCheckmark(ws *workspace.Workspace) string {
	mod, err := h.ModuleService.Get(ws.ModuleID())
	if err != nil {
		h.Logger.Error("rendering current workspace checkmark", "error", err)
		return ""
	}
	if mod.CurrentWorkspaceID != nil && *mod.CurrentWorkspaceID == ws.ID {
		return "✓"
	}
	return ""
}

func (h *Helpers) WorkspaceResourceCount(ws *workspace.Workspace) string {
	state, err := h.StateService.Get(ws.ID)
	if errors.Is(err, resource.ErrNotFound) {
		// not found most likely means state not loaded yet
		return ""
	} else if err != nil {
		h.Logger.Error("rendering workspace resource count", "error", err)
		return ""
	}
	return strconv.Itoa(len(state.Resources))
}

// TaskWorkspace retrieves either the task's workspace if it belongs to a
// workspace, or if it belongs to a module, then it retrieves the module's
// current workspace
func (h *Helpers) TaskWorkspace(t *task.Task) (resource.Resource, bool) {
	if ws := t.Workspace(); ws != nil {
		return ws, true
	}
	if mod := h.Module(t); mod != nil {
		if ws := h.ModuleCurrentWorkspace(mod); ws != nil {
			return ws, true
		}
		return nil, false
	}
	return nil, false
}

// TaskStatus provides a rendered colored task status.
func (h *Helpers) TaskStatus(t *task.Task, background bool) string {
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

	if background {
		return Padded.Background(color).Foreground(White).Render(string(t.State))
	} else {
		return Regular.Foreground(color).Render(string(t.State))
	}
}

func (h *Helpers) LatestRunReport(r *run.Run, table bool) string {
	if r.ApplyReport != nil {
		return h.RunReport(*r.ApplyReport, table)
	}
	if r.PlanReport != nil {
		return h.RunReport(*r.PlanReport, table)
	}
	return ""
}

// RunReport renders a colored summary of a run's changes. Set table to true if
// the report is rendered within a table row.
func (h *Helpers) RunReport(report run.Report, table bool) string {
	var background lipgloss.TerminalColor = lipgloss.NoColor{}
	if !table {
		background = RunReportBackgroundColor
	}
	additions := Regular.Background(background).Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := Regular.Background(background).Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := Regular.Background(background).Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

	s := fmt.Sprintf("%s%s%s", additions, changes, destructions)
	if !table {
		s = Padded.Background(background).Render(s)
	}
	return s
}

// RunReport renders a colored summary of a run's changes. Set table to true if
// the report is rendered within a table row.
func (h *Helpers) GroupReport(group *task.Group, table bool) string {
	var background lipgloss.TerminalColor = lipgloss.NoColor{}
	if !table {
		background = GroupReportBackgroundColor
	}
	slash := Regular.Background(background).Foreground(Black).Render("/")
	exited := Regular.Background(background).Foreground(Green).Render(fmt.Sprintf("%d", group.Exited()))
	total := Regular.Background(background).Foreground(Black).Render(fmt.Sprintf("%d", len(group.Tasks)))

	s := fmt.Sprintf("%s%s%s", exited, slash, total)
	if errored := group.Errored(); errored > 0 {
		erroredString := Regular.Background(background).Foreground(Red).Render(fmt.Sprintf("%d", errored))
		s = fmt.Sprintf("%s%s%s", erroredString, slash, s)
	}

	if !table {
		s = Padded.Background(background).Render(s)
	}
	return s
}

func (h *Helpers) CreateTasks(cmd string, fn task.Func, ids ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		switch len(ids) {
		case 0:
			return nil
		case 1:
			task, err := fn(ids[0])
			if err != nil {
				return ReportError(fmt.Errorf("creating task: %w", err))
			}
			return NewNavigationMsg(TaskKind, WithParent(task))
		default:
			group, err := h.TaskService.CreateGroup(cmd, fn, ids...)
			if err != nil {
				return ReportError(fmt.Errorf("creating task group: %w", err))
			}
			return NewNavigationMsg(TaskGroupKind, WithParent(group))
		}
	}
}

func (h *Helpers) CreateApplyTasks(opts *run.CreateOptions, ids ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		switch len(ids) {
		case 0:
			return nil
		case 1:
			// Only one task is to be created. If successful send user directly to task
			// page. Otherwise report an error.
			task, err := h.RunService.Apply(ids[0], opts)
			if err != nil {
				return ReportError(fmt.Errorf("creating apply task: %w", err))
			}
			return NewNavigationMsg(TaskKind, WithParent(task))
		default:
			// More than one task is to be created. If successful send user to
			// task group page.
			group, err := h.RunService.MultiApply(opts, ids...)
			if err != nil {
				return ReportError(fmt.Errorf("creating apply task group: %w", err))
			}
			return NewNavigationMsg(TaskGroupKind, WithParent(group))
		}
	}
}

func (h *Helpers) Move(workspaceID resource.ID, from state.ResourceAddress) tea.Cmd {
	return CmdHandler(PromptMsg{
		Prompt:       "Enter destination address: ",
		InitialValue: string(from),
		Action: func(v string) tea.Cmd {
			if v == "" {
				return nil
			}
			fn := func(workspaceID resource.ID) (*task.Task, error) {
				return h.StateService.Move(workspaceID, from, state.ResourceAddress(v))
			}
			return h.CreateTasks("state-mv", fn, workspaceID)
		},
		Key:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	})
}

func Breadcrumbs(title string, res resource.Resource, crumbs ...string) string {
	// format: title{task command}[workspace name](module path)
	switch res := res.(type) {
	case *task.Task:
		cmd := TitleCommand.Render(res.String())
		return Breadcrumbs(title, res.GetParent(), cmd)
	case *state.Resource:
		addr := TitleAddress.Render(res.String())
		return Breadcrumbs(title, res.GetParent().GetParent(), addr)
	case *task.Group:
		cmd := TitleCommand.Render(res.String())
		id := TitleID.Render(res.GetID().String())
		return Breadcrumbs(title, res.GetParent(), cmd, id)
	case *run.Run:
		// Skip run info in breadcrumbs
		return Breadcrumbs(title, res.GetParent(), crumbs...)
	case *workspace.Workspace:
		name := TitleWorkspace.Render(res.String())
		return Breadcrumbs(title, res.GetParent(), append(crumbs, name)...)
	case *module.Module:
		crumbs = append(crumbs, TitlePath.Render(res.String()))
	}
	return fmt.Sprintf("%s%s", Title.Render(title), strings.Join(crumbs, ""))
}
