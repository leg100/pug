package tree

import (
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/workspace"
)

type initMsg struct {
	modules    []*module.Module
	workspaces []*workspace.Workspace
}

type buildMsg struct{}
