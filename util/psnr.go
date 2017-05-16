package util

import (
	"math"

	"github.com/mdouchement/hdr"
)

// PSNR computes the Peak signal-to-noise ratio between 2 images.
// The PSNR is most commonly used to measure the quality of 2 compressed images.
// - 100 dB means that img2 has the same quality as img1.
// - 0 dB means that you do not compare the same image.
func PSNR(img1, img2 hdr.Image) (mse, snr, psnr, peak float64) {
	d := img1.Bounds()
	var signal float64
	var noise float64

	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			x1, y1, z1, _ := img1.HDRAt(x, y).HDRXYZA()
			x2, y2, z2, _ := img2.HDRAt(x, y).HDRXYZA()

			signal += x1*x1 + y1*y1 + z1*z1

			noise += math.Pow(x1-x2, 2)
			noise += math.Pow(y1-y2, 2)
			noise += math.Pow(z1-z2, 2)

			peak = math.Max(peak, x1)
			peak = math.Max(peak, y1)
			peak = math.Max(peak, z1)
		}
	}

	mse = noise / float64(d.Dx()*d.Dy())
	snr = 10 * math.Log10(signal/noise)
	psnr = 10 * math.Log10(peak*peak/mse)
	psnr = math.Min(psnr, 100) // Max quality is 100 dB, in this case MSE should equal 0.

	return
}
