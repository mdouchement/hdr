package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/filter"
	"github.com/mdouchement/hdr/util"
)

// A CustomReinhard05 is a custom Reinhard05 TMO implementation.
// It looks like a JPEG photo taken with a smartphone.
// It provides a quick render with less RAM consumption than Reinhard05.
type CustomReinhard05 struct {
	HDRImage   hdr.Image
	Brightness float64
	Chromatic  float64
	Light      float64
	f          float64
}

// NewDefaultCustomReinhard05 instanciates a new CustomReinhard05 TMO with default parameters.
func NewDefaultCustomReinhard05(m hdr.Image) *CustomReinhard05 {
	return NewCustomReinhard05(m, 0, 0, 0.1)
}

// NewCustomReinhard05 instanciates a new CustomReinhard05 TMO.
func NewCustomReinhard05(m hdr.Image, brightness, chromatic, light float64) *CustomReinhard05 {
	return &CustomReinhard05{
		HDRImage: m,
		// Brightness is included in [-50, 50] with 1 increment step.
		Brightness: brightness * 10,
		// Chromatic is included in [0, 1] with 0.01 increment step.
		Chromatic: chromatic,
		// Light is included in [0, 1] with 0.01 increment step.
		Light: light * 10,
	}
}

// Perform runs the TMO mapping.
func (t *CustomReinhard05) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())

	// Image brightness
	t.f = math.Exp(-t.Brightness)

	minSample, maxSample := t.tonemap()

	t.normalize(img, minSample, maxSample)

	return img
}

func (t *CustomReinhard05) tonemap() (minSample, maxSample float64) {
	qsImg := filter.NewQuickSampling(t.HDRImage, 0.6)

	minSample = math.Inf(1)
	maxSample = math.Inf(-1)
	minCh := make(chan float64)
	maxCh := make(chan float64)

	completed := util.ParallelR(qsImg.Bounds(), func(x1, y1, x2, y2 int) {
		min := 1.0
		max := 0.0

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := qsImg.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()
				_, lum, _, _ := pixel.HDRXYZA()

				var sample float64

				if lum != 0.0 {
					sample = t.sampling(r, lum)
					min = math.Min(min, sample)
					max = math.Max(max, sample)

					sample = t.sampling(g, lum)
					min = math.Min(min, sample)
					max = math.Max(max, sample)

					sample = t.sampling(b, lum)
					min = math.Min(min, sample)
					max = math.Max(max, sample)
				}
			}
		}

		minCh <- min
		maxCh <- max
	})

	for {
		select {
		case <-completed:
			return
		case sample := <-minCh:
			minSample = math.Min(minSample, sample)
		case sample := <-maxCh:
			maxSample = math.Max(maxSample, sample)
		}
	}
}

// sampling one channel
func (t *CustomReinhard05) sampling(sample, lum float64) float64 {
	if sample != 0.0 {
		// Local light adaptation
		il := t.Chromatic*sample + (1-t.Chromatic)*lum
		// Interpolated light adaptation
		ia := t.Light * il
		// Photoreceptor equation
		sample /= sample + math.Pow(t.f*ia, ia)
	}

	return sample
}

func (t *CustomReinhard05) normalize(img *image.RGBA64, minSample, maxSample float64) {
	completed := util.ParallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				img.SetRGBA64(x, y, color.RGBA64{
					R: t.nrmz(r, minSample, maxSample),
					G: t.nrmz(g, minSample, maxSample),
					B: t.nrmz(b, minSample, maxSample),
					A: RangeMax,
				})
			}
		}
	})

	<-completed
}

// normalize one channel
func (t *CustomReinhard05) nrmz(channel, minSample, maxSample float64) uint16 {
	// Normalize intensities
	channel = (channel - minSample) / (maxSample - minSample)

	// Gamma correction
	if channel > RangeMin {
		channel = math.Pow(channel, 1/reinhardGamma)
	}

	// Inverse pixel mapping
	channel = LinearInversePixelMapping(channel, LumPixFloor, LumSize)

	// Clamp to solid black and solid white
	channel = Clamp(channel)

	return uint16(channel)
}
