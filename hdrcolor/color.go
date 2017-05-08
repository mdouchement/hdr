package hdrcolor

import (
	"image/color"

	colorful "github.com/lucasb-eyer/go-colorful"
)

// Color can convert itself to alpha-premultiplied 16-bits per channel RGBA and HDR float64 RGB.
// The conversion may be lossy.
type Color interface {
	color.Color

	// HDRRGBA returns the red, green, blue and alpha values
	// for the HDR color.
	HDRRGBA() (r, g, b, a float64)

	// HDRXYZA returns the x, y, z and alpha values
	// for the HDR color.
	HDRXYZA() (x, y, z, a float64)
}

// RGB represents a HDR color in RGB color-space.
type RGB struct {
	R, G, B float64
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * 65535.0)
	g = uint32(c.G * 65535.0)
	b = uint32(c.B * 65535.0)
	a = 0xFFFF

	return
}

// HDRRGBA returns the red, green, blue and alpha values
// for the HDR color.
func (c RGB) HDRRGBA() (r, g, b, a float64) {
	r, g, b = c.R, c.G, c.B
	a = 0xFFFF

	return
}

// HDRXYZA returns the x, y, z and alpha values
// for the HDR color.
func (c RGB) HDRXYZA() (x, y, z, a float64) {
	x, y, z = colorful.LinearRgbToXyz(c.R, c.G, c.B)
	a = 0xFFFF

	return
}

// XYZ represents a HDR color in XYZ color-space.
type XYZ struct {
	X, Y, Z float64
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
func (c XYZ) RGBA() (r, g, b, a uint32) {
	rr, gg, bb, aa := c.HDRRGBA()
	r = uint32(rr * 65535.0)
	g = uint32(gg * 65535.0)
	b = uint32(bb * 65535.0)
	a = uint32(aa * 65535.0)

	return
}

// HDRRGBA returns the red, green, blue and alpha values
// for the HDR color.
func (c XYZ) HDRRGBA() (r, g, b, a float64) {
	r, g, b = colorful.XyzToLinearRgb(c.X, c.Y, c.Z)

	a = 0xFFFF

	return
}

// HDRXYZA returns the x, y, z and alpha values
// for the HDR color.
func (c XYZ) HDRXYZA() (x, y, z, a float64) {
	x, y, z = c.X, c.Y, c.Z
	a = 0xFFFF

	return
}

// Models for the standard color types.
var (
	RGBModel color.Model = color.ModelFunc(rgbModel)
	XYZModel color.Model = color.ModelFunc(xyzModel)
)

func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		// Already RGB
		return c
	}

	if hdrc, ok := c.(Color); ok {
		// HDR color
		r, g, b, _ := hdrc.HDRRGBA()
		return RGB{R: r, G: g, B: b}
	}

	// LDR color
	r, g, b, _ := c.RGBA()
	return RGB{R: float64(r), G: float64(g), B: float64(b)}
}

func xyzModel(c color.Color) color.Color {
	if _, ok := c.(XYZ); ok {
		// Already XYZ
		return c
	}

	if hdrc, ok := c.(Color); ok {
		// HDR color
		r, g, b, _ := hdrc.HDRRGBA()
		x, y, z := colorful.LinearRgbToXyz(r, g, b)
		return XYZ{X: x, Y: y, Z: z}
	}

	// LDR color
	r, g, b, _ := c.RGBA()
	x, y, z := colorful.LinearRgbToXyz(float64(r), float64(g), float64(b))
	return XYZ{X: x, Y: y, Z: z}
}
