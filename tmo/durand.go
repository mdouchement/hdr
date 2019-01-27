package tmo

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/filter"
	"github.com/mdouchement/hdr/parallel"
)

const (
	durandGamma    = 2.2
	durandInvGamma = 1 / durandGamma
)

// A Durand is a Fast Bilateral Filtering for the Display of High-Dynamic-Range Images.
//
// Reference:
// Fredo Durand and Julie Dorsey in Proceedings of SIGGRAPH 2002
type Durand struct {
	HDRImage hdr.Image
	Contrast float64
	base     hdr.Image
	lumOnce  sync.Once
	minLum   float64
	maxLum   float64
}

// NewDefaultDurand instanciates a new Durand TMO with default parameters.
func NewDefaultDurand(m hdr.Image) *Durand {
	return NewDurand(m, 5)
}

// NewDurand instanciates a new Durand TMO.
func NewDurand(m hdr.Image, contrast float64) *Durand {
	return &Durand{
		HDRImage: m,
		Contrast: contrast,
		minLum:   math.Inf(1),
		maxLum:   math.Inf(-1),
	}
}

// Perform runs the TMO mapping.
func (t *Durand) Perform() image.Image {
	bilateral := filter.NewYFastBilateralAuto(filter.NewLog10(t.HDRImage))
	bilateral.Perform()
	t.base = bilateral // In log10

	t.lumOnce.Do(t.luminance)

	img := image.NewRGBA64(t.HDRImage.Bounds())
	t.tonemap(img)

	return img
}

func (t *Durand) luminance() {
	maxCh := make(chan float64)
	minCh := make(chan float64)

	completed := parallel.TilesR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		min, max := math.Inf(1), math.Inf(-1)

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				_, Y, _, _ := t.base.HDRAt(x, y).HDRXYZA()
				min = math.Min(min, Y)
				max = math.Max(max, Y)
			}
		}

		minCh <- min
		maxCh <- max
	})

	for {
		select {
		case <-completed:
			return
		case min := <-minCh:
			t.minLum = math.Min(t.minLum, min)
		case max := <-maxCh:
			t.maxLum = math.Max(t.maxLum, max)
		}
	}
}

func (t *Durand) tonemap(m *image.RGBA64) {
	compressionFactor := math.Log10(t.Contrast) / (t.maxLum - t.minLum)
	absolute := compressionFactor * (t.maxLum - t.minLum)

	// Color correction
	k1 := 1.48
	k2 := 0.82
	pow := math.Pow(math.Pow(10, compressionFactor), k2)
	s := ((1 + k1) * pow) / (1 + k1*pow)

	completed := parallel.TilesR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				_, Y, _, _ := pixel.HDRXYZA()
				_, Yb, _, _ := t.base.HDRAt(x, y).HDRXYZA()

				Yd := t.clampToZero(Y / math.Pow(10, Yb)) // Luminance detail
				Yd = math.Log10(Yd)                       // In log10

				Yc := Yb*compressionFactor + Yd - absolute // Get the compressed luminance
				Yc = math.Pow(10, Yc)                      // Reverse all log10

				r, g, b, _ := pixel.HDRRGBA()

				// Remove old luminance
				r /= Y
				g /= Y
				b /= Y
				// Color correction
				r = math.Pow(r, s)
				g = math.Pow(g, s)
				b = math.Pow(b, s)
				// Apply new luminance
				r = t.clampToZero(r * Yc)
				g = t.clampToZero(g * Yc)
				b = t.clampToZero(b * Yc)
				// Gamma correction
				r = math.Pow(r, durandInvGamma)
				g = math.Pow(g, durandInvGamma)
				b = math.Pow(b, durandInvGamma)

				m.SetRGBA64(x, y, color.RGBA64{
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

func (t *Durand) normalize(channel float64) uint16 {
	channel = LinearInversePixelMapping(channel, LumPixFloor, LumSize)
	return uint16(LDRClamp(channel))
}

func (t *Durand) clampToZero(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return 0
	}
	return x
}
