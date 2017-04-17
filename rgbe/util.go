package rgbe

import (
	"bytes"
	"io"
	"math"
)

const (
	header0       = "#?RADIANCE" // official
	header1       = "#?RGBE"
	header2       = "#?AUTOPANO"
	WhiteEfficacy = 179.0
)

// imageMode represents the mode of the image.
type imageMode int

const (
	mRGBE imageMode = iota
)

// rgbeToRGB converts Radiance RGBE color space to RGB color space.
func rgbeToRGB(r, g, b, e byte, exposure float64) (rr, gg, bb float64) {
	if int(e) > 0 { // a non-zero pixel
		ee := int(e) - (128 + 8)
		f := math.Ldexp(1, ee) / exposure

		rr = float64(r) * f
		gg = float64(g) * f
		bb = float64(b) * f
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
