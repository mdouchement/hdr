package format

import (
	"math"
)

// FromRadianceBytes converts Radiance RGBE/XYZE color space to RGB/XYZ color space.
func FromRadianceBytes(b0, b1, b2, e byte, exposure float64) (bb0, bb1, bb2 float64) {
	if int(e) > 0 { // a non-zero pixel
		ee := int(e) - (128 + 8)
		f := math.Ldexp(1, ee) / exposure

		bb0 = float64(b0) * f
		bb1 = float64(b1) * f
		bb2 = float64(b2) * f
	}

	return
}

// ToRadianceBytes converts RGB/XYZ color space to Radiance RGBE/XYZE (4 bytes slice).
func ToRadianceBytes(f1, f2, f3 float64) []byte {
	pixel := make([]byte, 4)

	max := math.Max(f1, f2)
	max = math.Max(max, f3)

	if max > 1e-32 { // a non-zero pixel
		mantissa, exponent := math.Frexp(max)
		max = mantissa * 256 / max

		pixel[0] = byte(max * f1)       // R or X
		pixel[1] = byte(max * f2)       // G or Y
		pixel[2] = byte(max * f3)       // B or Z
		pixel[3] = byte(exponent + 128) // exposure (128 is the bias)
	}

	return pixel
}
