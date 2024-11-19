package explorer

import "github.com/leg100/pug/internal/tui/explorer/tree"

const (
	dirIcon       string = ""
	moduleIcon    string = "󰠲"
	workspaceIcon string = ""
)

var (
	_ tree.Node = (*dirNode)(nil)
	_ tree.Node = (*moduleNode)(nil)
	_ tree.Node = (*workspaceNode)(nil)
)

type dirNode struct {
	name string

	subdirs []dirNode
	modules []moduleNode
}

func (d dirNode) String() string {
	return d.name
}

func (d dirNode) Children() tree.Children {
	return d.modules
}

func (d dirNode) Hidden() bool { return false }

type moduleNode struct {
	name string

	workspaces []workspaceNode
}

func (d moduleNode) String() string {
	return d.name
}

func (d moduleNode) Children() tree.Children {
	return d.workspaces
}

func (d moduleNode) Hidden() bool { return false }

type workspaceNode struct {
	name string
}

func (d workspaceNode) String() string {
	return d.name
}

func (d workspaceNode) Children() tree.Children {
	return nil
}

func (d workspaceNode) Hidden() bool { return false }
