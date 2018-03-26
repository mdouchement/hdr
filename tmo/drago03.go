package tmo

import (
	"image"
	"image/color"
	"math"
	"sync"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/util"
)

// A Drago03 is an adaptive TMO implementation based on Frederic Drago's 2003 white paper.
//
// Reference:
// http://resources.mpi-inf.mpg.de/tmo/logmap/
type Drago03 struct {
	HDRImage hdr.Image
	Bias     float64
	lumOnce  sync.Once
	maxLum   float64
	avgLum   float64
	divider  float64
	biasP    float64
}

// NewDefaultDrago03 instanciates a new Drago03 TMO with default parameters.
func NewDefaultDrago03(m hdr.Image) *Drago03 {
	return NewDrago03(m, 0.5)
}

// NewDrago03 instanciates a new Drago03 TMO.
func NewDrago03(m hdr.Image, bias float64) *Drago03 {
	return &Drago03{
		HDRImage: m,
		// Bias is included in [0, 1] with 0.01 increment step.
		Bias: bias,
	}
}

// Perform runs the TMO mapping.
func (t *Drago03) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())

	t.biasP = math.Log10(t.Bias) / math.Log(0.5)

	t.lumOnce.Do(t.luminance)
	t.tonemap(img)

	return img
}

func (t *Drago03) luminance() {
	avgCh := make(chan float64)
	maxCh := make(chan float64)

	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		var avg float64
		var max float64

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				_, lum, _, _ := pixel.HDRXYZA()

				avg += math.Log(lum + 1e-4)
				max = math.Max(max, lum)
			}
		}

		avgCh <- avg
		maxCh <- max
	})

	for {
		select {
		case <-completed:
			goto NEXT
		case avg := <-avgCh:
			t.avgLum += avg
		case max := <-maxCh:
			t.maxLum = math.Max(t.maxLum, max)
		}
	}
NEXT:

	t.avgLum = math.Exp(t.avgLum / float64(t.HDRImage.Size()))
	// Normalize
	t.maxLum = t.maxLum / t.avgLum
	// Set divider
	t.divider = math.Log10(t.maxLum + 1.0)
}

func (t *Drago03) tonemap(img *image.RGBA64) {
	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		var lumAvgRatio float64
		var newLum float64

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				xx, yy, zz, _ := pixel.HDRXYZA()

				// Core Drago Equation
				lumAvgRatio = yy / t.avgLum
				newLum = (math.Log(lumAvgRatio+1.0) / math.Log(2.0+math.Pow(lumAvgRatio/t.maxLum, t.biasP)*8.0)) / t.divider

				// Re-scale to new luminance
				scale := newLum / yy
				xx *= scale
				yy *= scale
				zz *= scale

				// XYZ color-space to RGB conversion
				r, g, b := colorful.XyzToLinearRgb(xx, yy, zz)
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

func (t *Drago03) normalize(channel float64) uint16 {
	// Inverse pixel mapping
	channel = LinearInversePixelMapping(channel, LumPixFloor, LumSize)

	// Clamp to solid black and solid white
	channel = Clamp(channel)

	return uint16(channel)
}
