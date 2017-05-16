package util

import (
	"math"

	"github.com/mdouchement/hdr"
)

var (
	// C1 covariance coeff for SSIM img1
	C1 = math.Pow(0.01, 2)
	// C2 covariance coeff for SSIM img2
	C2 = math.Pow(0.03, 2)
)

// SSIM computes the Structural SIMilarity Index between 2 images.
// The SSIM is most commonly used to measure the quality of 2 compressed images.
// - 1 means that img2 has the same quality as img1.
// - 0 means that you do not compare the same image.
func SSIM(img1, img2 hdr.Image) (ssim float64) {
	d := img1.Bounds()
	var s1 float64
	var s2 float64
	var ss float64
	var s12 float64

	for y := 0; y < d.Dy(); y++ {
		for x := 0; x < d.Dx(); x++ {
			x1, y1, z1, _ := img1.HDRAt(x, y).HDRXYZA()
			x2, y2, z2, _ := img2.HDRAt(x, y).HDRXYZA()

			sum1 := x1 + y1 + z1
			sum2 := x2 + y2 + z2

			s1 += sum1
			s2 += sum2

			ss += sum1*sum1 + sum2*sum2

			s12 += sum1 * sum2
		}
	}

	vari := ss - s1*s1 - s2*s2
	covar := s12 - s1*s2

	return (2*s1*s2 + C1) * (2*covar + C2) / ((s1*s1 + s2*s2 + C1) * (vari + C2))
}
