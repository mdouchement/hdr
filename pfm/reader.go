package pfm

// Resources:
// http://www.pauldebevec.com/Research/HDR/PFM/
// http://netpbm.sourceforge.net/doc/pfm.html

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/format"
	"github.com/mdouchement/hdr/hdrcolor"
)

type decoder struct {
	r          io.Reader
	config     image.Config
	scale      float64 // Scale Factor
	mode       imageMode
	endianness endianness
}

func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r:     bufio.NewReader(r),
		scale: 1.0,
	}

	return d, d.parseHeader()
}

//--------------------------------------//
// Header parser                        //
//--------------------------------------//

func (d *decoder) parseHeader() error {
	for i := 0; i < 3; i++ {
		token, err := readUntil(d.r, '\n')
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)

		switch i {
		case 0:
			if token == header0 {
				// Header found
				d.mode = mColor
				d.config.ColorModel = hdrcolor.RGBModel
			} else if token == header1 {
				// Header found
				d.mode = mGrayscale
			} else {
				return FormatError("format not compatible")
			}
		case 1:
			if n, err := fmt.Sscanf(token, "%d %d", &d.config.Width, &d.config.Height); n < 2 || err != nil {
				return FormatError("missing image size specifier")
			}
		case 2:
			scale, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return FormatError("missing Scale Factor / Endianness specifier")
			}
			if scale < 0 {
				d.endianness = eLittleEndian
			} else {
				d.endianness = eBigEndian
			}
			d.scale = math.Abs(scale)
		}
	}

	return nil
}

//--------------------------------------//
// Reader                               //
//--------------------------------------//

// DecodeConfig returns the color model and dimensions of a PFM image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	d, err := newDecoder(r)
	if err != nil {
		return image.Config{}, err
	}
	return d.config, nil
}

// Decode reads a HDR image from r and returns an image.Image.
func Decode(r io.Reader) (img image.Image, err error) {
	d, err := newDecoder(r)
	if err != nil {
		return nil, err
	}

	imgRect := image.Rect(0, 0, d.config.Width, d.config.Height)
	switch d.mode {
	case mColor:
		img = hdr.NewRGB(imgRect)
	case mGrayscale:
		err = UnsupportedError("image mode Grayscale")
	default:
		err = UnsupportedError("image mode")
		return
	}

	m := img.(*hdr.RGB) // Only RGB format is supported

	pixel := make([]byte, 4*3) // RGB pixel (4 Bytes per channel)
	var R, G, B float64
	invScale := 1 / d.scale
	var endianFunc func(pixel []byte) (float64, float64, float64)
	if d.endianness == eLittleEndian {
		endianFunc = format.FromBytes
	} else {
		endianFunc = format.FromBytesBE
	}

	// The pixels in each row ordered left to right and the rows ordered bottom to top
	for y := d.config.Height - 1; y >= 0; y-- {
		for x := 0; x < d.config.Width; x++ {
			if _, err = io.ReadFull(d.r, pixel); err != nil {
				return
			}

			R, G, B = endianFunc(pixel)
			m.Set(x, y, hdrcolor.RGB{
				R: R * invScale,
				G: G * invScale,
				B: B * invScale,
			})
		}
	}

	return
}

func init() {
	image.RegisterFormat("pfm", header0, Decode, DecodeConfig)
	image.RegisterFormat("pfm", header1, Decode, DecodeConfig)
}
