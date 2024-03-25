package table

import (
	"github.com/leg100/pug/internal/resource"
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
		FlexFactor:     3,
	}
	WorkspaceColumn = Column{
		Key:        "workspace",
		Title:      "WORKSPACE",
		FlexFactor: 2,
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
		Key:        "run_status",
		Title:      "STATUS",
		FlexFactor: 1,
	}
	RunChangesColumn = Column{
		Key:        "run_changes",
		Title:      "CHANGES",
		FlexFactor: 1,
	}
)
