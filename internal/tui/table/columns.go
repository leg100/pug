package table

import (
	"github.com/leg100/pug/internal/resource"
	"github.com/mattn/go-runewidth"
)

var (
	IDColumn = Column{
		Key:   "id",
		Title: "ID", Width: resource.IDEncodedMaxLen,
	}

	ModuleColumn = Column{
		Key:            "module",
		Title:          "MODULE",
		TruncationFunc: runewidth.TruncateLeft,
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
)
