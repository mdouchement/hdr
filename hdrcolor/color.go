package hdrcolor

import (
	"image/color"
)

type Color interface {
	color.Color

	// HDRRGBA returns the alpha-premultiplied red, green, blue and alpha values
	// for the color.
	//
	// An alpha-premultiplied color component c has been scaled by alpha (a),
	// so has valid values 0 <= c <= a.
	HDRRGBA() (r, g, b, a float64)
}

// RGBE represents a HDR color.
type RGBE struct {
	R, G, B float64
}

func (c RGBE) RGBA() (r, g, b, a uint32) {
	// Ugly cast
	r = uint32(c.R)
	g = uint32(c.G)
	b = uint32(c.B)
	a = 255

	return
}

func (c RGBE) HDRRGBA() (r, g, b, a float64) {
	r, g, b = c.R, c.G, c.B
	a = 255.0

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
		return RGBE{r, g, b}
	}

	// LDR color
	r, g, b, _ := c.RGBA()
	return RGBE{float64(r), float64(g), float64(b)}
}
