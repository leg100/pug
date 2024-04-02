package tui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
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
	StateService     StateService
}

func (h *Helpers) ModulePath(res resource.Resource) string {
	modResource := res.Module()
	if modResource == nil {
		return ""
	}
	mod, err := h.ModuleService.Get(modResource.ID)
	if err != nil {
		slog.Error("rendering module path", "error", err)
		return ""
	}
	return mod.Path
}

func (h *Helpers) WorkspaceName(res resource.Resource) string {
	wsResource := res.Workspace()
	if wsResource == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(wsResource.ID)
	if err != nil {
		slog.Error("rendering workspace name", "error", err)
		return ""
	}
	return ws.Name
}

func (h *Helpers) CurrentWorkspaceName(workspaceID *resource.ID) string {
	if workspaceID == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(*workspaceID)
	if err != nil {
		slog.Error("rendering current workspace name", "error", err)
		return ""
	}
	return ws.Name
}

func (h *Helpers) ModuleCurrentRunStatus(mod *module.Module) string {
	if mod.CurrentWorkspaceID == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		slog.Error("rendering module current run status", "error", err)
		return ""
	}
	return h.WorkspaceCurrentRunStatus(ws)
}

func (h *Helpers) ModuleCurrentRunChanges(mod *module.Module, inherit lipgloss.Style) string {
	if mod.CurrentWorkspaceID == nil {
		return ""
	}
	ws, err := h.WorkspaceService.Get(*mod.CurrentWorkspaceID)
	if err != nil {
		slog.Error("rendering module current run changes", "error", err)
		return ""
	}
	return h.WorkspaceCurrentRunChanges(ws, inherit)
}

func (h *Helpers) WorkspaceCurrentRunStatus(ws *workspace.Workspace) string {
	if ws.CurrentRunID == nil {
		return ""
	}
	run, err := h.RunService.Get(*ws.CurrentRunID)
	if err != nil {
		slog.Error("rendering module current run status", "error", err)
		return ""
	}
	return h.RunStatus(run)
}

func (h *Helpers) WorkspaceCurrentRunChanges(ws *workspace.Workspace, inherit lipgloss.Style) string {
	if ws.CurrentRunID == nil {
		return ""
	}
	run, err := h.RunService.Get(*ws.CurrentRunID)
	if err != nil {
		slog.Error("rendering module current run changes", "error", err)
		return ""
	}
	return h.LatestRunReport(run, inherit)
}

// WorkspaceCurrentCheckmark returns a check mark if the workspace is the
// current workspace for its module.
func (h *Helpers) WorkspaceCurrentCheckmark(ws *workspace.Workspace, inherit lipgloss.Style) string {
	mod, err := h.ModuleService.Get(ws.ModuleID())
	if err != nil {
		slog.Error("rendering current workspace checkmark", "error", err)
		return ""
	}
	if mod.CurrentWorkspaceID != nil && *mod.CurrentWorkspaceID == ws.ID {
		return "✓"
	}
	return ""
}

func (h *Helpers) WorkspaceResourceCount(ws *workspace.Workspace) string {
	state, err := h.StateService.Get(ws.ID)
	if err != nil {
		slog.Error("rendering workspace resource count", "error", err)
		return ""
	}
	return strconv.Itoa(len(state.Resources))
}

func (h *Helpers) RunStatus(r *run.Run) string {
	var color lipgloss.Color

	switch r.Status {
	case run.Pending:
		color = Grey
	case run.PlanQueued:
		color = Orange
	case run.Planning:
		color = Blue
	case run.Planned:
		color = DeepBlue
	case run.PlannedAndFinished:
		color = GreenBlue
	case run.Applied:
		color = Black
	case run.Errored:
		color = Red
	}
	return Regular.Copy().Foreground(color).Render(string(r.Status))
}

func (h *Helpers) LatestRunReport(r *run.Run, inherit lipgloss.Style) string {
	switch r.Status {
	case run.Planned, run.PlannedAndFinished:
		return h.RunReport(r.PlanReport, inherit)
	case run.Applied:
		return h.RunReport(r.ApplyReport, inherit)
	default:
		return "-"
	}
}

func (h *Helpers) RunReport(report run.Report, inherit lipgloss.Style) string {
	if !report.HasChanges() {
		return "no changes"
	}

	inherit = Regular.Copy().Inherit(inherit)

	additions := inherit.Copy().Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := inherit.Copy().Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := inherit.Copy().Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

	return fmt.Sprintf("%s%s%s", additions, changes, destructions)
}

func (h *Helpers) Breadcrumbs(title string, parent resource.Resource) string {
	// format: title[workspace name](module path)
	var crumbs []string
	switch parent.Kind {
	case resource.Run:
		// get parent workspace
		parent = *parent.Parent
		fallthrough
	case resource.Workspace:
		ws, err := h.WorkspaceService.Get(parent.ID)
		if err != nil {
			slog.Error("rendering workspace name", "error", err)
			break
		}
		name := Regular.Copy().Foreground(Red).Render(ws.Name)
		crumbs = append(crumbs, fmt.Sprintf("[%s]", name))
		// now get parent of workspace which is module
		parent = *parent.Parent
		fallthrough
	case resource.Module:
		mod, err := h.ModuleService.Get(parent.ID)
		if err != nil {
			slog.Error("rendering module path", "error", err)
			break
		}
		path := Regular.Copy().Foreground(Blue).Render(mod.Path)
		crumbs = append(crumbs, fmt.Sprintf("(%s)", path))
	case resource.Global:
		// if parented by global, then state it is global
		global := Regular.Copy().Foreground(Blue).Render("all")
		crumbs = append(crumbs, fmt.Sprintf("(%s)", global))
	}
	return fmt.Sprintf("%s%s", Bold.Render(title), strings.Join(crumbs, ""))
}

func GlobalBreadcrumb(title string) string {
	title = Bold.Render(title)
	all := Regular.Copy().Foreground(Blue).Render("all")
	return fmt.Sprintf("%s(%s)", title, all)
}
