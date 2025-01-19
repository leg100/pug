package logging

// BySerialDesc sorts log messages by their serial.
func BySerialDesc(i, j Message) int {
	if i.ID.Serial < j.ID.Serial {
		return 1
	}
	return -1
}
