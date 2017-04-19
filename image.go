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

//===============//
// RGB           //
//===============//

// RGB is an in-memory image whose At method returns hdrcolor.RGB values.
type RGB struct {
	// Pix holds the image's pixels, in R, G, B order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewRGB returns a new HDR RGB image with the given bounds.
func NewRGB(r image.Rectangle) *RGB {
	w, h := r.Dx(), r.Dy()
	buf := make([]float64, 3*w*h)
	return &RGB{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *RGB) ColorModel() color.Model { return hdrcolor.RGBModel }

// Bounds implements Image.
func (p *RGB) Bounds() image.Rectangle { return p.Rect }

// At implements Image.
func (p *RGB) At(x, y int) color.Color {
	return p.RGBAt(x, y)
}

// HDRAt implements Image.
func (p *RGB) HDRAt(x, y int) hdrcolor.Color {
	return p.RGBAt(x, y)
}

// RGBAt returns the RGB color at this coordinate.
func (p *RGB) RGBAt(x, y int) hdrcolor.RGB {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hdrcolor.RGB{}
	}
	i := p.PixOffset(x, y)
	return hdrcolor.RGB{R: p.Pix[i+0], G: p.Pix[i+1], B: p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *RGB) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set implements Image.
func (p *RGB) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)

	c1 := hdrcolor.RGBModel.Convert(c).(hdrcolor.RGB)
	p.Pix[i+0] = c1.R
	p.Pix[i+1] = c1.G
	p.Pix[i+2] = c1.B
}

// SetRGB applies the given RGB color at this coordinate.
func (p *RGB) SetRGB(x, y int, c hdrcolor.RGB) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = c.R
	p.Pix[i+1] = c.G
	p.Pix[i+2] = c.B
}

//===============//
// XYZ           //
//===============//

// XYZ is an in-memory image whose At method returns hdrcolor.XYZ values.
type XYZ struct {
	// Pix holds the image's pixels, in X, Y and Z order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewXYZ returns a new HDR RGB image with the given bounds.
func NewXYZ(r image.Rectangle) *XYZ {
	w, h := r.Dx(), r.Dy()
	buf := make([]float64, 3*w*h)
	return &XYZ{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *XYZ) ColorModel() color.Model { return hdrcolor.XYZModel }

// Bounds implements Image.
func (p *XYZ) Bounds() image.Rectangle { return p.Rect }

// At implements Image.
func (p *XYZ) At(x, y int) color.Color {
	return p.XYZAt(x, y)
}

// HDRAt implements Image.
func (p *XYZ) HDRAt(x, y int) hdrcolor.Color {
	return p.XYZAt(x, y)
}

// XYZAt returns the XYZ color at this coordinate.
func (p *XYZ) XYZAt(x, y int) hdrcolor.XYZ {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hdrcolor.XYZ{}
	}
	i := p.PixOffset(x, y)
	return hdrcolor.XYZ{X: p.Pix[i+0], Y: p.Pix[i+1], Z: p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *XYZ) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set implements Image.
func (p *XYZ) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)

	c1 := hdrcolor.XYZModel.Convert(c).(hdrcolor.XYZ)
	p.Pix[i+0] = c1.X
	p.Pix[i+1] = c1.Y
	p.Pix[i+2] = c1.Z
}

// SetXYZ applies the given XYZ color at this coordinate.
func (p *XYZ) SetXYZ(x, y int, c hdrcolor.XYZ) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = c.X
	p.Pix[i+1] = c.Y
	p.Pix[i+2] = c.Z
}
