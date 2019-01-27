package tmo

import (
	"image"
	"image/color"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/parallel"
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
	img := image.NewRGBA64(t.HDRImage.Bounds())

	rmm, gmm, bmm := t.minmax()

	t.shiftRescale(img, rmm, gmm, bmm)

	return img
}

//nolint[dupl]
func (t *Linear) minmax() (rmm, gmm, bmm *minmax) {
	rmm, gmm, bmm = newMinMax(), newMinMax(), newMinMax()
	mmCh := make(chan []*minmax)

	completed := parallel.TilesR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
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

func (t *Linear) shiftRescale(img *image.RGBA64, rmm, gmm, bmm *minmax) {
	completed := parallel.TilesR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
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
	})

	<-completed
}

func shiftRescale(channel float64, mm *minmax) uint16 {
	if channel < RangeMin {
		channel = RangeMax * (channel + mm.min*-1) / (mm.max + (mm.min * -1))
	} else {
		channel = RangeMax * (channel - mm.min) / (mm.max - mm.min)
	}

	return uint16(channel)
}
