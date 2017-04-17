# HDR - High Dynamic Range

HDR is a library that handles RAW image format written with Golang.

It aims to provide tools to read [HDR](https://en.wikipedia.org/wiki/High-dynamic-range_imaging) files and convert it to a LDR (Low Dynamic Range, aka PNG/JPEG/etc.) in an `image.Image` object.


## Supported file formats

- Radiance RGBE (read-only)

## Supported tone mapping operators

- Linear (a very naive TMO implementation)

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
	output = "/Users/fragan/tmp/hdr/output.png"
)

func main() {
	fi, err := os.Open(input)
	check(err)
	defer fi.Close()

	m, fname, err := image.Decode(fi)
	check(err)
	fmt.Println("FName:", fname)

	if hdrm, ok := m.(hdr.Image); ok {
    // Here is the conversion from HDR to LDR.
		lin := tmo.NewLinear(hdrm)
		m = lin.Perform()
	}

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


## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
