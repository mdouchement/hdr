package util

// IntAbs returns the absolute value of x.
func IntAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
