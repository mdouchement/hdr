package hdr

import (
	"image"
	"image/color"

	"github.com/mdouchement/hdr/hdrcolor"
)

// Image is a finite rectangular grid of hdrcolor.Color values taken from a color
// model.
type Image interface {
	image.Image

	// HDRAt returns the HDR pixel color at given coordinates.
	HDRAt(x, y int) hdrcolor.Color
}

// NewRGBE returns a new RGBE image with the given bounds.
func NewRGBE(r image.Rectangle) *RGBE {
	w, h := r.Dx(), r.Dy()
	buf := make([]float64, 3*w*h)
	return &RGBE{buf, 3 * w, r}
}

// RGBE is an in-memory image whose At method returns hdrcolor.RGBE values.
type RGBE struct {
	// Pix holds the image's pixels, in R, G, B order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

func (p *RGBE) ColorModel() color.Model { return hdrcolor.RGBEModel }

func (p *RGBE) Bounds() image.Rectangle { return p.Rect }

func (p *RGBE) At(x, y int) color.Color {
	return p.RGBEAt(x, y)
}

func (p *RGBE) HDRAt(x, y int) hdrcolor.Color {
	return p.RGBEAt(x, y)
}

func (p *RGBE) RGBEAt(x, y int) hdrcolor.RGBE {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hdrcolor.RGBE{}
	}
	i := p.PixOffset(x, y)
	return hdrcolor.RGBE{R: p.Pix[i+0], G: p.Pix[i+1], B: p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *RGBE) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

func (p *RGBE) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)

	c1 := hdrcolor.RGBEModel.Convert(c).(hdrcolor.RGBE)
	p.Pix[i+0] = c1.R
	p.Pix[i+1] = c1.G
	p.Pix[i+2] = c1.B
}

func (p *RGBE) SetRGBE(x, y int, c hdrcolor.RGBE) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = c.R
	p.Pix[i+1] = c.G
	p.Pix[i+2] = c.B
}
