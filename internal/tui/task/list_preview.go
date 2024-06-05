package task

import (
	"errors"
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
	// height of the top list panel, not including borders
	listPanelHeight = 10
	// total width of borders to the left and right of a panel
	totalPanelBorderWidth = 2
	// total height of borders above and below a panel
	totalPanelBorderHeight = 2
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
	parent      resource.Resource
	width       int
	height      int
	runService  tui.RunService
	taskService tui.TaskService
	helpers     *tui.Helpers
}

func newListPreview(opts listPreviewOptions) ListPreview {
	columns := []table.Column{
		table.ModuleColumn,
		table.WorkspaceColumn,
	}
	// Don't show command column in a task group list because all its tasks
	// share the same command and the command is already included in the title.
	if opts.parent.GetKind() != resource.TaskGroup {
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

	lp := ListPreview{
		tasks:   opts.taskService,
		runs:    opts.runService,
		width:   opts.width,
		height:  opts.height,
		helpers: opts.helpers,
		taskMaker: &Maker{
			MakerID:     TaskListPreviewMakerID,
			RunService:  opts.runService,
			TaskService: opts.taskService,
			Helpers:     opts.helpers,
		},
		cache: tui.NewCache(),
	}

	// Create table for the top list panel
	lp.list = table.New(columns, renderer, lp.panelWidth(), lp.listHeight()).
		WithSortFunc(task.ByState).
		WithParent(opts.parent)

	return lp
}

// ListPreview is a composition of two panes: a top panel is a list of tasks;
// the bottom panel is the output of the currently highlighted task in the list,
// i.e. a preview.
type ListPreview struct {
	list      table.Model[resource.ID, *task.Task]
	tasks     tui.TaskService
	runs      tui.RunService
	taskMaker tui.Maker

	previewVisible bool
	previewFocused bool
	height         int
	width          int
	// map of task ID to task model
	cache   *tui.Cache
	helpers *tui.Helpers
	// currently highlighted task
	current tui.Page

	tea.Model
}

func (m ListPreview) Update(msg tea.Msg) (ListPreview, tea.Cmd) {
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
			return m, nil
		case key.Matches(msg, localKeys.TogglePreview):
			m.previewVisible = !m.previewVisible
			m.recalculateDimensions()
			return m, nil
		case key.Matches(msg, keys.Global.Enter):
			if row, ok := m.list.CurrentRow(); ok {
				return m, tui.NavigateTo(tui.TaskKind, tui.WithParent(row.Value))
			}
		}
		if m.previewVisible && m.previewFocused {
			// Preview pane is visible and focused, so send keys to the task
			// model for the currently highlighted table row if there is one.
			row, ok := m.list.CurrentRow()
			if !ok {
				return m, nil
			}
			page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
			cmd := m.cache.Update(tui.NewCacheKey(page), msg)
			return m, cmd
		} else {
			// Table pane is focused, so handle keys relevant to table rows.
			//
			// TODO: when preview is focused, we also want these keys to be
			// handled for the current row (but not selected rows).
			switch {
			case key.Matches(msg, keys.Common.Cancel):
				taskIDs := m.list.SelectedOrCurrentKeys()
				return m, m.helpers.CreateTasks("cancel", m.tasks.Cancel, taskIDs...)
			case key.Matches(msg, keys.Common.Apply):
				runIDs, err := m.pruneApplyableTasks()
				if err != nil {
					return m, tui.ReportError(err, "")
				}
				return m, tui.YesNoPrompt(
					fmt.Sprintf("Apply %d plans?", len(runIDs)),
					m.helpers.CreateTasks("apply", m.runs.ApplyPlan, runIDs...),
				)
			default:
				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)
				// Key might have changed the current table row, so don't return
				// yet, and let the code below check a task model exists for the
				// row
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.recalculateDimensions()
		return m, cmd
	default:
		// Forward remaining message types to both the table model and cached
		// task models
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.cache.UpdateAll(msg)...)
	}

	// Get currently highlighted task and ensure a model exists for it, and
	// ensure that that model is the current model.
	if row, ok := m.list.CurrentRow(); ok {
		page := tui.Page{Kind: tui.TaskKind, Resource: row.Value}
		if !m.cache.Exists(page) {
			// Create model
			model, err := m.taskMaker.Make(row.Value, m.panelWidth(), m.previewHeight())
			if err != nil {
				return m, tui.ReportError(err, "making task model")
			}
			// Cache newly created model
			m.cache.Put(page, model)
			// Initialize model
			cmds = append(cmds, model.Init())
		}
		m.current = page
	}

	return m, tea.Batch(cmds...)
}

func (m *ListPreview) recalculateDimensions() {
	m.list, _ = m.list.Update(tea.WindowSizeMsg{
		Height: m.listHeight(),
		Width:  m.panelWidth(),
	})
	_ = m.cache.UpdateAll(tea.WindowSizeMsg{
		Height: m.previewHeight(),
		Width:  m.panelWidth(),
	})
}

func (m ListPreview) panelWidth() int {
	return m.width - totalPanelBorderWidth
}

func (m ListPreview) listHeight() int {
	if m.previewVisible {
		return listPanelHeight
	}
	return m.height - totalPanelBorderHeight
}

func (m ListPreview) previewHeight() int {
	// calculate height of preview pane after accounting for:
	// (a) height of list panel above
	// (b) height of borders above and below both panels
	return m.height - listPanelHeight - (totalPanelBorderHeight * 2)
}

// pruneApplyableTasks removes from the selection any tasks that cannot be
// applied, i.e all tasks other than those that are a plan and are in the
// planned state. The run ID of each task after pruning is returned.
func (m ListPreview) pruneApplyableTasks() ([]resource.ID, error) {
	runIDs, err := m.list.Prune(func(task *task.Task) (resource.ID, error) {
		rr := task.Run()
		if rr == nil {
			return resource.ID{}, errors.New("task is not applyable")
		}
		run := rr.(*runpkg.Run)
		if run.Status != runpkg.Planned {
			return resource.ID{}, errors.New("task run is not in the planned state")
		}
		return run.ID, nil
	})
	if err != nil {
		return nil, err
	}
	return runIDs, nil
}

var (
	singlePaneBorder   = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	activePaneBorder   = lipgloss.NewStyle().Border(lipgloss.ThickBorder())
	inactivePaneBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(tui.LighterGrey)
)

func (m ListPreview) View() string {
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
	if m.previewVisible {
		if _, ok := m.list.CurrentRow(); ok {
			components = append(components, previewBorder.Render(
				m.cache.Get(m.current).View()),
			)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, components...)
}
