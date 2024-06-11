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
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

var TitleStyle = Bold.Copy().Foreground(TitleColor)

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
func (h *Helpers) TaskWorkspace(t *task.Task) resource.Resource {
	if ws := t.Workspace(); ws != nil {
		return ws
	}
	if mod := h.Module(t); mod != nil {
		return h.ModuleCurrentWorkspace(mod)
	}
	return nil
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
		return Regular.Copy().Padding(0, 1).Background(color).Foreground(White).Render(string(t.State))
	} else {
		return Regular.Copy().Foreground(color).Render(string(t.State))
	}
}

func (h *Helpers) RunStatus(r *run.Run, background bool) string {
	var color lipgloss.TerminalColor

	switch r.Status {
	case run.Pending:
		color = Grey
	case run.PlanQueued:
		color = Orange
	case run.Planning:
		color = Blue
	case run.Planned:
		color = DeepBlue
	case run.NoChanges:
		color = GreenBlue
	case run.Applying:
		color = lipgloss.AdaptiveColor{
			Light: string(DarkGreen),
			Dark:  string(LightGreen),
		}
	case run.Applied:
		color = Green
	case run.Errored:
		color = Red
	}

	if background {
		return Regular.Copy().Background(color).Padding(0, 1).Foreground(White).Render(string(r.Status))
	} else {
		return Regular.Copy().Foreground(color).Render(string(r.Status))
	}
}

func (h *Helpers) LatestRunReport(r *run.Run) string {
	if r.ApplyReport != nil {
		return h.RunReport(*r.ApplyReport)
	}
	if r.PlanReport != nil {
		return h.RunReport(*r.PlanReport)
	}
	return ""
}

func (h *Helpers) RunReport(report run.Report) string {
	additions := Regular.Copy().Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := Regular.Copy().Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := Regular.Copy().Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

	return fmt.Sprintf("%s%s%s", additions, changes, destructions)
}

func (h *Helpers) Breadcrumbs(title string, res resource.Resource, crumbs ...string) string {
	// format: title{task command}[workspace name](module path)
	switch res.GetKind() {
	case resource.Task:
		cmd := Regular.Copy().Foreground(Blue).Render(res.String())
		crumb := fmt.Sprintf("{%s}", cmd)
		return h.Breadcrumbs(title, res.GetParent(), crumb)
	case resource.StateResource:
		addr := Regular.Copy().Foreground(Blue).Render(res.String())
		crumb := fmt.Sprintf("{%s}", addr)
		return h.Breadcrumbs(title, res.GetParent(), crumb)
	case resource.TaskGroup:
		cmd := Regular.Copy().Foreground(Blue).Render(res.String())
		id := Regular.Copy().Foreground(Green).Render(res.GetID().String())
		crumb := fmt.Sprintf("{%s}[%s]", cmd, id)
		return h.Breadcrumbs(title, res.GetParent(), crumb)
	case resource.Run:
		// Skip run info in breadcrumbs
		return h.Breadcrumbs(title, res.GetParent(), crumbs...)
	case resource.Workspace:
		crumb := fmt.Sprintf("[%s]", Regular.Copy().Foreground(Red).Render(res.String()))
		return h.Breadcrumbs(title, res.GetParent(), append(crumbs, crumb)...)
	case resource.Module:
		path := Regular.Copy().Foreground(modulePathColor).Render(res.String())
		crumbs = append(crumbs, fmt.Sprintf("(%s)", path))
	}
	return fmt.Sprintf("%s%s", TitleStyle.Render(title), strings.Join(crumbs, ""))
}

func (h *Helpers) CreateTasks(cmd string, fn task.Func, ids ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		switch len(ids) {
		case 0:
			return nil
		case 1:
			task, err := fn(ids[0])
			if err != nil {
				return ReportError(err, "creating task")
			}
			return NewNavigationMsg(TaskKind, WithParent(task))
		default:
			group, err := h.TaskService.CreateGroup(cmd, fn, ids...)
			if err != nil {
				return ReportError(err, "creating task group")
			}
			return NewNavigationMsg(TaskGroupKind, WithParent(group))
		}
	}
}

func GlobalBreadcrumb(title, total string) string {
	title = TitleStyle.Render(title)
	return fmt.Sprintf("%s[%s]", title, total)
}

// RemoveDuplicateBindings removes duplicate bindings from a list of bindings. A
// binding is deemed a duplicate if another binding has the same list of keys.
func RemoveDuplicateBindings(bindings []key.Binding) []key.Binding {
	seen := make(map[string]struct{})
	var i int
	for _, b := range bindings {
		key := strings.Join(b.Keys(), " ")
		if _, ok := seen[key]; ok {
			// duplicate, skip
			continue
		}
		seen[key] = struct{}{}
		bindings[i] = b
		i++
	}
	return bindings[:i]
}

func BoolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
