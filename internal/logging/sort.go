package logging

// BySerialDesc sorts log messages by their serial.
func BySerialDesc(i, j Message) int {
	if i.Serial < j.Serial {
		return 1
	}
	return -1
}
