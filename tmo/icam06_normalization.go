package tmo

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/util"
)

// A ICam06Normalization is a part of iCAM06 TMO implementation.
type ICam06Normalization struct {
	HDRImage hdr.Image
	lumOnce  sync.Once
	maxLum   float64
}

// NewICam06Normalization instanciates a new ICam06Normalization TMO.
func NewICam06Normalization(m hdr.Image) *ICam06Normalization {
	return &ICam06Normalization{
		HDRImage: m,
	}
}

// Perform runs the TMO mapping.
func (t *ICam06Normalization) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())

	t.lumOnce.Do(t.luminance)
	t.tonemap(img)

	return img
}

func (t *ICam06Normalization) luminance() {
	maxCh := make(chan float64)

	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		max := math.Inf(-1)

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				_, lum, _, _ := pixel.HDRXYZA()

				max = math.Max(max, lum)
			}
		}

		maxCh <- max
	})

	for {
		select {
		case <-completed:
			return
		case max := <-maxCh:
			t.maxLum = math.Max(t.maxLum, max)
		}
	}
}

func (t *ICam06Normalization) tonemap(img *image.RGBA64) {
	size := t.HDRImage.Size()
	perc := make(percentiles, size*3) // FIXME high memory consumption

	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				xx, yy, zz, _ := pixel.HDRXYZA()

				// XYZ normalization
				xx /= t.maxLum
				yy /= t.maxLum
				zz /= t.maxLum

				// RGB-space conversion
				r, g, b := colorful.XyzToLinearRgb(xx, yy, zz)

				// Clipping, first part
				i := x * y
				perc[i] = r
				perc[size+i] = g
				perc[size*2+i] = b
			}
		}
	})

	<-completed

	perc.sort()
	minRGB := math.Min(perc.percentile(2), 0)
	maxRGB := perc.percentile(98)

	completed = util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				// Clipping, second part
				r = math.Max(math.Min((r-minRGB)/(maxRGB-minRGB), 1), 0)
				g = math.Max(math.Min((g-minRGB)/(maxRGB-minRGB), 1), 0)
				b = math.Max(math.Min((b-minRGB)/(maxRGB-minRGB), 1), 0)

				// RGB normalization
				img.SetRGBA64(x, y, color.RGBA64{
					R: t.normalize(r),
					G: t.normalize(g),
					B: t.normalize(b),
					A: RangeMax,
				})
			}
		}
	})

	<-completed
}

func (t *ICam06Normalization) normalize(channel float64) uint16 {
	c := WoB((channel >= -0.0031308) && (channel <= 0.0031308)) * channel * 12.92
	c += WoB(channel > 0.0031308) * (math.Pow(channel, 1/2.4)*1.055 - 0.055)
	return uint16(RangeMax * c)
}

//-----------------//
// Percentile      //
//-----------------//

type percentiles []float64

func (p percentiles) sort() {
	sort.Sort(p)
}

func (p percentiles) percentile(maxClippingPerc float64) float64 {
	n := float64(len(p))
	i := maxClippingPerc * n / 100
	return float64(p[int(i)])
}

func (p percentiles) Len() int {
	return len(p)
}

func (p percentiles) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p percentiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
