// Code generated by "stringer -type Kind ./internal/resource/kind.go"; DO NOT EDIT.

package resource

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Global-0]
	_ = x[Module-1]
	_ = x[Workspace-2]
	_ = x[Run-3]
	_ = x[Task-4]
}

const _Kind_name = "GlobalModuleWorkspaceRunTask"

var _Kind_index = [...]uint8{0, 6, 12, 21, 24, 28}

func (i Kind) String() string {
	if i < 0 || i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
