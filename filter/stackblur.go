package filter

import (
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// HDR port of https://github.com/esimov/stackblur-go

// A StackBlur is a fast almost Gaussian Blur implementation.
type StackBlur struct {
	HDRImage hdr.Image
	Radius   int
}

// NewStackBlur instanciates a new StackBlur.
func NewStackBlur(m hdr.Image, radius int) *StackBlur {
	return &StackBlur{
		HDRImage: m,
		Radius:   radius,
	}
}

// Perform runs the blur filter and return the filtered image.
func (f *StackBlur) Perform() hdr.Image {
	m := hdr.Copy(f.HDRImage)
	ms := m.(hdr.ImageSet)

	var stackEnd, stackIn, stackOut *blurStack
	var width, height = m.Bounds().Dx(), m.Bounds().Dy()
	var (
		div, widthMinus1, heightMinus1, radiusPlus1, sumFactor, p int
		rSum, gSum, bSum,
		rOutSum, gOutSum, bOutSum,
		rInSum, gInSum, bInSum,
		pr, pg, pb float64
	)

	div = f.Radius + f.Radius + 1
	widthMinus1 = width - 1
	heightMinus1 = height - 1
	radiusPlus1 = f.Radius + 1
	sumFactor = radiusPlus1 * (radiusPlus1 + 1) / 2

	bs := blurStack{}
	stackStart := bs.NewBlurStack()
	stack := stackStart

	for i := 1; i < div; i++ {
		stack.next = bs.NewBlurStack()
		stack = stack.next
		if i == radiusPlus1 {
			stackEnd = stack
		}
	}
	stack.next = stackStart

	// mulSum := mulTable[f.Radius]
	// shgSum := shgTable[f.Radius]
	divsum := float64((div + 1) >> 1)
	divsum *= divsum

	for y := 0; y < height; y++ {
		rInSum, gInSum, bInSum, rSum, gSum, bSum = 0, 0, 0, 0, 0, 0

		pr, pg, pb, _ = m.HDRAt(0, y).HDRRGBA()

		rOutSum = float64(radiusPlus1) * pr
		gOutSum = float64(radiusPlus1) * pg
		bOutSum = float64(radiusPlus1) * pb

		rSum += float64(sumFactor) * pr
		gSum += float64(sumFactor) * pg
		bSum += float64(sumFactor) * pb

		stack = stackStart

		for i := 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack = stack.next
		}

		for i := 1; i < radiusPlus1; i++ {
			p = i
			if widthMinus1 < i {
				p = widthMinus1
			}

			pr, pg, pb, _ = m.HDRAt(p, y).HDRRGBA()

			stack.r = pr
			stack.g = pg
			stack.b = pb

			rSum += stack.r * float64(radiusPlus1-i)
			gSum += stack.g * float64(radiusPlus1-i)
			bSum += stack.b * float64(radiusPlus1-i)

			rInSum += pr
			gInSum += pg
			bInSum += pb

			stack = stack.next
		}

		stackIn = stackStart
		stackOut = stackEnd

		for x := 0; x < width; x++ {
			// ms.Set(x, y, hdrcolor.RGB{
			// 	R: (rSum * mulSum) / math.Pow(2, shgSum),
			// 	G: (gSum * mulSum) / math.Pow(2, shgSum),
			// 	B: (bSum * mulSum) / math.Pow(2, shgSum),
			// })
			ms.Set(x, y, hdrcolor.RGB{
				R: rSum / divsum,
				G: gSum / divsum,
				B: bSum / divsum,
			})

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b

			p = x + f.Radius + 1

			if p > widthMinus1 {
				p = widthMinus1
			}

			stackIn.r, stackIn.g, stackIn.b, _ = m.HDRAt(p, y).HDRRGBA()

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb

			stackOut = stackOut.next
		}
	}

	for x := 0; x < width; x++ {
		rInSum, gInSum, bInSum, rSum, gSum, bSum = 0, 0, 0, 0, 0, 0

		pr, pg, pb, _ = m.HDRAt(x, 0).HDRRGBA()

		rOutSum = float64(radiusPlus1) * pr
		gOutSum = float64(radiusPlus1) * pg
		bOutSum = float64(radiusPlus1) * pb

		rSum += float64(sumFactor) * pr
		gSum += float64(sumFactor) * pg
		bSum += float64(sumFactor) * pb

		stack = stackStart

		for i := 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack = stack.next
		}

		for i := 1; i <= f.Radius; i++ {
			pr, pg, pb, _ = m.HDRAt(x, i).HDRRGBA()

			stack.r = pr
			stack.g = pg
			stack.b = pb

			rSum += stack.r * float64(radiusPlus1-i)
			gSum += stack.g * float64(radiusPlus1-i)
			bSum += stack.b * float64(radiusPlus1-i)

			rInSum += pr
			gInSum += pg
			bInSum += pb

			stack = stack.next
		}

		stackIn = stackStart
		stackOut = stackEnd

		for y := 0; y < height; y++ {
			// ms.Set(x, y, hdrcolor.RGB{
			// 	R: (rSum * mulSum) / math.Pow(2, shgSum),
			// 	G: (gSum * mulSum) / math.Pow(2, shgSum),
			// 	B: (bSum * mulSum) / math.Pow(2, shgSum),
			// })
			ms.Set(x, y, hdrcolor.RGB{
				R: rSum / divsum,
				G: gSum / divsum,
				B: bSum / divsum,
			})

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b

			p = y + radiusPlus1
			if p > heightMinus1 {
				p = heightMinus1
			}

			stackIn.r, stackIn.g, stackIn.b, _ = m.HDRAt(x, p).HDRRGBA()

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb

			stackOut = stackOut.next
		}
	}

	return m
}

//-----------------------------------------//
//                                         //
// stack                                   //
//                                         //
//-----------------------------------------//

// blurStack is a linked list containing the color value and a pointer to the next struct.
type blurStack struct {
	r, g, b float64
	next    *blurStack
}

// var mulTable = []float64{
// 	512, 512, 456, 512, 328, 456, 335, 512, 405, 328, 271, 456, 388, 335, 292, 512,
// 	454, 405, 364, 328, 298, 271, 496, 456, 420, 388, 360, 335, 312, 292, 273, 512,
// 	482, 454, 428, 405, 383, 364, 345, 328, 312, 298, 284, 271, 259, 496, 475, 456,
// 	437, 420, 404, 388, 374, 360, 347, 335, 323, 312, 302, 292, 282, 273, 265, 512,
// 	497, 482, 468, 454, 441, 428, 417, 405, 394, 383, 373, 364, 354, 345, 337, 328,
// 	320, 312, 305, 298, 291, 284, 278, 271, 265, 259, 507, 496, 485, 475, 465, 456,
// 	446, 437, 428, 420, 412, 404, 396, 388, 381, 374, 367, 360, 354, 347, 341, 335,
// 	329, 323, 318, 312, 307, 302, 297, 292, 287, 282, 278, 273, 269, 265, 261, 512,
// 	505, 497, 489, 482, 475, 468, 461, 454, 447, 441, 435, 428, 422, 417, 411, 405,
// 	399, 394, 389, 383, 378, 373, 368, 364, 359, 354, 350, 345, 341, 337, 332, 328,
// 	324, 320, 316, 312, 309, 305, 301, 298, 294, 291, 287, 284, 281, 278, 274, 271,
// 	268, 265, 262, 259, 257, 507, 501, 496, 491, 485, 480, 475, 470, 465, 460, 456,
// 	451, 446, 442, 437, 433, 428, 424, 420, 416, 412, 408, 404, 400, 396, 392, 388,
// 	385, 381, 377, 374, 370, 367, 363, 360, 357, 354, 350, 347, 344, 341, 338, 335,
// 	332, 329, 326, 323, 320, 318, 315, 312, 310, 307, 304, 302, 299, 297, 294, 292,
// 	289, 287, 285, 282, 280, 278, 275, 273, 271, 269, 267, 265, 263, 261, 259,
// }

// var shgTable = []float64{
// 	9, 11, 12, 13, 13, 14, 14, 15, 15, 15, 15, 16, 16, 16, 16, 17,
// 	17, 17, 17, 17, 17, 17, 18, 18, 18, 18, 18, 18, 18, 18, 18, 19,
// 	19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 20, 20, 20,
// 	20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 21,
// 	21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
// 	21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 22, 22, 22, 22, 22, 22,
// 	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
// 	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 23,
// 	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
// 	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
// 	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
// 	23, 23, 23, 23, 23, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
// 	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
// 	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
// 	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
// 	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
// }

// NewBlurStack is a constructor function returning a new struct of type blurStack.
func (f *blurStack) NewBlurStack() *blurStack {
	return &blurStack{r: f.r, g: f.g, b: f.b, next: f.next}
}
