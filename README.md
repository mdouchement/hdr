# HDR - High Dynamic Range

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/hdr)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/hdr)](https://goreportcard.com/report/github.com/mdouchement/hdr)
[![License](https://img.shields.io/github/license/mdouchement/hdr.svg)](http://opensource.org/licenses/MIT)

HDR is a library that handles RAW image format written with Golang.

It aims to provide tools to read [HDR](https://en.wikipedia.org/wiki/High-dynamic-range_imaging) files and convert it to a LDR (Low Dynamic Range, aka PNG/JPEG/etc.) in an `image.Image` object.


Documentations:
 - http://www.anyhere.com/gward/hdrenc/hdr_encodings.html


## Supported file formats

- Radiance RGBE/XYZE
- CRAD, homemade HDR file format

## Supported tone mapping operators

Read this [documentation](http://osp.wikidot.com/parameters-for-photographers) to find what TMO use.

Read this [documentation](https://hal.archives-ouvertes.fr/hal-00724931/document) to understand what is a TMO.

- Linear
- Logarithmic
- Normalization (a part of iCAM06 TMO)
- Drago '03    - Adaptive logarithmic mapping for displaying high contrast scenes
- Reinhard '05 - Photographic tone reproduction for digital images
  - Playing with parameters could provide better rendering
- Custom Reinhard '05
	- Rendering looks like a JPEG photo taken with a smartphone
- iCAM06       - A refined image appearance model for HDR image rendering

## Usage

```sh
go get github.com/mdouchement/hdr
```

```go
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"time"

	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/rgbe"
	"github.com/mdouchement/hdr/tmo"
)

// Samples:
//
// http://www.anyhere.com/gward/hdrenc/pages/originals.html
// http://resources.mpi-inf.mpg.de/tmo/logmap/ (High Contrast Scenes)

var (
	// input = "/Users/mdouchement/tmp/hdr/memorial_o876.hdr"
	// input = "/Users/mdouchement/tmp/hdr/MtTamWest_o281.hdr"
	// input = "/Users/mdouchement/tmp/hdr/rend02_oC95.hdr"
	// input = "/Users/mdouchement/tmp/hdr/Tree_oAC1.hdr"
	input  = "/Users/mdouchement/tmp/hdr/Apartment_float_o15C.hdr"
	output = "/Users/mdouchement/tmp/hdr/output.png"
)

func main() {
	fmt.Printf("Using %d CPUs\n", runtime.NumCPU())

	fi, err := os.Open(input)
	check(err)
	defer fi.Close()

	start := time.Now()

	m, fname, err := image.Decode(fi)
	check(err)

	fmt.Printf("Read image (%s) took %v\n", fname, time.Since(start))

	if hdrm, ok := m.(hdr.Image); ok {
		startTMO := time.Now()

		// t := tmo.NewLinear(hdrm)
		// t := tmo.NewLogarithmic(hdrm)
		// t := tmo.NewICam06Normalization(hdrm)
		// t := tmo.NewDefaultDrago03(hdrm)
		// t := tmo.NewDefaultCustomReinhard05(hdrm)
		t := tmo.NewDefaultReinhard05(hdrm)
		m = t.Perform()

		fmt.Println("Apply TMO took", time.Since(startTMO))
	}

	fo, err := os.Create(output)
	check(err)

	png.Encode(fo, m)

	fmt.Println("Total", time.Since(start))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
```


## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request

As possible, run the following commands to format and lint the code:

```sh
# Format
find . -name '*.go' -not -path './vendor*' -exec gofmt -s -w {} \;

# Lint
gometalinter --config=gometalinter.json ./...
```
