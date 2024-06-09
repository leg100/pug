package task

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leg100/pug/internal/resource"
	runpkg "github.com/leg100/pug/internal/run"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
	"github.com/leg100/pug/internal/tui/table"
)

const (
	// default height of the top list pane, not including borders
	defaultListPaneHeight = 10
	// total width of borders to the left and right of a pane
	totalPaneBorderWidth = 2
	// total height of borders above and below a pane
	totalPaneBorderHeight = 2
)

var (
	commandColumn = table.Column{
		Key:        "command",
		Title:      "COMMAND",
		FlexFactor: 1,
	}
	statusColumn = table.Column{
		Key:   "task_status",
		Title: "STATUS",
		Width: task.MaxStatusLen,
	}
	ageColumn = table.Column{
		Key:   "age",
		Title: "AGE",
		Width: 7,
	}
	runChangesColumn = table.Column{
		Key:        "run_changes",
		Title:      "RUN CHANGES",
		FlexFactor: 1,
	}
	runStatusColumn = table.Column{
		Key:   "run_status",
		Title: "RUN STATUS",
		Width: runpkg.MaxStatusLen,
	}
)

type listPreviewOptions struct {
	width             int
	height            int
	runService        tui.RunService
	taskService       tui.TaskService
	helpers           *tui.Helpers
	taskMaker         *Maker
	taskMakerID       MakerID
	hideCommandColumn bool
}

func newListPreview(opts listPreviewOptions) *listPreview {
	columns := []table.Column{
		table.ModuleColumn,
		table.WorkspaceColumn,
	}
	if opts.hideCommandColumn {
		columns = append(columns, commandColumn)
	}
	columns = append(columns,
		statusColumn,
		runStatusColumn,
		runChangesColumn,
		ageColumn,
	)

	renderer := func(t *task.Task) table.RenderedRow {
		row := table.RenderedRow{
			table.ModuleColumn.Key:    opts.helpers.ModulePath(t),
			table.WorkspaceColumn.Key: opts.helpers.WorkspaceName(t),
			commandColumn.Key:         t.CommandString(),
			ageColumn.Key:             tui.Ago(time.Now(), t.Updated),
			table.IDColumn.Key:        t.String(),
			statusColumn.Key:          opts.helpers.TaskStatus(t, false),
		}

		if rr := t.Run(); rr != nil {
			run := rr.(*runpkg.Run)
			if t.CommandString() == "plan" && run.PlanReport != nil {
				row[runChangesColumn.Key] = opts.helpers.RunReport(*run.PlanReport)
			} else if t.CommandString() == "apply" && run.ApplyReport != nil {
				row[runChangesColumn.Key] = opts.helpers.RunReport(*run.ApplyReport)
			}
			row[runStatusColumn.Key] = opts.helpers.RunStatus(run, false)
		}

		return row
	}

	lp := listPreview{
		tasks:          opts.taskService,
		runs:           opts.runService,
		width:          opts.width,
		height:         opts.height,
		helpers:        opts.helpers,
		taskMaker:      opts.taskMaker,
		taskMakerID:    opts.taskMakerID,
		cache:          tui.NewCache(),
		previewVisible: true,
	}

	// Create table for the top list pane
	lp.list = table.New(
		columns,
		renderer,
		lp.paneWidth(),
		lp.listHeight(),
		table.WithSortFunc(task.ByState),
	)

	return &lp
}

// listPreview is a composition of two panes: a top pane is a list of tasks;
// the bottom pane is the output of the currently highlighted task in the list,
// i.e. a preview.
type listPreview struct {
	list        table.Model[resource.ID, *task.Task]
	tasks       tui.TaskService
	runs        tui.RunService
	taskMaker   *Maker
	taskMakerID MakerID

	previewVisible bool
	previewFocused bool
	height         int
	width          int
	// userListHeightAdjustment is the adjustment the user has requested to the
	// default height of the list pane.
	userListHeightAdjustment int

	// map of task ID to task model
	cache   *tui.Cache
	helpers *tui.Helpers

	n *int
}

func (m *listPreview) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keys applicable when either pane is focused
		switch {
		case key.Matches(msg, keys.Navigation.SwitchPane):
			m.previewFocused = !m.previewFocused
		case key.Matches(msg, localKeys.TogglePreview):
			m.previewVisible = !m.previewVisible
			m.recalculateDimensions()
		case key.Matches(msg, localKeys.GrowPreview):
			// Grow the preview pane by shrinking the list pane
			m.userListHeightAdjustment--
			m.recalculateDimensions()
		case key.Matches(msg, localKeys.ShrinkPreview):
			// Shrink the preview pane by growing the list pane
			m.userListHeightAdjustment++
			m.recalculateDimensions()
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.list.CurrentRow(); ok {
				return tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		}
		if m.previewVisible && m.previewFocused {
			// Preview pane is visible and focused, so send keys to the task
			// model for the currently highlighted table row if there is one.
			row, ok := m.list.CurrentRow()
			if !ok {
				break
			}
			page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
			cmd := m.cache.Update(tui.NewCacheKey(page), msg)
			cmds = append(cmds, cmd)
		} else {
			// Table pane is focused, so handle keys relevant to table rows.
			//
			// TODO: when preview is focused, we also want these keys to be
			// handled for the current row (but not selected rows).
			switch {
			case key.Matches(msg, keys.Common.Cancel):
				taskIDs := m.list.SelectedOrCurrentKeys()
				return m.helpers.CreateTasks("cancel", m.tasks.Cancel, taskIDs...)
			case key.Matches(msg, keys.Common.Apply):
				runIDs, err := m.pruneApplyableTasks()
				if err != nil {
					return tui.ReportError(err, "applying tasks")
				}
				return tui.YesNoPrompt(
					fmt.Sprintf("Apply %d plans?", len(runIDs)),
					m.helpers.CreateTasks("apply", m.runs.ApplyPlan, runIDs...),
				)
			default:
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.recalculateDimensions()
	default:
		// Forward remaining message types to both the table model and cached
		// task models
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.cache.UpdateAll(msg)...)
	}

	if m.previewVisible {
		// Get currently highlighted task and ensure a model exists for it, and
		// ensure that that model is the current model.
		if row, ok := m.list.CurrentRow(); ok {
			page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
			if !m.cache.Exists(page) {
				// Create model
				model, err := m.taskMaker.makeWithID(row.Value, m.paneWidth(), m.previewHeight(), m.taskMakerID)
				if err != nil {
					return tui.ReportError(err, "making task model")
				}
				// Cache newly created model
				m.cache.Put(page, model)
				// Initialize model
				cmds = append(cmds, model.Init())
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *listPreview) recalculateDimensions() {
	m.list, _ = m.list.Update(tea.WindowSizeMsg{
		Height: m.listHeight(),
		Width:  m.paneWidth(),
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.paneWidth(),
	})
}

func (m listPreview) paneWidth() int {
	return m.width - totalPaneBorderWidth
}

func (m listPreview) listHeight() int {
	if m.previewVisible {
		// Ensure list pane is at least a height of 2 (the headings and one row)
		return max(2, defaultListPaneHeight+m.userListHeightAdjustment)
	}
	return m.height - totalPaneBorderHeight
}

func (m listPreview) previewHeight() int {
	// calculate height of preview pane after accounting for:
	// (a) height of list pane above
	// (b) height of borders above and below both panes
	return max(0, m.height-m.listHeight()-(totalPaneBorderHeight*2))
}

// pruneApplyableTasks removes from the selection any tasks that cannot be
// applied, i.e all tasks other than those that are a plan and are in the
// planned state. The run ID of each task after pruning is returned.
func (m *listPreview) pruneApplyableTasks() ([]resource.ID, error) {
	return m.list.Prune(func(task *task.Task) (resource.ID, bool) {
		rr := task.Run()
		if rr == nil {
			return resource.ID{}, true
		}
		run := rr.(*runpkg.Run)
		if run.Status != runpkg.Planned {
			return resource.ID{}, true
		}
		return run.ID, false
	})
}

var (
	singlePaneBorder   = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	activePaneBorder   = lipgloss.NewStyle().Border(lipgloss.ThickBorder())
	inactivePaneBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(tui.LighterGrey)
)

func (m listPreview) View() string {
	var (
		tableBorder   lipgloss.Style
		previewBorder lipgloss.Style
	)
	if !m.previewVisible {
		tableBorder = singlePaneBorder
	} else if m.previewFocused {
		tableBorder = inactivePaneBorder
		previewBorder = activePaneBorder
	} else {
		tableBorder = activePaneBorder
		previewBorder = inactivePaneBorder
	}
	components := []string{
		tableBorder.Render(m.list.View()),
	}
	// When preview pane is visible and there is a task model cached for the
	// current row, then render the task's output in the pane.
	if m.previewVisible {
		if model, ok := m.currentTaskModel(); ok {
			components = append(components, previewBorder.Render(model.View()))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

func (m listPreview) currentTaskModel() (tea.Model, bool) {
	row, ok := m.list.CurrentRow()
	if !ok {
		return nil, false
	}
	page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
	model := m.cache.Get(page)
	return model, model != nil
}
