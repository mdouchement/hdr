package rgbe

// Resources:
// http://www.graphics.cornell.edu/~bjw/rgbe
// http://www.anyhere.com/gward/hdrenc/pages/originals.html (samples)
// http://radsite.lbl.gov/radiance/framed.html (samples)

import (
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// A FormatError reports that the input is not a valid RGBE image.
type FormatError string

func (e FormatError) Error() string {
	return fmt.Sprintf("rgbe: invalid format: %s", e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return fmt.Sprintf("rgbe: unsupported feature: %s", e)
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return fmt.Sprintf("rgbe: internal error: %s", e)
}

type decoder struct {
	r        io.Reader
	config   image.Config
	exposure float64
	mode     imageMode
}

func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r:        r,
		exposure: 1.0,
	}

	return d, d.parseHeader()
}

//--------------------------------------//
// Header parser                        //
//--------------------------------------//

func (d *decoder) parseHeader() error {
	magic := false

	for {
		token, err := readUntil(d.r, '\n')
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)

		if token == "" {
			// End of header
			break
		}
		if token == "#?RADIANCE" || token == "#?RGBE" || token == "#?AUTOPANO" {
			// Format specifier found (magic number)
			magic = true
		}
		if strings.HasPrefix(token, "#") {
			// Skip commented line
			continue
		}
		if token == "FORMAT=32-bit_rle_rgbe" {
			// Header found
			d.mode = mRGBE
			d.config.ColorModel = hdrcolor.RGBModel
			continue
		}
		if token == "FORMAT=32-bit_rle_xyze" {
			return UnsupportedError("32-bit_rle_xyze format")
		}
		if strings.HasPrefix(token, "EXPOSURE=") {
			if n, err := fmt.Sscanf(token, "EXPOSURE=%f", &d.exposure); n < 1 || err != nil {
				return FormatError("invalid exposure specifier")
			}
		}
	}

	// ignore weird exposure adjustments
	if d.exposure > 1e12 || d.exposure < 1e-12 {
		d.exposure = 1.0
	}

	if !magic {
		return FormatError("format not compatible")
	}

	// image size
	token, err := readUntil(d.r, '\n')
	if err != nil {
		return err
	}
	if n, err := fmt.Sscanf(token, "-Y %d +X %d", &d.config.Height, &d.config.Width); n < 2 || err != nil {
		return FormatError("missing image size specifier")
	}

	return nil
}

//--------------------------------------//
// Pixels parser                        //
//--------------------------------------//

func (d *decoder) decode(dst image.Image, y int, scanline []byte) {
	for x := 0; x < d.config.Width; x++ {
		r, g, b := rgbeToRGB(
			scanline[4*x],
			scanline[4*x+1],
			scanline[4*x+2],
			scanline[4*x+3],
			d.exposure)

		switch d.mode {
		case mRGBE:
			img := dst.(*hdr.RGB)
			img.SetRGB(x, y, hdrcolor.RGB{R: r, G: g, B: b})
		}
	}
}

func (d *decoder) decodeRLE(dst image.Image, y int, scanline []byte) {
	for x := 0; x < d.config.Width; x++ {
		r, g, b := rgbeToRGB(
			scanline[x],
			scanline[x+d.config.Width],
			scanline[x+d.config.Width*2],
			scanline[x+d.config.Width*3],
			d.exposure)

		switch d.mode {
		case mRGBE:
			img := dst.(*hdr.RGB)
			img.SetRGB(x, y, hdrcolor.RGB{R: r, G: g, B: b})
		}
	}
}

func (d *decoder) readRLE(scanline []byte) (err error) {
	buf := make([]byte, 2)

	// --- each channel is encoded separately
	for ch := 0; ch < 4; ch++ {
		index := d.config.Width * ch
		peek := 0
		for peek < d.config.Width {

			// Read RLE
			if _, err = io.ReadFull(d.r, buf); err != nil {
				if err == io.EOF {
					err = nil
				}
				return
			}

			if buf[0] > 128 {
				// a run of the same value
				runLength := int(buf[0]) - 128
				for ; runLength > 0; runLength-- {
					scanline[index+peek] = buf[1]
					peek++
				}
			} else {
				// a non-run
				scanline[index+peek] = buf[1]
				peek++

				nonrunLength := int(buf[0]) - 1
				if nonrunLength > 0 {
					if _, err = io.ReadFull(d.r, scanline[index+peek:index+peek+nonrunLength]); err != nil {
						if err == io.EOF {
							err = nil
						}
						return
					}

					peek += nonrunLength
				}
			}
		}

		if peek != d.config.Width {
			err = FormatError("difference in size while reading RLE scanline")
			return
		}
	}

	return
}

//--------------------------------------//
// Reader                               //
//--------------------------------------//

// DecodeConfig returns the color model and dimensions of a RGBE image without
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
	case mRGBE:
		img = hdr.NewRGB(imgRect)
	default:
		err = UnsupportedError("image mode")
		return
	}

	scanline := make([]byte, d.config.Width*4) // 4 bytes for one pixel
	pixel := make([]byte, 4)                   // RGBE pixel

	for y := 0; y < d.config.Height; y++ {

		// Read rle header
		if _, err = io.ReadFull(d.r, pixel); err != nil {
			return
		}

		// fmt.Printf("%b != 2 || %b != 2 || %d != %d - %b != 0\n", pixel[0], pixel[1], int(pixel[2])<<8|int(pixel[3]), d.config.Width, pixel[2]&0x80)
		if pixel[0] != 2 || pixel[1] != 2 || int(pixel[2])<<8|int(pixel[3]) != d.config.Width {
			// --- simple scanline (not rle)

			var n int
			n, err = io.ReadFull(d.r, scanline[4:])
			if err != nil {
				return
			}

			if n != (4*d.config.Width - 4) {
				err = FormatError("not enough data to read in the simple format")
				return
			}

			// Restore first read pixel
			scanline[0], scanline[1], scanline[2], scanline[3] = pixel[0], pixel[1], pixel[2], pixel[3]

			d.decode(img, y, scanline)
		} else {
			// --- rle scanline

			if err = d.readRLE(scanline); err != nil {
				return
			}

			d.decodeRLE(img, y, scanline)
		}
	}

	return
}

func init() {
	image.RegisterFormat("hdr", header0, Decode, DecodeConfig)
	image.RegisterFormat("hdr", header1, Decode, DecodeConfig)
	image.RegisterFormat("hdr", header2, Decode, DecodeConfig)
}
