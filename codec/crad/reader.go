package crad

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"image"
	"io"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/format"
	"github.com/mdouchement/hdr/hdrcolor"
)

type decoder struct {
	r           io.Reader
	cr          io.ReadCloser
	h           *Header
	config      image.Config
	convert     func(pixel []byte) (float64, float64, float64)
	nbOfchannel int
	channelSize int
}

func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r: bufio.NewReader(r),
		h: new(Header),
	}

	return d, d.parseHeader()
}

//--------------------------------------//
// Header parser                        //
//--------------------------------------//

func (d *decoder) parseHeader() error {
	magic, err := readUntil(d.r, '\n')
	if err != nil {
		return err
	}
	if magic != header {
		return FormatError("format not compatible")
	}

	h, err := readUntil(d.r, '\n')
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(h), d.h); err != nil {
		return err
	}
	d.cr = newCompresserReader(d.r, d.h)

	switch d.h.Format {
	case FormatRGBE:
		d.config.ColorModel = hdrcolor.RGBModel
		d.channelSize = 1
		d.nbOfchannel = 4
		d.convert = func(p []byte) (float64, float64, float64) {
			return format.FromRadianceBytes(p[0], p[1], p[2], p[3], 1)
		}
	case FormatXYZE:
		d.config.ColorModel = hdrcolor.XYZModel
		d.channelSize = 1
		d.nbOfchannel = 4
		d.convert = func(p []byte) (float64, float64, float64) {
			return format.FromRadianceBytes(p[0], p[1], p[2], p[3], 1)
		}
	case FormatRGB:
		d.channelSize = 4
		d.nbOfchannel = 3
		d.config.ColorModel = hdrcolor.RGBModel
		d.convert = func(p []byte) (float64, float64, float64) {
			return format.FromBytes(binary.LittleEndian, p)
		}
	case FormatXYZ:
		d.config.ColorModel = hdrcolor.XYZModel
		d.channelSize = 4
		d.nbOfchannel = 3
		d.convert = func(p []byte) (float64, float64, float64) {
			return format.FromBytes(binary.LittleEndian, p)
		}
	case FormatLogLuv:
		d.config.ColorModel = hdrcolor.XYZModel
		d.channelSize = 1
		d.nbOfchannel = 4
		d.convert = func(p []byte) (float64, float64, float64) {
			return format.LogLuvToXYZ(p[0], p[1], p[2], p[3])
		}
	}
	d.config.Width = d.h.Width
	d.config.Height = d.h.Height

	return nil
}

//--------------------------------------//
// Pixels parser                        //
//--------------------------------------//

func (d *decoder) decode(dst image.Image, y int, scanline []byte) {
	size := d.nbOfchannel * d.channelSize

	for x := 0; x < d.config.Width; x++ {
		b0, b1, b2 := d.convert(scanline[x*size : x*size+size])

		switch d.h.Format {
		case FormatRGBE:
			fallthrough
		case FormatRGB:
			img := dst.(*hdr.RGB)
			img.SetRGB(x, y, hdrcolor.RGB{R: b0, G: b1, B: b2})
		case FormatXYZE:
			fallthrough
		case FormatLogLuv:
			fallthrough
		case FormatXYZ:
			img := dst.(*hdr.XYZ)
			img.SetXYZ(x, y, hdrcolor.XYZ{X: b0, Y: b1, Z: b2})
		}
	}
}

func (d *decoder) decodeSeparately(dst image.Image, y int, scanline []byte) {
	for x := 0; x < d.config.Width; x++ {
		pixel := make([]byte, 0, d.nbOfchannel*d.channelSize)
		for c := 0; c < d.nbOfchannel; c++ {
			pos := x*d.channelSize + c*d.channelSize*d.config.Width
			pixel = append(pixel, scanline[pos:pos+d.channelSize]...)
		}

		b0, b1, b2 := d.convert(pixel)

		switch d.h.Format {
		case FormatRGBE:
			fallthrough
		case FormatRGB:
			img := dst.(*hdr.RGB)
			img.SetRGB(x, y, hdrcolor.RGB{R: b0, G: b1, B: b2})
		case FormatXYZE:
			fallthrough
		case FormatLogLuv:
			fallthrough
		case FormatXYZ:
			img := dst.(*hdr.XYZ)
			img.SetXYZ(x, y, hdrcolor.XYZ{X: b0, Y: b1, Z: b2})
		}
	}
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
	switch d.h.Format {
	case FormatRGBE:
		fallthrough
	case FormatRGB:
		img = hdr.NewRGB(imgRect)
	case FormatXYZE:
		fallthrough
	case FormatLogLuv:
		fallthrough
	case FormatXYZ:
		img = hdr.NewXYZ(imgRect)
	default:
		err = UnsupportedError("image mode")
		return
	}

	scanline := make([]byte, d.config.Width*d.nbOfchannel*d.channelSize)

	for y := 0; y < d.config.Height; y++ {
		_, err = io.ReadFull(d.cr, scanline)
		if err != nil {
			return
		}

		switch d.h.RasterMode {
		case RasterModeNormal:
			d.decode(img, y, scanline)
		case RasterModeSeparately:
			d.decodeSeparately(img, y, scanline)
		}
	}

	return
}

func init() {
	image.RegisterFormat("crad", header, Decode, DecodeConfig)
}
