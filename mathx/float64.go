package mathx

// ClampF64 forces v to be between min and max.
func ClampF64(min, max, v float64) float64 {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}
