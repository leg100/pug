package run

func ByUpdatedDesc(i, j *Run) int {
	if i.Updated.Before(j.Updated) {
		return 1
	}
	return -1
}
