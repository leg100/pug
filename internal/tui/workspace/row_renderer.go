package workspace

import (
	"log/slog"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/table"
	"github.com/leg100/pug/internal/workspace"
)

type rowRenderer struct {
	ModuleService *module.Service
	RunService    *run.Service
	parent        resource.Resource
}

func (rr *rowRenderer) renderRow(ws *workspace.Workspace, inherit lipgloss.Style) table.RenderedRow {
	row := table.RenderedRow{
		table.WorkspaceColumn.Key: ws.Name,
	}

	if cr := ws.CurrentRunID; cr != nil {
		run, _ := rr.RunService.Get(*cr)
		row[table.RunStatusColumn.Key] = tui.RenderRunStatus(run.Status)
		row[table.RunChangesColumn.Key] = tui.RenderLatestRunReport(run, inherit)
	}

	mod, err := rr.ModuleService.Get(ws.ModuleID())
	if err != nil {
		slog.Error("rendering workspace row", "error", err)
		return row
	}

	// Show module path in global workspaces table
	if rr.parent.Kind == resource.Global {
		row[table.ModuleColumn.Key] = mod.Path
	}

	if mod.CurrentWorkspaceID != nil && *mod.CurrentWorkspaceID == ws.ID {
		row[currentColumn.Key] = "✓"
	}

	return row
}
