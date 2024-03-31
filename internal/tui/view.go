package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
)

type Breadcrumbs struct {
	ModuleService    *module.Service
	WorkspaceService *workspace.Service
}

// Breadcrumbs renders the breadcrumbs for a page, i.e. the ancestry of the
// page's resource.
func (b *Breadcrumbs) Render(title string, parent resource.Resource) string {
	// format: title[workspace name](module path)
	var crumbs []string
	switch parent.Kind {
	case resource.Run:
		// get parent workspace
		parent = *parent.Parent
		fallthrough
	case resource.Workspace:
		ws, err := b.WorkspaceService.Get(parent.ID)
		if err != nil {
			slog.Error("rendering workspace name", "error", err)
			break
		}
		name := Regular.Copy().Foreground(Red).Render(ws.Name)
		crumbs = append(crumbs, fmt.Sprintf("[%s]", name))
		// now get parent of workspace which is module
		parent = *parent.Parent
		fallthrough
	case resource.Module:
		mod, err := b.ModuleService.Get(parent.ID)
		if err != nil {
			slog.Error("rendering module path", "error", err)
			break
		}
		path := Regular.Copy().Foreground(Blue).Render(mod.Path)
		crumbs = append(crumbs, fmt.Sprintf("(%s)", path))
	case resource.Global:
		// if parented by global, then state it is global
		global := Regular.Copy().Foreground(Blue).Render("all")
		crumbs = append(crumbs, fmt.Sprintf("(%s)", global))
	}
	return fmt.Sprintf("%s%s", Bold.Render(title), strings.Join(crumbs, ""))
}

func GlobalBreadcrumb(title string) string {
	title = Bold.Render(title)
	all := Regular.Copy().Foreground(Blue).Render("all")
	return fmt.Sprintf("%s(%s)", title, all)
}
