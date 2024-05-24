package task

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/navigator"
)

const (
	ApplyCommand       = "apply"
	ReloadStateCommand = "reload state"
)

// CreateTasks returns a command that creates one or more tasks using the given
// IDs. If a task fails to be created then no further tasks will be created, and
// an error notification is sent. If all tasks are successfully created then a status
// notification is sent accordingly.
func CreateTasks(cmd string, issuer resource.Resource, fn task.Func, ids ...resource.ID) tea.Cmd {
	if len(ids) == 0 {
		return nil
	}
	return func() tea.Msg {
		multi, errs := task.CreateMulti(fn, ids...)
		return CreatedTasksMsg{
			Command:    cmd,
			Issuer:     issuer,
			Tasks:      multi,
			CreateErrs: errs,
		}
	}
}

func WaitTasks(created CreatedTasksMsg) tea.Cmd {
	return func() tea.Msg {
		created.Tasks.Wait()
		return CompletedTasksMsg(created)
	}
}

type CreatedTasksMsg struct {
	// The command of the completed tasks (all tasks are assumed to be running
	// the same command).
	Command string
	// The parent resource of the page on which the request to create tasks was
	// issued.
	Issuer resource.Resource
	// Successfully created tasks
	Tasks task.Multi
	// Errors from creating tasks
	CreateErrs []error
}

type CompletedTasksMsg CreatedTasksMsg

func HandleCreatedTasks(msg CreatedTasksMsg) (cmd tea.Cmd, info string, err error) {
	var cmds []tea.Cmd

	// Determine whether and where to navigate the user to.
	switch len(msg.Tasks) {
	case 0:
		// No tasks created, don't send user anywhere.
	case 1:
		if msg.Command == ApplyCommand {
			// Send user to the run's page
			cmds = append(cmds, navigator.Go(tui.RunKind, navigator.WithResource(msg.Tasks[0].Run()), navigator.WithTab(tui.ApplyTab)))
		} else {
			// Send user to the task's page
			cmds = append(cmds, navigator.Go(tui.TaskKind, navigator.WithResource(msg.Tasks[0])))
		}
	default:
		// Multiple tasks. Send the user to the appropriate listing for the model kind that
		// issued the request to create tasks.
		var (
			opts = []navigator.GoOption{navigator.WithResource(msg.Issuer)}
			kind tui.Kind
		)
		if msg.Command == ApplyCommand {
			// Send user to run list
			kind = tui.RunListKind
		} else {
			// Send user to task list
			kind = tui.TaskListKind
		}
		cmds = append(cmds, navigator.Go(kind, opts...))
	}

	if len(msg.Tasks) == 1 && len(msg.CreateErrs) == 0 {
		// User attempted to create only one task and it was successfully created.
		info = fmt.Sprintf("created %s task successfully", msg.Command)
	} else if len(msg.Tasks) == 0 && len(msg.CreateErrs) == 1 {
		// User attempted to create one task but it failed to be created
		err = fmt.Errorf("creating %s task failed: %w", msg.Command, msg.CreateErrs[0])
	} else if len(msg.Tasks) == 0 && len(msg.CreateErrs) > 1 {
		// User attempted to created multiple tasks and they all failed to be
		// created
		err = fmt.Errorf("creating %d %s tasks failed: see logs", len(msg.CreateErrs), msg.Command)
	} else if len(msg.CreateErrs) > 0 {
		// User attempted to create multiple tasks and at least one failed to be
		// created, and at least one succeeded
		err = fmt.Errorf("created %d %s tasks; %d failed to be created; see logs", len(msg.Tasks), msg.Command, len(msg.CreateErrs))
	} else {
		// User attempted to create multiple tasks and all were successfully
		// created.
		info = fmt.Sprintf("created %d %s tasks successfully", len(msg.Tasks), msg.Command)
	}

	// If any tasks were created, then wait for them to complete.
	if len(msg.Tasks) > 0 {
		cmds = append(cmds, WaitTasks(msg))
	}
	return tea.Batch(cmds...), info, err
}

func HandleCompletedTasks(msg CompletedTasksMsg) (info string, err error) {
	if len(msg.Tasks) == 1 && len(msg.CreateErrs) == 0 {
		// Only one task was created and it was successfully created.
		switch msg.Tasks[0].State {
		case task.Exited:
			info = fmt.Sprintf("completed %s task successfully", msg.Command)
		case task.Errored:
			err = fmt.Errorf("completed %s task unsuccessfully: %w", msg.Command, msg.Tasks[0].Err)
		case task.Canceled:
			info = fmt.Sprintf("successfully canceled %s task", msg.Command)
		}
	} else {
		// Originally more than one task was attempted to be created; summarise
		// final states to user.
		var (
			exited   int
			errored  int
			canceled int
		)
		for _, t := range msg.Tasks {
			switch t.State {
			case task.Exited:
				exited++
			case task.Errored:
				errored++
			case task.Canceled:
				canceled++
			}
		}
		info = fmt.Sprintf("completed %s tasks: (%d successful; %d errored; %d canceled; %d uncreated)",
			msg.Command,
			exited,
			errored,
			canceled,
			len(msg.CreateErrs),
		)
		// Upgrade info msg to an error if there were any errors
		if errored > 0 || len(msg.CreateErrs) > 0 {
			err = errors.New(info)
			info = ""
		}
	}
	return
}
