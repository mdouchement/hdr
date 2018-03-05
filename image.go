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
	// Size returns the number of pixels.
	Size() int
}

// ImageSet is an Image where we can set pixels.
type ImageSet interface {
	// Set adds pixel to Image at given x, y.
	Set(x, y int, c color.Color)
}

// Copy copies the src into a new image.
func Copy(src Image) Image {
	switch m := src.(type) {
	case *RGB:
		dst := NewRGB(m.Bounds())
		copy(dst.Pix, m.Pix)
		return dst
	case *RGB64:
		dst := NewRGB64(m.Bounds())
		copy(dst.Pix, m.Pix)
		return dst
	case *XYZ:
		dst := NewXYZ(m.Bounds())
		copy(dst.Pix, m.Pix)
		return dst
	case *XYZ64:
		dst := NewXYZ64(m.Bounds())
		copy(dst.Pix, m.Pix)
		return dst
	default:
		// fallback
		dst := NewRGB64(m.Bounds())
		for y := 0; y < m.Bounds().Dy(); y++ {
			for x := 0; x < m.Bounds().Dx(); x++ {
				dst.Set(x, y, src.HDRAt(x, y))
			}
		}
		return dst
	}
}

//===============//
// RGB           //
//===============//

// RGB is an in-memory 32 bits floating points image whose At method returns hdrcolor.RGB values.
type RGB struct {
	// Pix holds the image's pixels, in R, G, B order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float32
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewRGB returns a new HDR RGB image with the given bounds.
func NewRGB(r image.Rectangle) *RGB {
	w, h := r.Dx(), r.Dy()
	buf := make([]float32, 3*w*h)
	return &RGB{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *RGB) ColorModel() color.Model { return hdrcolor.RGBModel }

// Bounds implements Image.
func (p *RGB) Bounds() image.Rectangle { return p.Rect }

// Size implements Image.
func (p *RGB) Size() int {
	return p.Bounds().Dx() * p.Bounds().Dy()
}

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
	return hdrcolor.RGB{
		R: float64(p.Pix[i+0]),
		G: float64(p.Pix[i+1]),
		B: float64(p.Pix[i+2]),
	}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *RGB) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set adds pixel to Image at given x, y.
func (p *RGB) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)

	c1 := hdrcolor.RGBModel.Convert(c).(hdrcolor.RGB)
	p.Pix[i+0] = float32(c1.R)
	p.Pix[i+1] = float32(c1.G)
	p.Pix[i+2] = float32(c1.B)
}

// SetRGB applies the given RGB color at this coordinate.
func (p *RGB) SetRGB(x, y int, c hdrcolor.RGB) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = float32(c.R)
	p.Pix[i+1] = float32(c.G)
	p.Pix[i+2] = float32(c.B)
}

// RGB64 is an in-memory 64 bits floating points image whose At method returns hdrcolor.RGB values.
type RGB64 struct {
	// Pix holds the image's pixels, in R, G, B order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewRGB64 returns a new HDR RGB image with the given bounds.
func NewRGB64(r image.Rectangle) *RGB64 {
	w, h := r.Dx(), r.Dy()
	buf := make([]float64, 3*w*h)
	return &RGB64{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *RGB64) ColorModel() color.Model { return hdrcolor.RGBModel }

// Bounds implements Image.
func (p *RGB64) Bounds() image.Rectangle { return p.Rect }

// Size implements Image.
func (p *RGB64) Size() int {
	return p.Bounds().Dx() * p.Bounds().Dy()
}

// At implements Image.
func (p *RGB64) At(x, y int) color.Color {
	return p.RGBAt(x, y)
}

// HDRAt implements Image.
func (p *RGB64) HDRAt(x, y int) hdrcolor.Color {
	return p.RGBAt(x, y)
}

// RGBAt returns the RGB color at this coordinate.
func (p *RGB64) RGBAt(x, y int) hdrcolor.RGB {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hdrcolor.RGB{}
	}
	i := p.PixOffset(x, y)
	return hdrcolor.RGB{R: p.Pix[i+0], G: p.Pix[i+1], B: p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *RGB64) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set adds pixel to Image at given x, y.
func (p *RGB64) Set(x, y int, c color.Color) {
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
func (p *RGB64) SetRGB(x, y int, c hdrcolor.RGB) {
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

// XYZ is an in-memory 32 bits floating points image whose At method returns hdrcolor.XYZ values.
type XYZ struct {
	// Pix holds the image's pixels, in X, Y and Z order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float32
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewXYZ returns a new HDR RGB image with the given bounds.
func NewXYZ(r image.Rectangle) *XYZ {
	w, h := r.Dx(), r.Dy()
	buf := make([]float32, 3*w*h)
	return &XYZ{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *XYZ) ColorModel() color.Model { return hdrcolor.XYZModel }

// Bounds implements Image.
func (p *XYZ) Bounds() image.Rectangle { return p.Rect }

// Size implements Image.
func (p *XYZ) Size() int {
	return p.Bounds().Dx() * p.Bounds().Dy()
}

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
	return hdrcolor.XYZ{
		X: float64(p.Pix[i+0]),
		Y: float64(p.Pix[i+1]),
		Z: float64(p.Pix[i+2]),
	}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *XYZ) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set adds pixel to Image at given x, y.
func (p *XYZ) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)

	c1 := hdrcolor.XYZModel.Convert(c).(hdrcolor.XYZ)
	p.Pix[i+0] = float32(c1.X)
	p.Pix[i+1] = float32(c1.Y)
	p.Pix[i+2] = float32(c1.Z)
}

// SetXYZ applies the given XYZ color at this coordinate.
func (p *XYZ) SetXYZ(x, y int, c hdrcolor.XYZ) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = float32(c.X)
	p.Pix[i+1] = float32(c.Y)
	p.Pix[i+2] = float32(c.Z)
}

// XYZ64 is an in-memory 64 bits floating points image whose At method returns hdrcolor.XYZ values.
type XYZ64 struct {
	// Pix holds the image's pixels, in X, Y and Z order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*8].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewXYZ64 returns a new HDR RGB image with the given bounds.
func NewXYZ64(r image.Rectangle) *XYZ64 {
	w, h := r.Dx(), r.Dy()
	buf := make([]float64, 3*w*h)
	return &XYZ64{buf, 3 * w, r}
}

// ColorModel implements Image.
func (p *XYZ64) ColorModel() color.Model { return hdrcolor.XYZModel }

// Bounds implements Image.
func (p *XYZ64) Bounds() image.Rectangle { return p.Rect }

// Size implements Image.
func (p *XYZ64) Size() int {
	return p.Bounds().Dx() * p.Bounds().Dy()
}

// At implements Image.
func (p *XYZ64) At(x, y int) color.Color {
	return p.XYZAt(x, y)
}

// HDRAt implements Image.
func (p *XYZ64) HDRAt(x, y int) hdrcolor.Color {
	return p.XYZAt(x, y)
}

// XYZAt returns the XYZ color at this coordinate.
func (p *XYZ64) XYZAt(x, y int) hdrcolor.XYZ {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hdrcolor.XYZ{}
	}
	i := p.PixOffset(x, y)
	return hdrcolor.XYZ{X: p.Pix[i+0], Y: p.Pix[i+1], Z: p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *XYZ64) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// Set adds pixel to Image at given x, y.
func (p *XYZ64) Set(x, y int, c color.Color) {
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
func (p *XYZ64) SetXYZ(x, y int, c hdrcolor.XYZ) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i+0] = c.X
	p.Pix[i+1] = c.Y
	p.Pix[i+2] = c.Z
}
