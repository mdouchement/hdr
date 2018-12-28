# RGBE - Radiance RGBE/XYZE

A RGBE codec for Golang.


## Usage

```go
package main

import (
	"image"
	"image/png"
	"os"

	_ "github.com/mdouchement/hdrimage/rgbe"
)

var (
	input  = "/tmp/IMG_0020.rgbe"
	output = "/tmp/IMG_0020.png"
)

func main() {
	fi, err := os.Open(input)
	check(err)
	defer fi.Close()

	m, _, err := image.Decode(fi)
	check(err)

	fo, err := os.Create(output)
	check(err)

	png.Encode(fo, m)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
```
