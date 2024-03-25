package top

import (
	"errors"
	"fmt"

	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

func handleCreatedTasksMsg(msg tui.CreatedTasksMsg) (info string, err error) {
	if len(msg.Tasks) == 1 {
		// User created one task successfully
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
	return
}

func handleCompletedTasksMsg(msg tui.CompletedTasksMsg) (info string, err error) {
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
