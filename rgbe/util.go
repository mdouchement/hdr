package rgbe

import (
	"bytes"
	"io"
	"math"
)

const (
	header0 = "#?RADIANCE" // official
	header1 = "#?RGBE"
	header2 = "#?AUTOPANO"
)

// imageMode represents the mode of the image.
type imageMode int

const (
	mRGBE imageMode = iota
	mXYZE
)

// bytesToFloats converts Radiance RGBE/XYZE color space to RGB/XYZ color space.
func bytesToFloats(b0, b1, b2, e byte, exposure float64) (bb0, bb1, bb2 float64) {
	if int(e) > 0 { // a non-zero pixel
		ee := int(e) - (128 + 8)
		f := math.Ldexp(1, ee) / exposure

		bb0 = float64(b0) * f
		bb1 = float64(b1) * f
		bb2 = float64(b2) * f
	}

	return
}

func readUntil(r io.Reader, delimiter byte) (string, error) {
	buf := &bytes.Buffer{}
	p := make([]byte, 1)

	for {
		if _, err := r.Read(p); err != nil {
			return "", err
		}

		if p[0] != delimiter {
			buf.Write(p)
		} else {
			return buf.String(), nil
		}
	}
}
