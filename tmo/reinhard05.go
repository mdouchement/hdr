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
	gamma = 1.8
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
	gama       float64
}

// NewDefaultReinhard05 instanciates a new Reinhard05 TMO with default parameters.
func NewDefaultReinhard05(m hdr.Image) *Reinhard05 {
	return NewReinhard05(m, 0, 0, 1)
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
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Size().X, t.HDRImage.Bounds().Size().Y)
	img := image.NewRGBA64(imgRect)

	t.lumOnce.Do(t.luminance) // First pass

	tmp := hdr.NewRGB(imgRect)
	minCol, maxCol := t.tonemap(tmp) // Second pass

	t.normalize(img, tmp, minCol, maxCol) // Third pass

	return img
}

func (t *Reinhard05) luminance() {
	for y := 0; y < t.HDRImage.Bounds().Size().Y; y++ {
		for x := 0; x < t.HDRImage.Bounds().Size().X; x++ {
			pixel := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := pixel.HDRRGBA()

			_, lum, _ := colorful.Color{R: r, G: g, B: b}.Xyz() // Get luminance (Y) from the CIE XYZ-space.
			t.minLum = math.Min(t.minLum, lum)
			t.maxLum = math.Max(t.maxLum, lum)
			t.worldLum += math.Log((2.3e-5) + lum)

			t.cav[0] += r
			t.cav[1] += g
			t.cav[2] += b
			t.lav += lum
		}
	}

	size := float64(t.HDRImage.Bounds().Size().X * t.HDRImage.Bounds().Size().Y)
	t.worldLum /= size
	t.cav[0] /= size
	t.cav[1] /= size
	t.cav[2] /= size
	t.lav /= size

	t.minLum = math.Log(t.minLum)
	t.maxLum = math.Log(t.maxLum)
}

func (t *Reinhard05) tonemap(tmp *hdr.RGB) (minCol, maxCol float64) {
	// Image key
	k := (t.maxLum - t.worldLum) / (t.maxLum - t.minLum)
	// Image contrast based on key value
	m := (0.3 + (0.7 * math.Pow(k, 1.4)))
	// Image brightness
	f := math.Exp(-t.Brightness)

	minCol = 1.0
	maxCol = 0.0
	minCh := make(chan float64)
	maxCh := make(chan float64)

	completed := parallel(t.HDRImage.Bounds().Size().X, t.HDRImage.Bounds().Size().Y, func(x1, y1, x2, y2 int) {
		min := 1.0
		max := 0.0

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				_, lum, _ := colorful.Color{R: r, G: g, B: b}.Xyz() // Get luminance (Y) from the CIE XYZ-space.

				var col float64
				p := hdrcolor.RGB{}

				if lum != 0.0 {
					for c := 0; c < 3; c++ {
						switch c {
						case 0:
							col = r
						case 1:
							col = g
						case 2:
							col = b
						}

						if col != 0.0 {
							il := t.Chromatic*col + (1-t.Chromatic)*lum
							ig := t.Chromatic*t.cav[c] + (1-t.Chromatic)*t.lav
							ia := t.Light*il + (1-t.Light)*ig
							col /= col + math.Pow(f*ia, m)
						}

						min = math.Min(min, col)
						max = math.Max(max, col)

						switch c {
						case 0:
							p.R = col
						case 1:
							p.G = col
						case 2:
							p.B = col
						}
					}

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
		case col := <-minCh:
			minCol = math.Min(minCol, col)
		case col := <-maxCh:
			maxCol = math.Max(maxCol, col)
		}
	}
}

func (t *Reinhard05) normalize(img *image.RGBA64, tmp *hdr.RGB, minCol, maxCol float64) {
	completed := parallel(t.HDRImage.Bounds().Size().X, t.HDRImage.Bounds().Size().Y, func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				img.SetRGBA64(x, y, color.RGBA64{
					R: t.nrmz(r, minCol, maxCol),
					G: t.nrmz(g, minCol, maxCol),
					B: t.nrmz(b, minCol, maxCol),
					A: RangeMax,
				})
			}
		}
	})

	<-completed
}

// normalize one channel
func (t *Reinhard05) nrmz(channel, minCol, maxCol float64) uint16 {
	// Normalize intensities
	channel = (channel - minCol) / (maxCol - minCol)

	// Gamma correction
	if channel < 0 {
		channel = 0
	} else {
		channel = math.Pow(channel, 1/gamma)
	}

	// Inverse pixel mapping
	channel = LinearInversePixelMapping(channel, LumPixFloor, LumSize)

	// Clamp to solid black and solid white
	if channel > RangeMax {
		channel = RangeMax
	}

	return uint16(channel)
}
