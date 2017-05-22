package crad

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// ebytesToFloats converts Radiance RGBE/XYZE color space to RGB/XYZ color space.
func ebytesToFloats(b0, b1, b2, e byte, exposure float64) (bb0, bb1, bb2 float64) {
	if int(e) > 0 { // a non-zero pixel
		ee := int(e) - (128 + 8)
		f := math.Ldexp(1, ee) / exposure

		bb0 = float64(b0) * f
		bb1 = float64(b1) * f
		bb2 = float64(b2) * f
	}

	return
}

// floatsToEBytes converts RGB/XYZ color space to Radiance RGBE/XYZE (4 bytes slice).
func floatsToEBytes(f1, f2, f3 float64) []byte {
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

func floatsToBytes(f1, f2, f3 float64) []byte {
	pixel := make([]byte, 0, 3*4)

	pixel = append(pixel, float32bytes(float32(f1))...)
	pixel = append(pixel, float32bytes(float32(f2))...)
	pixel = append(pixel, float32bytes(float32(f3))...)

	return pixel
}

func float32bytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

func bytesToFloats(pixel []byte) (float64, float64, float64) {
	f1 := float32frombytes(pixel[0:4])
	f2 := float32frombytes(pixel[4:8])
	f3 := float32frombytes(pixel[8:12])

	return float64(f1), float64(f2), float64(f3)
}

func float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)

	return float
}

func csRGBToYCoCg(r, g, b float64) (y, co, cg float64) {
	co = r - b
	t := b + (co / 2)
	cg = g - t
	y = t + (cg / 2)

	return
}

func csYCoCgToRGB(y, co, cg float64) (r, g, b float64) {
	t := y - (cg / 2)
	g = cg + t
	b = t - (co / 2)
	r = co + b

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

// A FormatError reports that the input is not a valid CRAD image.
type FormatError string

func (e FormatError) Error() string {
	return "crad: invalid format: " + string(e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return "crad: unsupported feature: " + string(e)
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return "crad: internal error: " + string(e)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
