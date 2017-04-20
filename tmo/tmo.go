package tmo

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"sync"

	"github.com/mdouchement/hdr"
)

const (
	// RangeMin is the LDR lower boundary.
	RangeMin = 0
	// RangeMax is the LDR higher boundary.
	RangeMax = 65535 // max uint16
	// LumSize defines the luminance range [0, LumSize]
	LumSize = RangeMax + 3
)

var (
	// LumPixFloor is the default lunMap maping for LinearInversePixelMapping func.
	LumPixFloor []float64
	ncpu        = runtime.NumCPU()
)

func init() {
	LumPixFloor = make([]float64, LumSize)
	for p := 1; p < LumSize; p++ {
		LumPixFloor[p] = float64(p-1) / RangeMax
	}
}

// A ToneMappingOperator is an algorithm that converts hdr.Image to image.Image.
//
//
// HDR is a high dynamic range (HDR) technique used in imaging and photography
// to reproduce a greater dynamic range of luminosity than is possible
// with standard digital imaging or photographic techniques.
// The aim is to present a similar range of luminance to that experienced
// through the human visual system.
// The human eye, through adaptation of the iris and other methods,
// adjusts constantly to adapt to a broad range of luminance present in the environment.
// The brain continuously interprets this information so that a viewer can see in a wide range of light conditions.
type ToneMappingOperator interface {
	// Perform runs the TMO mapping.
	Perform() (image.Image, error)
}

func parallelR(r image.Rectangle, f func(x1, y1, x2, y2 int)) chan struct{} {
	return parallel(r.Dx(), r.Dy(), f)
}

func parallel(width, height int, f func(x1, y1, x2, y2 int)) chan struct{} {
	// FIXME use context
	wg := &sync.WaitGroup{}
	completed := make(chan struct{})

	for _, rect := range hdr.Split(0, 0, width, height, ncpu) {
		wg.Add(1)
		go func(rect image.Rectangle) {
			defer wg.Done()

			f(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)

		}(rect)
	}

	go func() {
		wg.Wait()
		close(completed)
	}()

	return completed
}

// LinearInversePixelMapping is an linear inverse pixel mapping.
// It is preference to have slightly more solid black 0 and solid white RangeMax in spectrum
// by stretching a mapping.
func LinearInversePixelMapping(lum float64, lumMap []float64, lumSize int) float64 {
	rangeLow, rangeUp := 0, lumSize

	for {
		rangeMid := (rangeLow + rangeUp) / 2

		if rangeMid == rangeLow {
			return float64(rangeLow) // Avoid conversion by the caller.
		}

		if lum < lumMap[rangeMid] {
			rangeUp = rangeMid
		} else {
			rangeLow = rangeMid
		}
	}
}

// Clamp to solid black and solid white.
func Clamp(channel float64) float64 {
	if channel < RangeMin {
		channel = RangeMin
	}
	if channel > RangeMax {
		channel = RangeMax
	}
	return channel
}

//--------------------------------------//
// MinMax data                          //
//--------------------------------------//

type minmax struct {
	min float64
	max float64
}

func newMinMax() *minmax {
	return &minmax{
		min: math.Inf(1),
		max: math.Inf(-1),
	}
}

func (mm *minmax) update(v float64) {
	mm.min = math.Min(mm.min, v)
	mm.max = math.Max(mm.max, v)
}

func (mm *minmax) String() string {
	return fmt.Sprintf("min: %f, max: %f", mm.min, mm.max)
}
