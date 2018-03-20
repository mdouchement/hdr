package filter

import (
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// HDR port of https://github.com/esimov/stackblur-go

// StackBlur performs a fast almost Gaussian Blur implementation.
func StackBlur(src hdr.Image, radius int) hdr.Image {
	m := hdr.Copy(src)
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

	div = radius + radius + 1
	widthMinus1 = width - 1
	heightMinus1 = height - 1
	radiusPlus1 = radius + 1
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

			p = x + radius + 1

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

		for i := 1; i <= radius; i++ {
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

// NewBlurStack is a constructor function returning a new struct of type blurStack.
func (f *blurStack) NewBlurStack() *blurStack {
	return &blurStack{r: f.r, g: f.g, b: f.b, next: f.next}
}
