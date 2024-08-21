package state

const (
	Empty ReloadSummary = iota
	Unchanged
	Updated
)

type ReloadSummary int

func (s ReloadSummary) String() string {
	switch s {
	case Empty:
		return "empty"
	case Unchanged:
		return "unchanged"
	case Updated:
		return "updated"
	default:
		return ""
	}
}

func newReloadSummary(before, after *State) ReloadSummary {
	if after == nil {
		return Empty
	}
	if after.Serial == -1 {
		return Empty
	}
	if before == nil {
		return Updated
	}
	if before.Serial == after.Serial {
		return Unchanged
	}
	return Updated
}
