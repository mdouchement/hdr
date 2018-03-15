package util

// IntAbs returns the absolute value of x.
func IntAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Clamp force v to be between min and max.
func Clamp(min, max, v float64) float64 {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}
