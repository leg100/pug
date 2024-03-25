package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/run"
)

func RenderRunStatus(status run.Status) string {
	var color lipgloss.Color

	switch status {
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
	return Regular.Copy().Foreground(color).Render(string(status))
}

func RenderLatestRunReport(r *run.Run, inherit lipgloss.Style) string {
	switch r.Status {
	case run.Planned, run.PlannedAndFinished:
		return RenderRunReport(r.PlanReport, inherit)
	case run.Applied:
		return RenderRunReport(r.ApplyReport, inherit)
	default:
		return "-"
	}
}

func RenderRunReport(report run.Report, inherit lipgloss.Style) string {
	if !report.HasChanges() {
		return "no changes"
	}

	inherit = Regular.Copy().Inherit(inherit)

	additions := inherit.Copy().Foreground(Green).Render(fmt.Sprintf("+%d", report.Additions))
	changes := inherit.Copy().Foreground(Blue).Render(fmt.Sprintf("~%d", report.Changes))
	destructions := inherit.Copy().Foreground(Red).Render(fmt.Sprintf("-%d", report.Destructions))

	return fmt.Sprintf("%s%s%s", additions, changes, destructions)
}
