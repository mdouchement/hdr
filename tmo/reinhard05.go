package tmo

import (
	"image"
	"image/color"
	"math"
	"sync"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

const (
	reinhardGamma = 1.8
)

// A Reinhard05 is a TMO implementation based on Erik Reinhard's 2005 white paper.
//
// Reference:
// Dynamic Range Reduction Inspired by Photoreceptor Physiology.
// E. Reinhard and K. Devlin.
// In IEEE Transactions on Visualization and Computer Graphics, 2005.
type Reinhard05 struct {
	HDRImage   hdr.Image
	Brightness float64
	Chromatic  float64
	Light      float64
	lumOnce    sync.Once
	cav        []float64
	lav        float64
	minLum     float64
	maxLum     float64
	worldLum   float64
	k          float64
	m          float64
	f          float64
	gama       float64
}

// NewDefaultReinhard05 instanciates a new Reinhard05 TMO with default parameters.
func NewDefaultReinhard05(m hdr.Image) *Reinhard05 {
	return NewReinhard05(m, -5, 0.89, 0.89)
}

// NewReinhard05 instanciates a new Reinhard05 TMO.
func NewReinhard05(m hdr.Image, brightness, chromatic, light float64) *Reinhard05 {
	return &Reinhard05{
		HDRImage: m,
		// Brightness is included in [-20, 20] with 0.1 increment step.
		Brightness: brightness,
		// Chromatic is included in [0, 1] with 0.01 increment step.
		Chromatic: chromatic,
		// Light is included in [0, 1] with 0.01 increment step.
		Light:  light,
		cav:    make([]float64, 3),
		minLum: math.Inf(1),
		maxLum: math.Inf(-1),
	}
}

// Perform runs the TMO mapping.
func (t *Reinhard05) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())

	t.lumOnce.Do(t.luminance) // First pass

	// FIXME
	// Extra memory consumption (x2)
	// A temporary image avoids original image modifications and let user applies another TMO on the image.
	//   - It is quite speed to re-read the original file from filesystem.
	// We can avoid this tmp image by re-calculates the sampling but it costs an extra CPU consumption.
	// We can reduce memory consuption by streaming the data to a swap file but iowait could happen.
	tmp := hdr.NewRGB(t.HDRImage.Bounds())

	minSample, maxSample := t.tonemap(tmp) // Second pass

	t.normalize(img, tmp, minSample, maxSample) // Third pass

	return img
}

func (t *Reinhard05) luminance() {
	reinhardCh := make(chan *Reinhard05)

	completed := parallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		tt := NewDefaultReinhard05(nil)

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				_, lum, _ := colorful.Color{R: r, G: g, B: b}.Xyz() // Get luminance (Y) from the CIE XYZ-space.
				tt.minLum = math.Min(tt.minLum, lum)
				tt.maxLum = math.Max(tt.maxLum, lum)
				tt.worldLum += math.Log((2.3e-5) + lum)

				tt.cav[0] += r
				tt.cav[1] += g
				tt.cav[2] += b
				tt.lav += lum
			}
		}

		reinhardCh <- tt
	})

	for {
		select {
		case <-completed:
			goto NEXT
		case tt := <-reinhardCh:
			t.minLum = math.Min(t.minLum, tt.minLum)
			t.maxLum = math.Max(t.maxLum, tt.maxLum)
			t.worldLum += tt.worldLum

			t.cav[0] += tt.cav[0]
			t.cav[1] += tt.cav[1]
			t.cav[2] += tt.cav[2]
			t.lav += tt.lav
		}
	}
NEXT:

	size := float64(t.HDRImage.Size())
	t.worldLum /= size
	t.cav[0] /= size
	t.cav[1] /= size
	t.cav[2] /= size
	t.lav /= size

	t.minLum = math.Log(t.minLum)
	t.maxLum = math.Log(t.maxLum)

	// Image key
	t.k = (t.maxLum - t.worldLum) / (t.maxLum - t.minLum)
	// Image contrast based on key value
	t.m = (0.3 + (0.7 * math.Pow(t.k, 1.4)))
	// Image brightness
	t.f = math.Exp(-t.Brightness)
}

func (t *Reinhard05) tonemap(tmp *hdr.RGB) (minSample, maxSample float64) {
	minSample = 1.0
	maxSample = 0.0
	minCh := make(chan float64)
	maxCh := make(chan float64)

	completed := parallelR(t.HDRImage.Bounds(), func(x1, y1, x2, y2 int) {
		min := 1.0
		max := 0.0

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				_, lum, _ := colorful.Color{R: r, G: g, B: b}.Xyz() // Get luminance (Y) from the CIE XYZ-space.

				var sample float64
				p := hdrcolor.RGB{}

				if lum != 0.0 {
					sample = t.sampling(r, lum, 0)
					min = math.Min(min, sample)
					max = math.Max(max, sample)
					p.R = sample

					sample = t.sampling(g, lum, 1)
					min = math.Min(min, sample)
					max = math.Max(max, sample)
					p.G = sample

					sample = t.sampling(b, lum, 2)
					min = math.Min(min, sample)
					max = math.Max(max, sample)
					p.B = sample

					tmp.SetRGB(x, y, p)
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
func (t *Reinhard05) sampling(sample, lum float64, c int) float64 {
	if sample != 0.0 {
		// Local light adaptation
		il := t.Chromatic*sample + (1-t.Chromatic)*lum
		// Global light adaptation
		ig := t.Chromatic*t.cav[c] + (1-t.Chromatic)*t.lav
		// Interpolated light adaptation
		ia := t.Light*il + (1-t.Light)*ig
		// Photoreceptor equation
		sample /= sample + math.Pow(t.f*ia, t.m)
	}

	return sample
}

func (t *Reinhard05) normalize(img *image.RGBA64, tmp *hdr.RGB, minSample, maxSample float64) {
	completed := parallelR(tmp.Bounds(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := tmp.HDRAt(x, y)
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
func (t *Reinhard05) nrmz(channel, minSample, maxSample float64) uint16 {
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
