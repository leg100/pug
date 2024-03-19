package tui

import (
	"fmt"
	"strings"

	"github.com/leg100/pug/internal/resource"
)

// Breadcrumbs renders the breadcrumbs for a page, i.e. the ancestry of the
// page's resource.
func Breadcrumbs(title string, parent resource.Resource) string {
	// format: <title>(<path>:<workspace>:<run>)
	var crumbs []string
	switch parent.Kind {
	case resource.Run:
		// if parented by a run, then include its ID
		runID := Regular.Copy().Foreground(LightGrey).Render(parent.Run().String())
		crumbs = append([]string{fmt.Sprintf("{%s}", runID)}, crumbs...)
		fallthrough
	case resource.Workspace:
		// if parented by a workspace, then include its name
		name := Regular.Copy().Foreground(Red).Render(parent.Workspace().String())
		crumbs = append([]string{fmt.Sprintf("[%s]", name)}, crumbs...)
		fallthrough
	case resource.Module:
		// if parented by a module, then include its path
		path := Regular.Copy().Foreground(Blue).Render(parent.Module().String())
		crumbs = append([]string{fmt.Sprintf("(%s)", path)}, crumbs...)
	case resource.Global:
		// if parented by global, then state it is global
		global := Regular.Copy().Foreground(Blue).Render("global")
		crumbs = append([]string{fmt.Sprintf("(%s)", global)}, crumbs...)
	}
	return fmt.Sprintf("%s%s", Bold.Render(title), strings.Join(crumbs, ""))
}
