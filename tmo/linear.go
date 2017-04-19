package tmo

import (
	"image"
	"image/color"

	"github.com/mdouchement/hdr"
)

// A Linear is a naive TMO implementation.
type Linear struct {
	HDRImage hdr.Image
}

// NewLinear instanciates a new Linear TMO.
func NewLinear(m hdr.Image) *Linear {
	return &Linear{
		HDRImage: m,
	}
}

// Perform runs the TMO mapping.
func (t *Linear) Perform() image.Image {
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy())
	img := image.NewRGBA64(imgRect)

	rmm, gmm, bmm := t.minmax()

	t.shiftRescale(img, rmm, gmm, bmm)

	return img
}

func (t *Linear) minmax() (rmm, gmm, bmm *minmax) {
	rmm, gmm, bmm = newMinMax(), newMinMax(), newMinMax()

	for y := 0; y < t.HDRImage.Bounds().Dy(); y++ {
		for x := 0; x < t.HDRImage.Bounds().Dx(); x++ {
			pixel := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := pixel.HDRRGBA()

			rmm.update(r)
			gmm.update(g)
			bmm.update(b)
		}
	}

	return
}

func (t *Linear) shiftRescale(img *image.RGBA64, rmm, gmm, bmm *minmax) {
	for y := 0; y < t.HDRImage.Bounds().Dy(); y++ {
		for x := 0; x < t.HDRImage.Bounds().Dx(); x++ {
			pixel := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := pixel.HDRRGBA()

			img.SetRGBA64(x, y, color.RGBA64{
				R: shiftRescale(r, rmm),
				G: shiftRescale(g, gmm),
				B: shiftRescale(b, bmm),
				A: RangeMax,
			})
		}
	}
}

func shiftRescale(channel float64, mm *minmax) uint16 {
	if channel < RangeMin {
		channel = RangeMax * (channel + mm.min*-1) / (mm.max + (mm.min * -1))
	} else {
		channel = RangeMax * (channel - mm.min) / (mm.max - mm.min)
	}

	return uint16(channel)
}
