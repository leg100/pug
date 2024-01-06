package module

func ByPath(i, j *Module) int {
	if i.Path < j.Path {
		return -1
	}
	return 1
}
