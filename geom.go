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
	b := true

	for y := n/3 + 1; y > 0; y-- {
		for x := n; x > 1; x-- {
			if x*y == n {
				nx = x
				ny = y

				// Exit loops
				y = 1
				break
			}
			if !b && x*y < n {
				nx = x
				ny = y

				// Try to find better combination
				b = false
			}
		}
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
