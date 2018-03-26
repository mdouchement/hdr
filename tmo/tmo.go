package tmo

import (
	"fmt"
	"image"
	"math"
	"sort"
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

// WoB returns 1.0 if b is true, else 0.0
func WoB(b bool) float64 {
	if b {
		return 1
	}
	return 0
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

//-----------------//
// Percentile      //
//-----------------//

type percentiles []float64

func (p percentiles) sort() {
	sort.Sort(p)
}

func (p percentiles) percentile(clipping float64) float64 {
	n := float64(len(p))
	i := int(clipping * n)
	return float64(p[i])
}

func (p percentiles) Len() int {
	return len(p)
}

func (p percentiles) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p percentiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
