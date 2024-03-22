package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
)

// CreateTasks returns a command that creates one or more tasks using the given
// IDs. If a task fails to be created then no further tasks will be created, and
// an error notification is sent. If all tasks are successfully created then a status
// notification is sent accordingly.
func CreateTasks(fn task.Func, ids ...resource.ID) tea.Cmd {
	// Handle the case where a user has pressed a key on an empty table with
	// zero rows
	//if len(ids) == 0 {
	//	return nil
	//}

	//// If items have been selected then clear the selection
	//var deselectCmd tea.Cmd
	//if len(ids) > 1 {
	//	deselectCmd = tui.CmdHandler(table.DeselectMsg{})
	//}

	return func() tea.Msg {
		var (
			task *task.Task
			err  error
		)
		for _, id := range ids {
			if task, err = fn(id); err != nil {
				return NewErrorMsg(err, "creating task")
			}
		}
		return InfoMsg(fmt.Sprintf("Created %d %s tasks", len(ids), task.CommandString()))

		//if len(ids) > 1 {
		//	// User has selected multiple rows, so send them to the task *list*
		//	// page
		//	//
		//	// TODO: pass in parameter specifying the parent resource for the
		//	// task listing, i.e. module, workspace, run, etc.
		//	return navigationMsg{
		//		target: page{kind: TaskListKind},
		//	}
		//} else {
		//	// User has highlighted a single row, so send them to the task page.
		//	return navigationMsg{
		//		target: page{kind: TaskKind, resource: task.Resource},
		//	}
		//}
	}

	//return tea.Batch(cmd, deselectCmd)
}

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind Kind, parent *resource.Resource) tea.Cmd {
	return func() tea.Msg {
		page := Page{Kind: kind}
		if parent != nil {
			page.Parent = *parent
		}
		return NavigationMsg(page)
	}
}

func ReportError(err error, msg string, args ...any) tea.Cmd {
	return CmdHandler(NewErrorMsg(err, msg, args...))
}
