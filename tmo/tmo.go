package tmo

import (
	"image"
)

const (
	RangeMin = 0
	RangeMax = 255
)

type ToneMappingOperator interface {
	// Perform runs the TMO mapping.
	Perform() (image.Image, error)
}
