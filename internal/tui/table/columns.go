package table

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/run"
)

var (
	IDColumn = Column{
		Key:   "id",
		Title: "ID", Width: resource.IDEncodedMaxLen,
		FlexFactor: 1,
	}
	ModuleColumn = Column{
		Key:            "module",
		Title:          "MODULE",
		TruncationFunc: TruncateLeft,
		FlexFactor:     2,
	}
	WorkspaceColumn = Column{
		Key:        "workspace",
		Title:      "WORKSPACE",
		FlexFactor: 1,
	}
	RunColumn = Column{
		Key:        "run",
		Title:      "RUN",
		Width:      resource.IDEncodedMaxLen,
		FlexFactor: 1,
	}
	TaskColumn = Column{
		Key:        "task",
		Title:      "TASK",
		Width:      resource.IDEncodedMaxLen,
		FlexFactor: 1,
	}
	RunStatusColumn = Column{
		Key:   "run_status",
		Title: "STATUS",
		Width: run.MaxStatusLen,
	}
	RunChangesColumn = Column{
		Key:        "run_changes",
		Title:      "CHANGES",
		FlexFactor: 1,
	}
	ResourceCountColumn = Column{
		Key:   "resource_count",
		Title: "RESOURCES",
		Width: len("RESOURCES"),
	}
)
