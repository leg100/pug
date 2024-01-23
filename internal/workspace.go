package internal

import (
	"bufio"
	"io"
	"log/slog"
	"strings"

	"golang.org/x/sync/errgroup"
)

type workspace struct {
	name   string
	module module
}

func (w workspace) plan(runner *runner) error {
	task, err := runner.run(taskspec{
		prog: "tofu",
		args: []string{"plan", "-out", "plan.out"},
		path: w.module.path,
	})
	if err != nil {
		return err
	}
	_, err = io.ReadAll(task.out)
	if err != nil {
		return err
	}
	return task.err
}

func findWorkspaces(modules []module) (workspaces []workspace, err error) {
	g := new(errgroup.Group)
	runner := NewRunner(5)
	for _, mod := range modules {
		if !mod.initialized {
			continue
		}
		m := mod
		g.Go(func() error {
			task, err := runner.run(taskspec{
				prog: "tofu",
				args: []string{"workspace", "list"},
				path: m.path,
			})
			if err != nil {
				return err
			}
			// should output something like this:
			//
			// 1> terraform workspace list
			//   default
			//   non-default-1
			// * non-default-2
			scanner := bufio.NewScanner(task.out)
			for scanner.Scan() {
				out := strings.TrimSpace(scanner.Text())
				if out == "" {
					continue
				}
				if strings.HasPrefix(out, "*") {
					var found bool
					_, out, found = strings.Cut(out, " ")
					if !found {
						slog.Warn("finding workspaces: malformed output", "module", m, "output", scanner.Text())
						continue
					}
				}
				workspaces = append(workspaces, workspace{
					name:   out,
					module: m,
				})
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			return task.err
		})
	}
	return workspaces, g.Wait()
}
