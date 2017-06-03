package crad

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/format"
)

type encoder struct {
	w           io.Writer
	m           hdr.Image
	h           *Header
	bytesAt     func(x, y int) []byte
	nbOfchannel int
	channelSize int
}

func newEncoder(w io.Writer, m hdr.Image, h *Header) *encoder {
	return &encoder{
		w: w,
		m: m,
		h: h,
	}
}

//--------------------------------------//
// Header stuff                         //
//--------------------------------------//

func (e *encoder) configureHeader() error {
	// Header - Defaults
	if e.h.Depth == 0 {
		e.h.Depth = 32
	}
	if e.h.Compression == "" {
		e.h.Compression = CompressionGzip
	}
	if e.h.RasterMode == "" {
		e.h.RasterMode = RasterModeNormal
	}
	if e.h.Format == "" {
		switch e.m.(type) {
		case *hdr.RGB:
			e.h.Format = FormatRGBE
		case *hdr.XYZ:
			e.h.Format = FormatXYZE
		default:
			return UnsupportedError("color model")
		}
	}

	// Header - Format
	switch e.h.Format {
	case FormatRGBE:
		e.channelSize = 1
		e.nbOfchannel = 4
		e.bytesAt = func(x, y int) []byte {
			r, g, b, _ := e.m.HDRAt(x, y).HDRRGBA()
			return format.ToRadianceBytes(r, g, b)
		}
	case FormatXYZE:
		e.channelSize = 1
		e.nbOfchannel = 4
		e.bytesAt = func(x, y int) []byte {
			xx, yy, zz, _ := e.m.HDRAt(x, y).HDRXYZA()
			return format.ToRadianceBytes(xx, yy, zz)
		}
	case FormatRGB:
		e.channelSize = 4
		e.nbOfchannel = 3
		e.bytesAt = func(x, y int) []byte {
			r, g, b, _ := e.m.HDRAt(x, y).HDRRGBA()
			return format.ToBytes(r, g, b)
		}
	case FormatXYZ:
		e.channelSize = 4
		e.nbOfchannel = 3
		e.bytesAt = func(x, y int) []byte {
			xx, yy, zz, _ := e.m.HDRAt(x, y).HDRXYZA()
			return format.ToBytes(xx, yy, zz)
		}
	}

	// Header - Size
	d := e.m.Bounds().Size()
	e.h.Width = d.X
	e.h.Height = d.Y

	return nil
}

func (e *encoder) writeHeader() error {
	_, err := io.WriteString(e.w, header+"\n")
	if err != nil {
		return err
	}

	raw, err := json.Marshal(e.h)
	if err != nil {
		return err
	}

	_, err = e.w.Write(append(raw, '\n'))

	return err
}

//--------------------------------------//
// Pixels writer                        //
//--------------------------------------//

func (e *encoder) encode(w compresserWriter) error {
	d := e.m.Bounds().Size()

	var err error
	for y := 0; y < d.Y; y++ {
		for x := 0; x < d.X; x++ {
			pixel := e.bytesAt(x, y)
			_, err = w.Write(pixel)

			if err != nil {
				return err
			}
		}
	}

	return w.Flush()
}

func (e *encoder) encodeSeparately(w compresserWriter) error {
	d := e.m.Bounds().Size()
	writeline := make([]byte, d.X*e.nbOfchannel*e.channelSize)

	var err error
	for y := 0; y < d.Y; y++ {
		for x := 0; x < d.X; x++ {
			// Separate colors
			pixel := e.bytesAt(x, y)

			for c := 0; c < e.nbOfchannel; c++ {
				pos := x*e.channelSize + c*e.channelSize*e.h.Width
				for i := 0; i < e.channelSize; i++ {
					writeline[pos+i] = pixel[c*e.channelSize+i]
				}
			}
		}
		_, err = w.Write(writeline)

		if err != nil {
			return err
		}
	}

	return w.Flush()
}

// Encode writes the Image m to w in CRAD format.
func Encode(w io.Writer, m hdr.Image) error {
	return EncodeWithOptions(w, m, Mode1)
}

// EncodeWithOptions writes the Image m to w in CRAD format.
func EncodeWithOptions(w io.Writer, m hdr.Image, h *Header) error {
	e := newEncoder(w, m, h)

	if err := e.configureHeader(); err != nil {
		return err
	}

	if err := e.writeHeader(); err != nil {
		return err
	}

	wb := bufio.NewWriter(e.w)
	wc := newCompresserWriter(wb, e.h)
	defer wc.Close()

	// Write raster
	switch e.h.RasterMode {
	case RasterModeNormal:
		if err := e.encode(wc); err != nil {
			return err
		}
	case RasterModeSeparately:
		if err := e.encodeSeparately(wc); err != nil {
			return err
		}
	default:
		return UnsupportedError("raster mode")
	}

	return wb.Flush()
}
