package filter

import (
	"image"
	"image/color"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// A QuickSampling allows to iterate on pixels over a reduced image size.
type QuickSampling struct {
	HDRImage hdr.Image
	sampling float32
	rect     image.Rectangle
}

// NewQuickSampling instanciates a new QuickSampling for the given sampling included in [0, 1].
func NewQuickSampling(img hdr.Image, sampling float32) *QuickSampling {
	if sampling < 0 || sampling > 1 {
		panic("Invalid sampling value. It must be include in [0, 1]")
	}

	y := float32(img.Bounds().Max.Y)
	return &QuickSampling{
		HDRImage: img,
		sampling: sampling,
		rect:     image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, img.Bounds().Max.X, int(y*sampling)),
	}
}

// ColorModel delegates to HDRImage function.
func (f *QuickSampling) ColorModel() color.Model {
	return f.HDRImage.ColorModel()
}

// Bounds implements Image with quick sampling HDRImage.
func (f *QuickSampling) Bounds() image.Rectangle {
	return f.rect
}

// Size implements Image with quick sampling HDRImage.
func (f *QuickSampling) Size() int {
	return f.Bounds().Dx() * f.Bounds().Dy()
}

// At implements Image with quick sampling HDRImage.
func (f *QuickSampling) At(x, y int) color.Color {
	rx, ry := f.realAt(x, y)
	return f.HDRImage.At(rx, ry)
}

// HDRAt implements Image with quick sampling on HDRImage.
func (f *QuickSampling) HDRAt(x, y int) hdrcolor.Color {
	rx, ry := f.realAt(x, y)
	return f.HDRImage.HDRAt(rx, ry)
}

func (f *QuickSampling) realAt(x, y int) (int, int) {
	return x, int(float32(y) * f.sampling)
}
