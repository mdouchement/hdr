package filter

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
	"gonum.org/v1/gonum/mat"
)

const yDimension = 3

// A YFastBilateral (very fast) filter is a non-linear, edge-preserving and noise-reducing
// smoothing filter for images. The intensity value at each pixel in an image is
// replaced by a weighted average of intensity values from nearby pixels.
//
// References:
// https://github.com/mdouchement/bilateral
// http://people.csail.mit.edu/sparis/bf/
type YFastBilateral struct {
	HDRImage   hdr.Image
	SigmaRange float64
	SigmaSpace float64
	minmaxOnce sync.Once
	min        float64
	max        float64
	// size:
	// 0 -> smallWidth
	// 1 -> smallHeight
	// 2 -> smallLuminance
	size     []int
	grid     *mat.Dense
	auto     bool
	newColor func(z1, z2, z3 float64) hdrcolor.Color
}

// NewYFastBilateralAuto instanciates a new YFastBilateral with automatic sigma values.
func NewYFastBilateralAuto(m hdr.Image) *YFastBilateral {
	f := NewYFastBilateral(m, 16, 0.1)
	f.auto = true
	return f
}

// NewYFastBilateral instanciates a new YFastBilateral.
func NewYFastBilateral(m hdr.Image, sigmaSpace, sigmaRange float64) *YFastBilateral {
	f := &YFastBilateral{
		HDRImage:   m,
		SigmaRange: sigmaRange,
		SigmaSpace: sigmaSpace,
		min:        math.Inf(1),
		max:        math.Inf(-1),
		size:       make([]int, yDimension),
	}

	switch m.ColorModel() {
	case hdrcolor.XYZModel:
		f.newColor = func(x, y, z float64) hdrcolor.Color {
			return hdrcolor.XYZ{X: x, Y: y, Z: z}
		}
	case hdrcolor.RGBModel:
		fallthrough
	default:
		f.newColor = func(x, y, z float64) hdrcolor.Color {
			return hdrcolor.RGBModel.Convert(hdrcolor.XYZ{X: x, Y: y, Z: z}).(hdrcolor.Color)
		}
	}

	return f
}

// Perform runs the bilateral filter.
func (f *YFastBilateral) Perform() {
	f.minmaxOnce.Do(f.minmax)
	f.downsampling()
	f.convolution()
	f.normalize()
}

// ColorModel returns the Image's color model.
func (f *YFastBilateral) ColorModel() color.Model {
	return f.HDRImage.ColorModel()
}

// Bounds implements image.Image interface.
func (f *YFastBilateral) Bounds() image.Rectangle {
	return f.HDRImage.Bounds()
}

// Size implements Image.
func (f *YFastBilateral) Size() int {
	return f.HDRImage.Size()
}

// HDRAt computes the interpolation and returns the filtered color at the given coordinates.
func (f *YFastBilateral) HDRAt(x, y int) hdrcolor.Color {
	X, Y, Z, _ := f.HDRImage.HDRAt(x, y).HDRXYZA()

	// Grid coords
	gw := float64(x)/f.SigmaSpace + paddingS // Grid width
	gh := float64(y)/f.SigmaSpace + paddingS // Grid height
	gc := (Y-f.min)/f.SigmaRange + paddingR  // Grid Y
	Y2 := f.trilinearInterpolation(gw, gh, gc)

	delta := Y - Y2
	return f.newColor(X-delta, Y2, Z-delta)
}

// At computes the interpolation and returns the filtered color at the given coordinates.
func (f *YFastBilateral) At(x, y int) color.Color {
	r, g, b, _ := f.HDRAt(x, y).HDRRGBA()
	return color.RGBA{
		R: uint8(clamp(0, 255, int(r*255))),
		G: uint8(clamp(0, 255, int(g*255))),
		B: uint8(clamp(0, 255, int(b*255))),
		A: 255,
	}
}

// HDRResultImage computes the interpolation and returns the filtered image.
func (f *YFastBilateral) HDRResultImage() hdr.Image {
	d := f.HDRImage.Bounds()
	dst := hdr.NewRGB(d)
	for x := 0; x < d.Dx(); x++ {
		for y := 0; y < d.Dy(); y++ {
			dst.Set(x, y, f.HDRAt(x, y))
		}
	}
	return dst
}

// ResultImage computes the interpolation and returns the filtered image.
func (f *YFastBilateral) ResultImage() hdr.Image {
	d := f.HDRImage.Bounds()
	dst := hdr.NewRGB(d)
	for x := 0; x < d.Dx(); x++ {
		for y := 0; y < d.Dy(); y++ {
			dst.Set(x, y, f.HDRAt(x, y))
		}
	}
	return dst
}

func (f *YFastBilateral) minmax() {
	d := f.HDRImage.Bounds()
	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			_, Y, _, _ := f.HDRImage.HDRAt(x, y).HDRXYZA()
			f.min = math.Min(f.min, Y)
			f.max = math.Max(f.max, Y)
		}
	}

	if f.auto {
		f.SigmaRange = (f.max - f.min) * 0.1
	}

	f.size[0] = int(float64(d.Dx()-1)/f.SigmaSpace) + 1 + 2*paddingS
	f.size[1] = int(float64(d.Dy()-1)/f.SigmaSpace) + 1 + 2*paddingS
	f.size[2] = int((f.max-f.min)/f.SigmaRange) + 1 + 2*paddingR

	fmt.Println("ssp:", f.SigmaSpace, " - sra:", f.SigmaRange)
	fmt.Println("min:", f.min, "- max:", f.max)
	fmt.Println("size:", mul(f.size...), f.size)
}

func (f *YFastBilateral) downsampling() {
	d := f.HDRImage.Bounds()
	offset := make([]int, yDimension)

	size := mul(f.size...)
	dim := yDimension - 1 // # 1 luminance and 1 threshold (edge weight)
	f.grid = mat.NewDense(size, dim, make([]float64, dim*size))

	for x := 0; x < d.Dx(); x++ {
		offset[0] = int(float64(x)/f.SigmaSpace+0.5) + paddingS

		for y := 0; y < d.Dy(); y++ {
			offset[1] = int(float64(y)/f.SigmaSpace+0.5) + paddingS

			_, Y, _, _ := f.HDRImage.HDRAt(x, y).HDRXYZA()
			offset[2] = int((Y-f.min)/f.SigmaRange+0.5) + paddingR

			i := f.offset(offset...)
			v := f.grid.RawRowView(i)
			v[0] += Y // luminance
			v[1]++    // threshold
			f.grid.SetRow(i, v)
		}
	}
}

func (f *YFastBilateral) convolution() {
	size := mul(f.size...)
	dim := yDimension - 1 // # luminance and 1 threshold (edge weight)
	buffer := mat.NewDense(size, dim, make([]float64, dim*size))

	for dim := 0; dim < yDimension; dim++ { // x, y, and luminance
		off := make([]int, yDimension)
		off[dim] = 1 // Wanted dimension offset

		for n := 0; n < 2; n++ { // itterations (pass?)
			f.grid, buffer = buffer, f.grid

			for x := 1; x < f.size[0]-1; x++ {
				for y := 1; y < f.size[1]-1; y++ {

					for z := 1; z < f.size[2]-1; z++ {
						vg := f.grid.RowView(f.offset(x, y, z)).(*mat.VecDense)
						prev := buffer.RowView(f.offset(x-off[0], y-off[1], z-off[2])).(*mat.VecDense)
						curr := buffer.RowView(f.offset(x, y, z)).(*mat.VecDense)
						next := buffer.RowView(f.offset(x+off[0], y+off[1], z+off[2])).(*mat.VecDense)

						// (prev + 2.0 * curr + next) / 4.0
						vg.AddVec(prev, next)
						vg.AddScaledVec(vg, 2, curr)
						vg.ScaleVec(0.25, vg)
					}
				}
			}
		}
	}
	return
}

func (f *YFastBilateral) normalize() {
	r, _ := f.grid.Dims()
	for i := 0; i < r; i++ {
		if threshold := f.grid.At(i, 1); threshold != 0 {
			f.grid.Set(i, 0, f.grid.At(i, 0)/threshold)
		}
	}
}

func (f *YFastBilateral) trilinearInterpolation(gx, gy, gz float64) float64 {
	width := f.size[0]
	height := f.size[1]
	depth := f.size[2]

	// Index
	x := clamp(0, width-1, int(gx))
	xx := clamp(0, width-1, x+1)
	y := clamp(0, height-1, int(gy))
	yy := clamp(0, height-1, y+1)
	z := clamp(0, depth-1, int(gz))
	zz := clamp(0, depth-1, z+1)

	// Alpha
	xa := gx - float64(x)
	ya := gy - float64(y)
	za := gz - float64(z)

	// Interpolation
	return (1.0-ya)*(1.0-xa)*(1.0-za)*f.grid.At(f.offset(x, y, z), 0) +
		(1.0-ya)*xa*(1.0-za)*f.grid.At(f.offset(xx, y, z), 0) +
		ya*(1.0-xa)*(1.0-za)*f.grid.At(f.offset(x, yy, z), 0) +
		ya*xa*(1.0-za)*f.grid.At(f.offset(xx, yy, z), 0) +
		(1.0-ya)*(1.0-xa)*za*f.grid.At(f.offset(x, y, zz), 0) +
		(1.0-ya)*xa*za*f.grid.At(f.offset(xx, y, zz), 0) +
		ya*(1.0-xa)*za*f.grid.At(f.offset(x, yy, zz), 0) +
		ya*xa*za*f.grid.At(f.offset(xx, yy, zz), 0)
}

// slice[x + WIDTH*y + WIDTH*HEIGHT*z)]
func (f *YFastBilateral) offset(size ...int) (n int) {
	n = size[0] // x
	for i, v := range size[1:] {
		n += v * mul(f.size[0:i+1]...) // y, z
	}
	return
}
