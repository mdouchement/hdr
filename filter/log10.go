package filter

import (
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
	"github.com/mdouchement/hdr/xmath"
)

// A Log10 applies a log10 for all pixels of the image.
type Log10 struct {
	HDRImage hdr.Image
	hdrat    func(x, y int) hdrcolor.Color
}

// NewLog10 instanciates a new Log10 filter.
func NewLog10(m hdr.Image) *Log10 {
	f := &Log10{
		HDRImage: m,
	}

	log10 := func(x float64) float64 {
		if x < 0.0001 {
			x = 0.0001
		}
		return math.Log10(x)
	}

	switch m.ColorModel() {
	case hdrcolor.XYZModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			X, Y, Z, _ := f.HDRImage.HDRAt(x, y).HDRXYZA()
			return hdrcolor.XYZ{
				X: log10(X),
				Y: log10(Y),
				Z: log10(Z),
			}
		}
	case hdrcolor.RGBModel:
		fallthrough
	default:
		f.hdrat = func(x, y int) hdrcolor.Color {
			r, g, b, _ := f.HDRImage.HDRAt(x, y).HDRRGBA()
			return hdrcolor.RGB{
				R: log10(r),
				G: log10(g),
				B: log10(b),
			}
		}
	}

	return f
}

// ColorModel returns the Image's color model.
func (f *Log10) ColorModel() color.Model {
	return f.HDRImage.ColorModel()
}

// Bounds implements image.Image interface.
func (f *Log10) Bounds() image.Rectangle {
	return f.HDRImage.Bounds()
}

// Size implements Image.
func (f *Log10) Size() int {
	return f.HDRImage.Size()
}

// HDRAt computes the log10(x) and returns the filtered color at the given coordinates.
func (f *Log10) HDRAt(x, y int) hdrcolor.Color {
	return f.hdrat(x, y)
}

// At computes the log10(x) and returns the filtered color at the given coordinates.
func (f *Log10) At(x, y int) color.Color {
	r, g, b, _ := f.HDRAt(x, y).HDRRGBA()
	return color.RGBA{
		R: uint8(xmath.Clamp(0, 255, int(r*255))),
		G: uint8(xmath.Clamp(0, 255, int(g*255))),
		B: uint8(xmath.Clamp(0, 255, int(b*255))),
		A: 255,
	}
}
