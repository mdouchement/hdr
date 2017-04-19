package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
)

// A Logarithmic is a naive TMO implementation.
type Logarithmic struct {
	HDRImage hdr.Image
}

// NewLogarithmic instanciates a new Logarithmic TMO.
func NewLogarithmic(m hdr.Image) *Logarithmic {
	return &Logarithmic{
		HDRImage: m,
	}
}

// Perform runs the TMO mapping.
func (t *Logarithmic) Perform() image.Image {
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy())
	img := image.NewRGBA64(imgRect)

	rmm, gmm, bmm := t.minmax()

	t.shiftLogRescale(img, rmm, gmm, bmm)

	return img
}

func (t *Logarithmic) minmax() (rmm, gmm, bmm *minmax) {
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

func (t *Logarithmic) shiftLogRescale(img *image.RGBA64, rmm, gmm, bmm *minmax) {
	// Calculate max for rescale
	rmax, gmax, bmax := logMax(rmm), logMax(gmm), logMax(bmm)

	for y := 0; y < t.HDRImage.Bounds().Dy(); y++ {
		for x := 0; x < t.HDRImage.Bounds().Dx(); x++ {
			pixel := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := pixel.HDRRGBA()

			img.SetRGBA64(x, y, color.RGBA64{
				R: shiftLogRescale(r, rmm, rmax),
				G: shiftLogRescale(g, gmm, gmax),
				B: shiftLogRescale(b, bmm, bmax),
				A: RangeMax,
			})
		}
	}
}

func shiftLogRescale(channel float64, mm *minmax, max float64) uint16 {
	// ShiftLog
	if channel < RangeMin {
		channel = math.Log(channel + mm.min*-1)
	} else {
		channel = math.Log(channel - mm.min)
	}

	// Rescale
	channel = RangeMax * channel / max

	return uint16(channel)
}

func logMax(mm *minmax) float64 {
	if mm.min < RangeMin {
		return math.Log(mm.max + (mm.min * -1))
	}
	return math.Log(mm.max - mm.min)
}
