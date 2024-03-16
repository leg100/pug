package run

const (
	CreateAction ChangeAction = "create"
	UpdateAction ChangeAction = "update"
	DeleteAction ChangeAction = "delete"
)

type (
	// planFile represents the schema of a plan file
	planFile struct {
		ResourceChanges []ResourceChange  `json:"resource_changes"`
		OutputChanges   map[string]Change `json:"output_changes"`
	}

	// ResourceChange represents a proposed change to a resource in a plan file
	ResourceChange struct {
		Change Change
	}

	// Change represents the type of change being made
	Change struct {
		Actions []ChangeAction
	}

	ChangeAction string
)

//lint:ignore U1000 intend to use shortly
func (pf *planFile) resourceChanges() (resource report) {
	for _, rc := range pf.ResourceChanges {
		for _, action := range rc.Change.Actions {
			switch action {
			case CreateAction:
				resource.Additions++
			case UpdateAction:
				resource.Changes++
			case DeleteAction:
				resource.Destructions++
			}
		}
	}
	return
}

//lint:ignore U1000 intend to use shortly
func (pf *planFile) outputChanges() (output report) {
	for _, rc := range pf.OutputChanges {
		for _, action := range rc.Actions {
			switch action {
			case CreateAction:
				output.Additions++
			case UpdateAction:
				output.Changes++
			case DeleteAction:
				output.Destructions++
			}
		}
	}
	return
}
