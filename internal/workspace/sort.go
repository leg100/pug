package workspace

// Sort sorts workspaces accordingly:
//
// 1. first by their module path, lexicographically.
// 2. then, if module paths are equal, then by their workspace name, lexicographically
func Sort(i, j *Workspace) int {
	switch {
	case i.ModulePath() < j.ModulePath():
		return -1
	case i.ModulePath() > j.ModulePath():
		return 1
	default:
		// same module, compare workspace name
		switch {
		case i.Name() < j.Name():
			return -1
		case i.Name() > j.Name():
			return 1
		default:
			// same workspace (unlikely)
			return 0
		}
	}
}
