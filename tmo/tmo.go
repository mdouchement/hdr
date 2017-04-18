package tmo

import (
	"image"
)

const (
	// RangeMin is the LDR lower boundary.
	RangeMin = 0
	// RangeMax is the LDR higher boundary.
	RangeMax = 65535 // max uint16
)

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

// Inverse pixel mapping
func pixelBinarySearch(lum float64, lumMap []float64, lumSize int) float64 {
	rangeLow, rangeMid, rangeUp := 0, 0, lumSize

	for {
		rangeMid = (rangeLow + rangeUp) / 2

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
