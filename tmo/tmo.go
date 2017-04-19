package tmo

import (
	"image"
	"runtime"
	"sync"
)

const (
	// RangeMin is the LDR lower boundary.
	RangeMin = 0
	// RangeMax is the LDR higher boundary.
	RangeMax = 65535 // max uint16
)

var ncpu = runtime.NumCPU()

func init() {
	runtime.GOMAXPROCS(ncpu)
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

func parallel(width, height int, f func(x1, y1, x2, y2 int)) chan struct{} {
	// FIXME use context
	wg := &sync.WaitGroup{}
	completed := make(chan struct{})

	for _, rect := range split(0, 0, width, height) {
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

// split image
func split(x1, y1, x2, y2 int) []image.Rectangle {
	switch ncpu {
	case 2:
		ym := (y1 + y2) / 2
		return []image.Rectangle{
			{image.Point{x1, y1}, image.Point{x2, ym}},
			{image.Point{x1, ym}, image.Point{x2, y2}},
		}
	case 4:
		xm := (x1 + x2) / 2
		ym := (y1 + y2) / 2
		return []image.Rectangle{
			{image.Point{x1, y1}, image.Point{xm, ym}},
			{image.Point{xm, y1}, image.Point{x2, ym}},
			{image.Point{x1, ym}, image.Point{xm, y2}},
			{image.Point{xm, ym}, image.Point{x2, y2}},
		}
	case 6:
		xm := (x1 + x2) / 2
		ym := (y1 + y2) / 3
		return []image.Rectangle{
			{image.Point{x1, y1}, image.Point{xm, ym}},
			{image.Point{xm, y1}, image.Point{x2, ym}},
			{image.Point{x1, ym}, image.Point{xm, 2 * ym}},
			{image.Point{xm, ym}, image.Point{x2, 2 * ym}},
			{image.Point{x1, 2 * ym}, image.Point{xm, y2}},
			{image.Point{xm, 2 * ym}, image.Point{x2, y2}},
		}
	default: // 8 and more
		xm := (x1 + x2) / 2
		ym := (y1 + y2) / 4
		return []image.Rectangle{
			{image.Point{x1, y1}, image.Point{xm, ym}},
			{image.Point{xm, y1}, image.Point{x2, ym}},
			{image.Point{x1, ym}, image.Point{xm, 2 * ym}},
			{image.Point{xm, ym}, image.Point{x2, 2 * ym}},
			{image.Point{x1, 2 * ym}, image.Point{xm, 3 * ym}},
			{image.Point{xm, 2 * ym}, image.Point{x2, 3 * ym}},
			{image.Point{x1, 3 * ym}, image.Point{xm, y2}},
			{image.Point{xm, 3 * ym}, image.Point{x2, y2}},
		}
	}
}
