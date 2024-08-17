package plan

import (
	"errors"
	"io"
	"os"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/workspace"
)

type apply struct {
	opts       CreateOptions
	plan       *plan
	mod        *module.Module
	ws         *workspace.Workspace
	terragrunt bool
}

func (r *apply) taskSpec() (task.Spec, error) {
	if r.plan != nil && !r.plan.HasChanges {
		return task.Spec{}, errors.New("plan does not have any changes to apply")
	}
	spec := task.Spec{
		Parent:      r.ws,
		Path:        r.mod.Path,
		Command:     []string{"apply"},
		Args:        cmdArgs(r.opts, r.ws),
		Env:         []string{r.ws.TerraformEnv()},
		Blocking:    true,
		Description: "apply",
		// If terragrunt is in use then respect module dependencies.
		RespectModuleDependencies: r.terragrunt,
		// Module dependencies are reversed for a destroy.
		InverseDependencyOrder: r.opts.Destroy,
		BeforeExited: func(t *task.Task) (task.Summary, error) {
			out, err := io.ReadAll(t.NewReader(false))
			if err != nil {
				return nil, err
			}
			if r.plan != nil {
				// Plan file can now be safely removed
				_ = os.RemoveAll(r.plan.ArtefactsPath)
			}
			report, err := parseApplyReport(string(out))
			if err != nil {
				return nil, err
			}
			return report, nil
		},
	}
	if r.plan != nil {
		spec.Args = append(spec.Args, r.plan.planPath())
	} else {
		spec.Args = append(spec.Args, "-auto-approve")
	}
	if r.opts.Destroy {
		spec.Description += " (destroy)"
	}
	return spec, nil
}

func IsApplyTask(t *task.Task) bool {
	return len(t.Command) > 0 && t.Command[0] == "apply"
}
