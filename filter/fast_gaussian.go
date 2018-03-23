package filter

import (
	"math"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// fast gaussian blur based on http://blog.ivank.net/fastest-gaussian-blur.html
// and Golang implementation https://github.com/tajtiattila/blur

// FastGaussian blurs im using a fast approximation of gaussian blur.
// The algorithm has a computational complexity independent of radius.
func FastGaussian(src hdr.Image, radius int) hdr.Image {
	boxes := determineBoxes(float64(radius), 3)
	tmp := hdr.EmptyAs(src)
	dst := hdr.EmptyAs(src)
	boxBlur4(dst, tmp, src, (boxes[0]-1)/2)
	boxBlur4(dst, tmp, dst, (boxes[1]-1)/2)
	boxBlur4(dst, tmp, dst, (boxes[2]-1)/2)

	return dst
}

func boxBlur4(dst, scratch, src hdr.Image, radius int) {
	if src == scratch || dst == scratch {
		panic("scratch must be different than src and dst")
	}
	boxBlurH(scratch.(hdr.ImageSet), src, radius)
	boxBlurV(dst.(hdr.ImageSet), scratch, radius)
}

func boxBlurH(dst hdr.ImageSet, src hdr.Image, radius int) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	r1 := radius + 1
	r1f := float64(r1)
	r2f := float64(2*radius + 1)
	var vr, vg, vb float64

	for y := 0; y < h; y++ {
		fvr, fvg, fvb, _ := src.HDRAt(0, y).HDRRGBA()
		lvr, lvg, lvb, _ := src.HDRAt(w-1, y).HDRRGBA()

		vr = r1f * fvr
		vg = r1f * fvg
		vb = r1f * fvb

		for x := 0; x < radius; x++ {
			r, g, b, _ := src.HDRAt(x, y).HDRRGBA()
			vr += r
			vg += g
			vb += b
		}

		for x := 0; x < r1; x++ {
			r, g, b, _ := src.HDRAt(x+radius, y).HDRRGBA()
			vr += r - fvr
			vg += g - fvg
			vb += b - fvb

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})
		}

		for x := r1; x < w-radius; x++ {
			r, g, b, _ := src.HDRAt(x+radius, y).HDRRGBA()
			r1, g1, b1, _ := src.HDRAt(x-r1, y).HDRRGBA()

			vr += r - r1
			vg += g - g1
			vb += b - b1

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})
		}

		for x := w - radius; x < w; x++ {
			r, g, b, _ := src.HDRAt(x-r1, y).HDRRGBA()

			vr += lvr - r
			vg += lvg - g
			vb += lvb - b

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})
		}
	}
}

func boxBlurV(dst hdr.ImageSet, src hdr.Image, radius int) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()

	r1 := radius + 1
	r1f := float64(r1)
	r2f := float64(2*radius + 1)
	var vr, vg, vb float64

	for x := 0; x < w; x++ {
		fvr, fvg, fvb, _ := src.HDRAt(x, 0).HDRRGBA()
		lvr, lvg, lvb, _ := src.HDRAt(x, h-1).HDRRGBA()

		vr = r1f * fvr
		vg = r1f * fvg
		vb = r1f * fvb

		for y := 0; y < radius; y++ {
			r, g, b, _ := src.HDRAt(x, y).HDRRGBA()
			vr += r
			vg += g
			vb += b
		}

		for y := 0; y < r1; y++ {
			r, g, b, _ := src.HDRAt(x, y+radius).HDRRGBA()
			vr += r - fvr
			vg += g - fvg
			vb += b - fvb

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})
		}
		for y := r1; y < h-radius; y++ {
			r, g, b, _ := src.HDRAt(x, y+radius).HDRRGBA()
			r1, g1, b1, _ := src.HDRAt(x, y-r1).HDRRGBA()

			vr += r - r1
			vg += g - g1
			vb += b - b1

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})
		}

		for y := h - radius; y < h; y++ {
			r, g, b, _ := src.HDRAt(x, y-r1).HDRRGBA()

			vr += lvr - r
			vg += lvg - g
			vb += lvb - b

			dst.Set(x, y, hdrcolor.RGB{R: vr / r2f, G: vg / r2f, B: vb / r2f})

		}
	}
}

func determineBoxes(sigma float64, nbox int) []int {
	// standard deviation, number of boxes
	idealWeight := math.Sqrt((12 * sigma * sigma / float64(nbox)) + 1)
	wlo := int(math.Floor(idealWeight))
	if wlo%2 == 0 {
		wlo--
	}
	wup := wlo + 2

	idealMedian := (12*sigma*sigma - float64(nbox*wlo*wlo+4*nbox*wlo+3*nbox)) / (-4*float64(wlo) - 4)
	median := int(math.Floor(idealMedian + 0.5))

	boxsizes := make([]int, nbox)
	for i := range boxsizes {
		if i < median {
			boxsizes[i] = wlo
		} else {
			boxsizes[i] = wup
		}
	}
	return boxsizes
}
