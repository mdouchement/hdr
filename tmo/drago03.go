package tmo

import (
	"image"
	"image/color"
	"math"
	"sync"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mdouchement/hdr"
)

// A Drago03 is an adaptive Drago03 TMO.
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
	return NewDrago03(m, 0)
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
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Size().X, t.HDRImage.Bounds().Size().Y)
	img := image.NewRGBA64(imgRect)

	t.biasP = math.Log10(t.Bias) / math.Log(0.5)

	t.lumOnce.Do(t.luminance)
	t.tonemap(img)

	return img
}

func (t *Drago03) luminance() {
	for y := 0; y < t.HDRImage.Bounds().Dy(); y++ {
		for x := 0; x < t.HDRImage.Bounds().Dx(); x++ {
			pixel := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := pixel.HDRRGBA()

			_, lum, _ := colorful.Color{R: r, G: g, B: b}.Xyz() // Get luminance (Y) from the CIE XYZ-space.
			t.avgLum += math.Log(lum + 1e-4)
			t.maxLum = math.Max(t.maxLum, lum)
		}
	}

	t.avgLum = math.Exp(t.avgLum / float64(t.HDRImage.Bounds().Dx()*t.HDRImage.Bounds().Dy()))
	// Normalize
	t.maxLum = t.maxLum / t.avgLum
	// Set divider
	t.divider = math.Log10(t.maxLum + 1.0)
}

func (t *Drago03) tonemap(img *image.RGBA64) {
	var lumAvgRatio float64
	var newLum float64

	completed := parallel(t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				xx, yy, zz := colorful.Color{R: r, G: g, B: b}.Xyz() // Convert to CIE XYZ-space, Y is the luminance value.

				// Core Drago Equation
				lumAvgRatio = yy / t.avgLum
				newLum = (math.Log(lumAvgRatio+1.0) / math.Log(2.0+math.Pow(lumAvgRatio/t.maxLum, t.biasP)*8.0)) / t.divider

				// Re-scale to new luminance
				scale := newLum / yy
				xx *= scale
				yy *= scale
				zz *= scale

				// XYZ color-space to RGB conversion
				rgb := colorful.Xyz(xx, yy, zz)
				img.SetRGBA64(x, y, color.RGBA64{
					R: t.normalize(rgb.R),
					G: t.normalize(rgb.G),
					B: t.normalize(rgb.B),
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
