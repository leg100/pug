// Code generated by "stringer -type=Kind"; DO NOT EDIT.

package tui

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ModuleListKind-0]
	_ = x[ModuleKind-1]
	_ = x[WorkspaceListKind-2]
	_ = x[WorkspaceKind-3]
	_ = x[RunListKind-4]
	_ = x[RunKind-5]
	_ = x[TaskListKind-6]
	_ = x[TaskKind-7]
	_ = x[TaskDetailsKind-8]
	_ = x[LogListKind-9]
	_ = x[LogKind-10]
}

const _Kind_name = "ModuleListKindModuleKindWorkspaceListKindWorkspaceKindRunListKindRunKindTaskListKindTaskKindTaskDetailsKindLogListKindLogKind"

var _Kind_index = [...]uint8{0, 14, 24, 41, 54, 65, 72, 84, 92, 107, 118, 125}

func (i Kind) String() string {
	if i < 0 || i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
