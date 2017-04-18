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

// RGBE represents a HDR color.
type RGBE struct {
	R, G, B float64
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
func (c RGBE) RGBA() (r, g, b, a uint32) {
	// Ugly cast
	r = uint32(c.R)
	g = uint32(c.G)
	b = uint32(c.B)
	a = 4294967295 // Max uint32

	return
}

// HDRRGBA returns the red, green, blue and alpha values
// for the HDR color.
func (c RGBE) HDRRGBA() (r, g, b, a float64) {
	r, g, b = c.R, c.G, c.B
	a = 4294967295.0 // Max uint32 in float64

	return
}

// Models for the standard color types.
var (
	RGBEModel color.Model = color.ModelFunc(rgbeModel)
)

func rgbeModel(c color.Color) color.Color {
	if _, ok := c.(RGBE); ok {
		// Already RGBE
		return c
	}

	if hdrc, ok := c.(Color); ok {
		// HDR color
		r, g, b, _ := hdrc.HDRRGBA()
		return RGBE{R: r, G: g, B: b}
	}

	// LDR color
	r, g, b, _ := c.RGBA()
	return RGBE{R: float64(r), G: float64(g), B: float64(b)}
}
