package hdrcolor

import (
	"image/color"
)

// Color can convert itself to alpha-premultiplied 16-bits per channel RGBA and HDR float64 RGB.
// The conversion may be lossy.
type Color interface {
	color.Color

	// HDRRGBA returns the red, green, blue and alpha values
	// for the HDR color.
	HDRRGBA() (r, g, b, a float64)
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
	// Ugly cast
	r = uint32(c.R)
	g = uint32(c.G)
	b = uint32(c.B)
	a = 4294967295 // Max uint32

	return
}

// HDRRGBA returns the red, green, blue and alpha values
// for the HDR color.
func (c RGB) HDRRGBA() (r, g, b, a float64) {
	r, g, b = c.R, c.G, c.B
	a = 4294967295.0 // Max uint32 in float64

	return
}

// Models for the standard color types.
var (
	RGBModel color.Model = color.ModelFunc(rgbModel)
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
