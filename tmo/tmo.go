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
