package top

import (
	"errors"
	"testing"

	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/stretchr/testify/assert"
)

func Test_handleCompletedTasksMsg(t *testing.T) {
	tests := []struct {
		msg     tui.CompletedTasksMsg
		want    string
		wantErr string
	}{
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks:   task.Multi{&task.Task{State: task.Exited}},
			},
			"completed init task successfully",
			"",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Canceled},
				},
			},
			"successfully canceled init task",
			"",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Errored, Err: errors.New("exit code 1")},
				},
			},
			"",
			"completed init task unsuccessfully: exit code 1",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Exited},
					&task.Task{State: task.Exited},
					&task.Task{State: task.Exited},
				},
			},
			"completed init tasks: (3 successful; 0 errored; 0 canceled; 0 uncreated)",
			"",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Exited},
					&task.Task{State: task.Exited},
					&task.Task{State: task.Errored},
				},
			},
			"",
			"completed init tasks: (2 successful; 1 errored; 0 canceled; 0 uncreated)",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Exited},
					&task.Task{State: task.Errored},
					&task.Task{State: task.Canceled},
				},
			},
			"",
			"completed init tasks: (1 successful; 1 errored; 1 canceled; 0 uncreated)",
		},
		{
			tui.CompletedTasksMsg{
				Command: "init",
				Tasks: task.Multi{
					&task.Task{State: task.Exited},
					&task.Task{State: task.Exited},
					&task.Task{State: task.Exited},
				},
				CreateErrs: []error{errors.New("creation error")},
			},
			"",
			"completed init tasks: (3 successful; 0 errored; 0 canceled; 1 uncreated)",
		},
	}
	for _, tt := range tests {
		// Name the test after whatever is wanted
		name := tt.want
		if tt.wantErr != "" {
			name = tt.wantErr
		}
		t.Run(name, func(t *testing.T) {
			got, gotErr := handleCompletedTasksMsg(tt.msg)
			assert.Equal(t, tt.want, got)

			// If test expects an error then check error string matches.
			// Otherwise there should be no error.
			if tt.wantErr != "" && assert.Error(t, gotErr) {
				assert.Equal(t, tt.wantErr, gotErr.Error())
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}
