# HDR - High Dynamic Range

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/hdr)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/hdr)](https://goreportcard.com/report/github.com/mdouchement/hdr)
[![License](https://img.shields.io/github/license/mdouchement/hdr.svg)](http://opensource.org/licenses/MIT)

HDR is a library that handles RAW image format written with Golang.

It aims to provide tools to read [HDR](https://en.wikipedia.org/wiki/High-dynamic-range_imaging) files and convert it to a LDR (Low Dynamic Range, aka PNG/JPEG/etc.) in an `image.Image` object.


## Supported file formats

- Radiance RGBE (read-only)

## Supported tone mapping operators

- Linear (a naive TMO implementation)
- Reinhard '05 Tone Mapping Operator

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
