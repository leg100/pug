package table

func gcd(x, y int) int {
	if x == 0 {
		return y
	} else if y == 0 {
		return x
	}

	return gcd(y%x, x)
}
