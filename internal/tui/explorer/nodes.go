package explorer

import (
	"fmt"
	"path/filepath"

	"github.com/leg100/pug/internal/resource"
)

const (
	dirIcon       string = ""
	moduleIcon    string = "󰠲"
	workspaceIcon string = ""
)

type dirNode struct {
	path string
}

func (d dirNode) String() string {
	return fmt.Sprintf("%s %s", dirIcon, filepath.Base(d.path))
}

type moduleNode struct {
	id   resource.ID
	path string
}

func (m moduleNode) String() string {
	return fmt.Sprintf("%s %s", moduleIcon, filepath.Base(m.path))
}

type workspaceNode struct {
	id   resource.ID
	name string
}

func (w workspaceNode) String() string {
	return fmt.Sprintf("%s %s", workspaceIcon, w.name)
}
