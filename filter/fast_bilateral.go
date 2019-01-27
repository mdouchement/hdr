package filter

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/big"
	"sync"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
	"github.com/mdouchement/hdr/mathx"
	"gonum.org/v1/gonum/mat"
)

const (
	dimension = 5 // x, y & colors (z1, z2, z3)
	// padding space
	paddingS = 2
	// padding range (color)
	paddingR = 2
	// colors' index
	c1 = 0
	c2 = 1
	c3 = 2
)

// A FastBilateral filter is a non-linear, edge-preserving and noise-reducing
// smoothing filter for images. The intensity value at each pixel in an image is
// replaced by a weighted average of intensity values from nearby pixels.
//
// References:
// https://github.com/mdouchement/bilateral
// http://people.csail.mit.edu/sparis/bf/
type FastBilateral struct {
	HDRImage   hdr.Image
	SigmaRange float64
	SigmaSpace float64
	minmaxOnce sync.Once
	min        []float64
	max        []float64
	// Grid size:
	// 0 -> smallWidth
	// 1 -> smallHeight
	// 2 -> smallColor1Depth (gray & color)
	// 3 -> smallColor2Depth (color)
	// 4 -> smallColor3Depth (color)
	size []int
	grid *grid
	auto bool
}

// NewFastBilateralAuto instanciates a new FastBilateral with automatic sigma values.
func NewFastBilateralAuto(m hdr.Image) *FastBilateral {
	f := NewFastBilateral(m, 16, 0.1)
	f.auto = true
	return f
}

// NewFastBilateral instanciates a new FastBilateral.
func NewFastBilateral(m hdr.Image, sigmaSpace, sigmaRange float64) *FastBilateral {
	fbl := &FastBilateral{
		HDRImage:   m,
		SigmaRange: sigmaRange,
		SigmaSpace: sigmaSpace,
		min:        make([]float64, dimension-2),
		max:        make([]float64, dimension-2),
		size:       make([]int, dimension),
	}
	for i := range fbl.min {
		fbl.min[i] = math.Inf(1)
		fbl.max[i] = math.Inf(-1)
	}

	return fbl
}

// Perform runs the bilateral filter.
func (f *FastBilateral) Perform() {
	f.minmaxOnce.Do(f.minmax)
	f.downsampling()
	f.convolution()
}

// ColorModel returns the Image's color model.
func (f *FastBilateral) ColorModel() color.Model {
	return hdrcolor.RGBModel
}

// Bounds implements image.Image interface.
func (f *FastBilateral) Bounds() image.Rectangle {
	return f.HDRImage.Bounds()
}

// Size implements Image.
func (f *FastBilateral) Size() int {
	return f.HDRImage.Size()
}

// HDRAt computes the interpolation and returns the filtered color at the given coordinates.
func (f *FastBilateral) HDRAt(x, y int) hdrcolor.Color {
	pixel := f.HDRImage.HDRAt(x, y)
	r, g, b, _ := pixel.HDRRGBA()
	rgb := []float64{r, g, b}

	offset := make([]float64, dimension)
	// Grid coords
	offset[0] = float64(x)/f.SigmaSpace + paddingS // Grid width
	offset[1] = float64(y)/f.SigmaSpace + paddingS // Grid height
	for z := 0; z < dimension-2; z++ {
		offset[2+z] = (rgb[z]-f.min[z])/f.SigmaRange + paddingR // Grid color
	}

	c := f.nLinearInterpolation(offset...)
	c.colors.ScaleVec(1/c.threshold, c.colors) // Normalize

	return hdrcolor.RGB{
		R: c.colors.AtVec(c1),
		G: c.colors.AtVec(c2),
		B: c.colors.AtVec(c3),
	}
}

// At computes the interpolation and returns the filtered color at the given coordinates.
func (f *FastBilateral) At(x, y int) color.Color {
	r, g, b, _ := f.HDRAt(x, y).HDRRGBA()
	return color.RGBA{
		R: uint8(mathx.Clamp(0, 255, int(r*255))),
		G: uint8(mathx.Clamp(0, 255, int(g*255))),
		B: uint8(mathx.Clamp(0, 255, int(b*255))),
		A: 255,
	}
}

// HDRResultImage computes the interpolation and returns the filtered image.
func (f *FastBilateral) HDRResultImage() hdr.Image {
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
func (f *FastBilateral) ResultImage() hdr.Image {
	d := f.HDRImage.Bounds()
	dst := hdr.NewRGB(d)
	for x := 0; x < d.Dx(); x++ {
		for y := 0; y < d.Dy(); y++ {
			dst.Set(x, y, f.HDRAt(x, y))
		}
	}
	return dst
}

func (f *FastBilateral) minmax() {
	d := f.HDRImage.Bounds()
	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			r, g, b, _ := f.HDRImage.HDRAt(x, y).HDRRGBA()
			for ci, c := range []float64{r, g, b} {
				f.min[ci] = math.Min(f.min[ci], c)
				f.max[ci] = math.Max(f.max[ci], c)
			}
		}
	}

	if f.auto {
		min := math.Inf(1)
		max := math.Inf(-1)
		for n := 0; n < dimension-2; n++ {
			min = math.Min(min, f.min[n])
			max = math.Max(max, f.max[n])
		}
		f.SigmaRange = (max - min) * 0.1
	}

	f.size[0] = int(float64(d.Dx()-1)/f.SigmaSpace) + 1 + 2*paddingS
	f.size[1] = int(float64(d.Dy()-1)/f.SigmaSpace) + 1 + 2*paddingS
	for c := 0; c < dimension-2; c++ {
		f.size[2+c] = int((f.max[c]-f.min[c])/f.SigmaRange) + 1 + 2*paddingR
	}

	// fmt.Println("ssp:", f.SigmaSpace, " - sra:", f.SigmaRange)
	// fmt.Println("min:", f.min, "- max:", f.max)
	// fmt.Println("size:", mathx.Mul(f.size...), f.size)
}

func (f *FastBilateral) downsampling() {
	d := f.HDRImage.Bounds()
	offset := make([]int, dimension)

	dim := dimension - 2
	f.grid = newGrid(f.size, dim)

	for x := 0; x < d.Dx(); x++ {
		offset[0] = int(1*float64(x)/f.SigmaSpace+0.5) + paddingS

		for y := 0; y < d.Dy(); y++ {
			offset[1] = int(1*float64(y)/f.SigmaSpace+0.5) + paddingS

			r, g, b, _ := f.HDRImage.HDRAt(x, y).HDRRGBA()
			rgb := []float64{r, g, b}

			for z := 0; z < dimension-2; z++ {
				offset[2+z] = int((rgb[z]-f.min[z])/f.SigmaRange+0.5) + paddingR
			}

			v := f.grid.At(offset...)
			v.colors.AddVec(v.colors, mat.NewVecDense(dim, rgb))
			v.threshold++
		}
	}
}

func (f *FastBilateral) convolution() {
	dim := dimension - 2
	buffer := newGrid(f.size, dim)

	var vg *cell
	var prev *cell
	var curr *cell
	var next *cell

	for dim := 0; dim < dimension; dim++ { // x, y, and colors depths
		off := make([]int, dimension)
		off[dim] = 1 // Wanted dimension offset

		for n := 0; n < 2; n++ { // itterations (pass?)
			f.grid, buffer = buffer, f.grid

			for x := 1; x < f.size[0]-1; x++ {
				for y := 1; y < f.size[1]-1; y++ {

					for z1 := 1; z1 < f.size[2+c1]-1; z1++ {
						for z2 := 1; z2 < f.size[2+c2]-1; z2++ {
							for z3 := 1; z3 < f.size[2+c3]-1; z3++ {
								vg = f.grid.At(x, y, z1, z2, z3)
								prev = buffer.At(x-off[0], y-off[1], z1-off[2], z2-off[3], z3-off[4])
								curr = buffer.At(x, y, z1, z2, z3)
								next = buffer.At(x+off[0], y+off[1], z1+off[2], z2+off[3], z3+off[4])

								// (prev + 2.0 * curr + next) / 4.0
								vg.Add(prev, next)
								vg.AddScaled(vg, 2, curr)
								vg.Scale(0.25, vg)
							}
						}
					}
				}
			}
		}
	}
}

// Perform linear interpolation.
func (f *FastBilateral) nLinearInterpolation(offset ...float64) *cell {
	permutations := 1 << uint(dimension)
	index := make([]int, dimension)
	indexx := make([]int, dimension)
	alpha := make([]float64, dimension)

	for n, s := range f.size {
		off := offset[n]
		size := s - 1
		index[n] = mathx.Clamp(0, size, int(off))
		indexx[n] = mathx.Clamp(0, size, index[n]+1)
		alpha[n] = off - float64(index[n])
	}

	// Interpolation
	c := &cell{colors: mat.NewVecDense(dimension-2, nil)}
	bitset := big.NewInt(int64(0)) // Use to perform all the interpolation's permutations
	off := make([]int, dimension)
	var scale float64
	for i := 0; i < permutations; i++ {
		bitset.SetUint64(uint64(i))
		scale = 1.0
		for n := 0; n < dimension; n++ {
			if bitset.Bit(n) == 1 {
				off[n] = index[n]
				scale *= 1.0 - alpha[n]
			} else {
				off[n] = indexx[n]
				scale *= alpha[n]
			}
		}
		c.AddScaled(c, scale, f.grid.At(off...))
	}

	return c
}

//-----------------------------------------//
//                                         //
// Grid                                    //
//                                         //
//-----------------------------------------//

const (
	xi = 0
	yi = 1
	zi = 2
)

type (
	// Convenient matrix for FastBilateral filter.
	grid struct {
		size  []int           // X, Y & Zs' dimensions
		cells [][][][][]*cell // X, Y, Z1, Z2 & Z3  coords
	}
	// An entry of the grid
	cell struct {
		colors    *mat.VecDense
		threshold float64 // aka image edges
	}
)

func newGrid(size []int, n int) *grid {
	if len(size) < 3 {
		panic("Grid size must be greater or equals to 3")
	}

	cells := make([][][][][]*cell, size[xi])
	for x := range cells {
		cells[x] = make([][][][]*cell, size[yi])
		for y := range cells[x] {
			cells[x][y] = make([][][]*cell, size[zi])
			for z1 := range cells[x][y] {
				cells[x][y][z1] = make([][]*cell, size[zi+1])
				for z2 := range cells[x][y][z1] {
					cells[x][y][z1][z2] = make([]*cell, size[zi+2])
					for z3 := range cells[x][y][z1][z2] {
						cells[x][y][z1][z2][z3] = &cell{colors: mat.NewVecDense(n, nil)}
					}
				}
			}
		}
	}

	return &grid{
		size:  size,
		cells: cells,
	}
}

func (g *grid) At(offsets ...int) *cell {
	return g.cells[offsets[xi]][offsets[yi]][offsets[zi]][offsets[zi+1]][offsets[zi+2]]
}

func (c *cell) Add(a, b *cell) {
	c.colors.AddVec(a.colors, b.colors)
	c.threshold = a.threshold + b.threshold
}

func (c *cell) Scale(alpha float64, a *cell) {
	c.colors.ScaleVec(alpha, a.colors)
	c.threshold = alpha * a.threshold
}

func (c *cell) AddScaled(a *cell, alpha float64, b *cell) {
	c.colors.AddScaledVec(a.colors, alpha, b.colors)
	c.threshold = a.threshold + alpha*b.threshold
}

func (c *cell) Copy() *cell {
	return &cell{
		colors:    mat.VecDenseCopyOf(c.colors),
		threshold: c.threshold,
	}
}

func (c *cell) String() string {
	return fmt.Sprintf("[c: %v t: %f]", c.colors.RawVector().Data, c.threshold)
}
