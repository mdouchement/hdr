# RGBE - Radiance RGBE

A RGBE codec reader for Golang.


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

## License

**BSD-style**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
