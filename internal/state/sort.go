package state

func Sort(i, j *Resource) int {
	if i.Address.String() < j.Address.String() {
		return -1
	} else {
		return 1
	}
}
