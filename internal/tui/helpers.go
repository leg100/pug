package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
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

func (h *Helpers) ModuleCurrentRunStatus(mod *module.Module) string {
	if mod.CurrentWorkspaceID == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("rendering module current run status", "error", err)
		return ""
	}
	return h.WorkspaceCurrentRunStatus(ws)
}

func (h *Helpers) ModuleCurrentRunChanges(mod *module.Module) string {
	if mod.CurrentWorkspaceID == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		h.Logger.Error("rendering module current run changes", "error", err)
		return ""
	}
	return h.WorkspaceCurrentRunChanges(ws)
}

func (h *Helpers) WorkspaceCurrentRunStatus(ws *workspace.Workspace) string {
	if ws.CurrentRunID == nil {
		return ""
	}
	run, err := h.RunService.Get(*ws.CurrentRunID)
	if err != nil {
		h.Logger.Error("rendering module current run status", "error", err)
		return ""
	}
	return h.RunStatus(run)
}

func (h *Helpers) WorkspaceCurrentRunChanges(ws *workspace.Workspace) string {
	if ws.CurrentRunID == nil {
		return ""
	}
	run, err := h.RunService.Get(*ws.CurrentRunID)
	if err != nil {
		h.Logger.Error("rendering module current run changes", "error", err)
		return ""
	}
	return h.LatestRunReport(run)
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

func (h *Helpers) RunStatus(r *run.Run) string {
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
	case run.Stale:
		color = Orange
	}
	return Regular.Copy().Foreground(color).Render(string(r.Status))
}

func (h *Helpers) LatestRunReport(r *run.Run) string {
	switch r.Status {
	case run.Planned, run.NoChanges, run.Stale:
		return h.RunReport(r.PlanReport)
	case run.Applied:
		return h.RunReport(r.ApplyReport)
	default:
		return "-"
	}
}

func (h *Helpers) RunReport(report run.Report) string {
	additions := Regular.Copy().Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := Regular.Copy().Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := Regular.Copy().Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

	return fmt.Sprintf("%s%s%s", additions, changes, destructions)
}

func Breadcrumbs(title string, res resource.Resource, suffix string, crumbs ...string) string {
	// format: title{task command}[workspace name](module path)suffix
	switch res.GetKind() {
	case resource.Task:
		cmd := Regular.Copy().Foreground(Green).Render(res.String())
		crumb := fmt.Sprintf("{%s}", cmd)
		return Breadcrumbs(title, res.GetParent(), suffix, crumb)
	case resource.Run:
		id := Regular.Copy().Foreground(Blue).Render(res.String())
		crumbs = append(crumbs, fmt.Sprintf("{%s}", id))
		return Breadcrumbs(title, res.GetParent(), suffix, crumbs...)
	case resource.Workspace:
		name := Regular.Copy().Foreground(Red).Render(res.String())
		crumbs = append(crumbs, fmt.Sprintf("[%s]", name))
		return Breadcrumbs(title, res.GetParent(), suffix, crumbs...)
	case resource.Module:
		path := Regular.Copy().Foreground(modulePathColor).Render(res.String())
		crumbs = append(crumbs, fmt.Sprintf("(%s)", path))
	case resource.Global:
		all := Regular.Copy().Foreground(globalColor).Render("all")
		crumbs = []string{fmt.Sprintf("(%s)", all)}
	}
	return fmt.Sprintf("%s%s%s", TitleStyle.Render(title), strings.Join(crumbs, ""), suffix)
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
