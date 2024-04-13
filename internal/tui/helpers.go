package tui

import (
	"fmt"
	"strconv"
	"strings"

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
	modResource := res.Module()
	if modResource == nil {
		return ""
	}
	mod, err := h.ModuleService.Get(modResource.ID)
	if err != nil {
		h.Logger.Error("rendering module path", "error", err)
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
		h.Logger.Error("rendering workspace name", "error", err)
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
		h.Logger.Error("rendering current workspace name", "error", err)
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
	if err != nil {
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
	}
	return Regular.Copy().Foreground(color).Render(string(r.Status))
}

func (h *Helpers) LatestRunReport(r *run.Run) string {
	switch r.Status {
	case run.Planned, run.NoChanges:
		return h.RunReport(r.PlanReport)
	case run.Applied:
		return h.RunReport(r.ApplyReport)
	default:
		return "-"
	}
}

func (h *Helpers) RunReport(report run.Report) string {
	if !report.HasChanges() {
		return "-"
	}

	additions := Regular.Copy().Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := Regular.Copy().Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := Regular.Copy().Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

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
			h.Logger.Error("rendering workspace name", "error", err)
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
			h.Logger.Error("rendering module path", "error", err)
			break
		}
		path := Regular.Copy().Foreground(modulePathColor).Render(mod.Path)
		crumbs = append(crumbs, fmt.Sprintf("(%s)", path))
	case resource.Global:
		// if parented by global, then state it is global
		global := Regular.Copy().Foreground(globalColor).Render("all")
		crumbs = append(crumbs, fmt.Sprintf("(%s)", global))
	}
	return fmt.Sprintf("%s%s", TitleStyle.Render(title), strings.Join(crumbs, ""))
}

func GlobalBreadcrumb(title string) string {
	title = TitleStyle.Render(title)
	all := Regular.Copy().Foreground(globalColor).Render("all")
	return fmt.Sprintf("%s(%s)", title, all)
}
