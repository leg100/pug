package navigator

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/tui"
)

// page is an instance of a model, providing the necessary info to
// either construct a model or to identify a previously constructed model.
type page struct {
	// The model kind.
	kind tui.Kind
	// The model's resource. In the case of global listings, this is the global
	// resource. For all other listings, this is the parent resource of the
	// listed resources.
	resource resource.Resource
	// The active tab on the model.
	tab tui.TabTitle
}
