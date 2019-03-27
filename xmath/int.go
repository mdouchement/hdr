package xmath

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Clamp forces v to be between min and max.
func Clamp(min, max, v int) int {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}

// Mul multiplies the given ints together.
func Mul(ints ...int) (n int) {
	n = 1
	for _, v := range ints {
		if v != 0 {
			n *= v
		}
	}
	return
}
