package tmo

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
)

type Linear struct {
	HDRImage hdr.Image
}

func NewLinear(m hdr.Image) *Linear {
	return &Linear{
		HDRImage: m,
	}
}

func (t *Linear) Perform() image.Image {
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Size().X, t.HDRImage.Bounds().Size().Y)
	img := image.NewRGBA64(imgRect)

	rmm, gmm, bmm := t.minmax()

	t.shiftRescale(img, rmm, gmm, bmm)

	return img
}

func (t *Linear) minmax() (rmm, gmm, bmm *minmax) {
	rmm, gmm, bmm = newMinMax(), newMinMax(), newMinMax()

	for y := 0; y < t.HDRImage.Bounds().Size().Y; y++ {
		for x := 0; x < t.HDRImage.Bounds().Size().X; x++ {
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
	for y := 0; y < t.HDRImage.Bounds().Size().Y; y++ {
		for x := 0; x < t.HDRImage.Bounds().Size().X; x++ {
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
	var v float64

	if channel < 0 {
		v = RangeMax * (channel + mm.min*-1) / (mm.max + (mm.min * -1))
	} else {
		v = RangeMax * (channel - mm.min) / (mm.max - mm.min)
	}

	return uint16(v)
}

//--------------------------------------//
// MinMax data                          //
//--------------------------------------//

type minmax struct {
	min float64
	max float64
}

func newMinMax() *minmax {
	return &minmax{
		min: math.Inf(1),
		max: math.Inf(-1),
	}
}

func (mm *minmax) update(v float64) {
	mm.min = math.Min(mm.min, v)
	mm.max = math.Max(mm.max, v)
}

func (mm *minmax) String() string {
	return fmt.Sprintf("min: %f, max: %f", mm.min, mm.max)
}
