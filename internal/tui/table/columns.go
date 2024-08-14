package table

var (
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
