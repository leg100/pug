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
	SummaryColumn = Column{
		Key:        "summary",
		Title:      "SUMMARY",
		FlexFactor: 1,
	}
	ResourceCountColumn = Column{
		Key:   "resource_count",
		Title: "RESOURCES",
		Width: len("RESOURCES"),
	}
	CostColumn = Column{
		Key:        "cost",
		Title:      "COST",
		FlexFactor: 1,
		RightAlign: true,
	}
)
