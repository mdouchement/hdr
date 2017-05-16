package util

import (
	"image"
	"runtime"
	"sync"

	"github.com/mdouchement/hdr"
)

var ncpu = runtime.NumCPU()

// ParallelR runs Parallel with the given r boundaries.
func ParallelR(r image.Rectangle, f func(x1, y1, x2, y2 int)) chan struct{} {
	return Parallel(r.Dx(), r.Dy(), f)
}

// Parallel runs f in runtime.NumCPU() parallel tiles.
func Parallel(width, height int, f func(x1, y1, x2, y2 int)) chan struct{} {
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
