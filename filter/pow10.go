package filter

import (
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
	"github.com/mdouchement/hdr/mathx"
)

// A Pow10 applies a pow10 for all pixels of the image.
type Pow10 struct {
	HDRImage hdr.Image
	hdrat    func(x, y int) hdrcolor.Color
}

// NewPow10 instanciates a new Pow10 filter.
func NewPow10(m hdr.Image) *Pow10 {
	f := &Pow10{
		HDRImage: m,
	}

	pow10 := func(x float64) float64 {
		if x > 12 {
			x = 0
		}
		return math.Pow(10, x)
	}

	switch m.ColorModel() {
	case hdrcolor.XYZModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			X, Y, Z, _ := f.HDRImage.HDRAt(x, y).HDRXYZA()
			return hdrcolor.XYZ{
				X: pow10(X),
				Y: pow10(Y),
				Z: pow10(Z),
			}
		}
	case hdrcolor.RGBModel:
		fallthrough
	default:
		f.hdrat = func(x, y int) hdrcolor.Color {
			r, g, b, _ := f.HDRImage.HDRAt(x, y).HDRRGBA()
			return hdrcolor.RGB{
				R: pow10(r),
				G: pow10(g),
				B: pow10(b),
			}
		}
	}

	return f
}

// ColorModel returns the Image's color model.
func (f *Pow10) ColorModel() color.Model {
	return f.HDRImage.ColorModel()
}

// Bounds implements image.Image interface.
func (f *Pow10) Bounds() image.Rectangle {
	return f.HDRImage.Bounds()
}

// Size implements Image.
func (f *Pow10) Size() int {
	return f.HDRImage.Size()
}

// HDRAt computes the pow10(x) and returns the filtered color at the given coordinates.
func (f *Pow10) HDRAt(x, y int) hdrcolor.Color {
	return f.hdrat(x, y)
}

// At computes the pow10(x) and returns the filtered color at the given coordinates.
func (f *Pow10) At(x, y int) color.Color {
	r, g, b, _ := f.HDRAt(x, y).HDRRGBA()
	return color.RGBA{
		R: uint8(mathx.Clamp(0, 255, int(r*255))),
		G: uint8(mathx.Clamp(0, 255, int(g*255))),
		B: uint8(mathx.Clamp(0, 255, int(b*255))),
		A: 255,
	}
}
