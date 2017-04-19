package tmo

import (
	"image"
	"math"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

const (
	maxLum      = 20000
	iCAM06Gamma = 1
)

// A ICam06 is a TMO implementation based on
// Mark D. Fairchild, Jiangtao Kuang and Garrett M. Johnson's 2006 white paper.
//
// Reference:
// iCAM for high-dynamic-range image rendering.
// Mark D. Fairchild, Jiangtao Kuang and Garrett M. Johnson.
// In SIGGRAPH '06 ACM SIGGRAPH 2006 Research posters. Article No. 185
type ICam06 struct {
	HDRImage    hdr.Image
	Contrast    float64
	MinClipping float64
	MaxClipping float64
}

// NewICam06 instanciates a new ICam06 TMO.
func NewICam06(m hdr.Image, contrast, minClipping, maxClipping float64) *ICam06 {
	return &ICam06{
		HDRImage:    m,
		Contrast:    contrast,
		MinClipping: minClipping,
		MaxClipping: maxClipping,
	}
}

// Perform runs the TMO mapping.
func (t *ICam06) Perform() image.Image {
	imgRect := image.Rect(0, 0, t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy())
	img := image.NewRGBA64(imgRect)

	tmp := hdr.NewXYZ(imgRect)
	t.normalize(tmp)
	t.bilateralFilter(tmp)

	return img
}

//================================//
// Step 1: Normalization          //
//================================//

func (t *ICam06) normalize(img *hdr.XYZ) {
	max := math.Inf(-1)
	maxCh := make(chan float64)

	// Find max luminance
	completed := parallel(t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy(), func(x1, y1, x2, y2 int) {
		max := math.Inf(-1)

		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := t.HDRImage.HDRAt(x, y)
				r, g, b, _ := pixel.HDRRGBA()

				xx, yy, zz := colorful.Color{R: r, G: g, B: b}.Xyz() // Convert to CIE XYZ-space, Y is the luminance value.

				max = math.Max(max, yy)

				img.SetXYZ(x, y, hdrcolor.XYZ{X: xx, Y: yy, Z: zz}) // Provision the XYZ temporary image.
			}
		}
	})

	for {
		select {
		case <-completed:
			return
		case m := <-maxCh:
			max = math.Max(max, m)
		}
	}

	// Normalisation
	completed = parallel(t.HDRImage.Bounds().Dx(), t.HDRImage.Bounds().Dy(), func(x1, y1, x2, y2 int) {
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				pixel := img.XYZAt(x, y)

				img.SetXYZ(x, y, hdrcolor.XYZ{
					X: icam06Normalization(pixel.X, max),
					Y: icam06Normalization(pixel.Y, max),
					Z: icam06Normalization(pixel.Z, max),
				})
			}
		}
	})

	<-completed
	return
}

//================================//
// Step 2: Bilateral filte        //
//================================//
//
// The image is splited in two layers:
// - The base layer
// - The details layer

func (t *ICam06) bilateralFilter(img *hdr.XYZ) {
}

//================================//

func icam06Normalization(channel, max float64) float64 {
	channel = maxLum * channel / max

	if channel < 0.00000001 {
		channel = 0.00000001
	}

	return channel
}

// TODO should not be efficient to use it.
func maxtab(sl []float64) float64 {
	max := math.Inf(-1)
	for _, v := range sl {
		max = math.Max(max, v)
	}
	return max
}
