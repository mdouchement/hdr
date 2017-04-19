package hdr

import "image"

// Split tries to split the given rectangle coordinates in n tiles.
func Split(x1, y1, x2, y2, n int) []image.Rectangle {
	return SplitWithRectangle(image.Rectangle{image.Point{x1, y1}, image.Point{x2, y2}}, n)
}

// SplitWithRectangle tries to split the given rectangle (image) in n tiles.
func SplitWithRectangle(r image.Rectangle, n int) []image.Rectangle {
	if n < 2 {
		return []image.Rectangle{r}
	}

	var nx int
	var ny int

	// TODO find a function to calculate this dynamically.
	switch n {
	case 2:
		fallthrough
	case 3:
		nx = 1 // 0 in fact but this avoid integer divide by zero error
		ny = 2
	case 4:
		// Could be used for this case:
		// sqrt := math.Sqrt(float64(n))
		// integer, fractional := math.Modf(sqrt)
		// if fractional == 0 { ny = nx = int(integer) }
		fallthrough
	case 5:
		nx = 2
		ny = 2
	case 6:
		fallthrough
	case 7:
		nx = 2
		ny = 3
	case 8:
		nx = 2
		ny = 4
	case 9:
		// Could be used for this case:
		// sqrt := math.Sqrt(float64(n))
		// integer, fractional := math.Modf(sqrt)
		// if fractional == 0 { ny = nx = int(integer) }
		nx = 3
		ny = 3
	case 10:
		fallthrough
	case 11:
		nx = 2
		ny = 5
	case 12:
		fallthrough
	case 13:
		nx = 3
		ny = 4
	case 14:
		nx = 2
		ny = 7
	case 15:
		nx = 3
		ny = 5
	default: // 16
		// Could be used for this case:
		// sqrt := math.Sqrt(float64(n))
		// integer, fractional := math.Modf(sqrt)
		// if fractional == 0 { ny = nx = int(integer) }
		nx = 4
		ny = 4
	}

	tileWidth := (r.Min.X + r.Max.X) / nx
	tileHeight := (r.Min.Y + r.Max.Y) / ny

	splits := make([]image.Rectangle, 0, n*n)
	for y := 0; y < ny; y++ {
		for x := 0; x < nx; x++ {
			splits = append(splits, image.Rectangle{
				image.Point{tileWidth * x, tileHeight * y},
				image.Point{tileWidth*x + tileWidth, tileHeight*y + tileHeight},
			})
		}
	}

	return splits
}
