package filter

import (
	"image"
	"image/color"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
	"github.com/mdouchement/hdr/mathx"
)

// An Apply filter let's you apply any function on two colors.
type Apply struct {
	HDRImage1 hdr.Image
	HDRImage2 hdr.Image
	hdrat     func(x, y int) hdrcolor.Color
	apply     func(c1, c2 hdrcolor.Color) hdrcolor.Color
}

// NewApply1 instanciates a new Apply filter.
func NewApply1(m1 hdr.Image, apply func(c1, c2 hdrcolor.Color) hdrcolor.Color) *Apply {
	f := &Apply{
		HDRImage1: m1,
		apply:     apply,
	}

	switch m1.ColorModel() {
	case hdrcolor.XYZModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			c := f.apply(f.HDRImage1.HDRAt(x, y), nil)
			return hdrcolor.XYZModel.Convert(c.(color.Color)).(hdrcolor.Color)
		}
	case hdrcolor.RGBModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			c := f.apply(f.HDRImage1.HDRAt(x, y), nil)
			return hdrcolor.RGBModel.Convert(c.(color.Color)).(hdrcolor.Color)
		}
	default:
		panic("Color Model not supported")
	}

	return f
}

// NewApply2 instanciates a new Apply filter.
func NewApply2(m1, m2 hdr.Image, apply func(c1, c2 hdrcolor.Color) hdrcolor.Color) *Apply {
	f := &Apply{
		HDRImage1: m1,
		HDRImage2: m2,
		apply:     apply,
	}

	switch m1.ColorModel() {
	case hdrcolor.XYZModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			c := f.apply(f.HDRImage1.HDRAt(x, y), f.HDRImage2.HDRAt(x, y))
			return hdrcolor.XYZModel.Convert(c.(color.Color)).(hdrcolor.Color)
		}
	case hdrcolor.RGBModel:
		f.hdrat = func(x, y int) hdrcolor.Color {
			c := f.apply(f.HDRImage1.HDRAt(x, y), f.HDRImage2.HDRAt(x, y))
			return hdrcolor.RGBModel.Convert(c.(color.Color)).(hdrcolor.Color)
		}
	default:
		panic("Color Model not supported")
	}

	return f
}

// ColorModel returns the Image's color model.
func (f *Apply) ColorModel() color.Model {
	return f.HDRImage1.ColorModel()
}

// Bounds implements image.Image interface.
func (f *Apply) Bounds() image.Rectangle {
	return f.HDRImage1.Bounds()
}

// Size implements Image.
func (f *Apply) Size() int {
	return f.HDRImage1.Size()
}

// HDRAt computes the log10(x) and returns the filtered color at the given coordinates.
func (f *Apply) HDRAt(x, y int) hdrcolor.Color {
	return f.hdrat(x, y)
}

// At computes the log10(x) and returns the filtered color at the given coordinates.
func (f *Apply) At(x, y int) color.Color {
	r, g, b, _ := f.HDRAt(x, y).HDRRGBA()
	return color.RGBA{
		R: uint8(mathx.Clamp(0, 255, int(r*255))),
		G: uint8(mathx.Clamp(0, 255, int(g*255))),
		B: uint8(mathx.Clamp(0, 255, int(b*255))),
		A: 255,
	}
}
