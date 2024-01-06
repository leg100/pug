package state

func Sort(i, j *Resource) int {
	if i.Address < j.Address {
		return -1
	} else {
		return 1
	}
}
