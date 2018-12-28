package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/util"
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
	img := image.NewRGBA64(t.HDRImage.Bounds())

	rmm, gmm, bmm := t.minmax()

	t.shiftLogRescale(img, rmm, gmm, bmm)

	return img
}

//nolint[dupl]
func (t *Logarithmic) minmax() (rmm, gmm, bmm *minmax) {
	rmm, gmm, bmm = newMinMax(), newMinMax(), newMinMax()
	mmCh := make(chan []*minmax)

	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		rrmm, ggmm, bbmm := newMinMax(), newMinMax(), newMinMax()

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				rrmm.update(r)
				ggmm.update(g)
				bbmm.update(b)
			}
		}
		mmCh <- []*minmax{rrmm, ggmm, bbmm}

	})

	for {
		select {
		case <-completed:
			return
		case mm := <-mmCh:
			rmm.update(mm[0].min)
			rmm.update(mm[0].max)

			gmm.update(mm[1].min)
			gmm.update(mm[1].max)

			bmm.update(mm[2].min)
			bmm.update(mm[2].max)
		}
	}
}

func (t *Logarithmic) shiftLogRescale(img *image.RGBA64, rmm, gmm, bmm *minmax) {
	// Calculate max for rescale
	rmax, gmax, bmax := logMax(rmm), logMax(gmm), logMax(bmm)

	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
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
	})

	<-completed
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
