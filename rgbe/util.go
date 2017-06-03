package rgbe

import (
	"bytes"
	"io"

	"github.com/mdouchement/hdr"
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

// // bytesToFloats converts Radiance RGBE/XYZE color space to RGB/XYZ color space.
// func bytesToFloats(b0, b1, b2, e byte, exposure float64) (bb0, bb1, bb2 float64) {
// 	if int(e) > 0 { // a non-zero pixel
// 		ee := int(e) - (128 + 8)
// 		f := math.Ldexp(1, ee) / exposure
//
// 		bb0 = float64(b0) * f
// 		bb1 = float64(b1) * f
// 		bb2 = float64(b2) * f
// 	}
//
// 	return
// }
//
// // floatsToBytes converts RGB/XYZ color space to Radiance RGBE/XYZE (4 bytes slice).
// func floatsToBytes(f1, f2, f3 float64) []byte {
// 	pixel := make([]byte, 4)
//
// 	max := math.Max(f1, f2)
// 	max = math.Max(max, f3)
//
// 	if max > 1e-32 { // a non-zero pixel
// 		mantissa, exponent := math.Frexp(max)
// 		max = mantissa * 256 / max
//
// 		pixel[0] = byte(max * f1)       // R or X
// 		pixel[1] = byte(max * f2)       // G or Y
// 		pixel[2] = byte(max * f3)       // B or Z
// 		pixel[3] = byte(exponent + 128) // exposure (128 is the bias)
// 	}
//
// 	return pixel
// }

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

type ar struct {
	at func(x, y int) (float64, float64, float64)
}

func newAR(m hdr.Image) *ar {
	s := &ar{}

	switch v := m.(type) {
	case *hdr.RGB:
		s.at = func(x, y int) (float64, float64, float64) {
			p := v.RGBAt(x, y)
			return p.R, p.G, p.B
		}
	case *hdr.XYZ:
		s.at = func(x, y int) (float64, float64, float64) {
			p := v.XYZAt(x, y)
			return p.X, p.Y, p.Z
		}
	}

	return s
}

// A FormatError reports that the input is not a valid RGBE image.
type FormatError string

func (e FormatError) Error() string {
	return "rgbe: invalid format: " + string(e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return "rgbe: unsupported feature: " + string(e)
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return "rgbe: internal error: " + string(e)
}
