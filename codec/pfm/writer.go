package pfm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/format"
	"github.com/mdouchement/hdr/hdrcolor"
)

type encoder struct {
	w    io.Writer
	m    hdr.Image
	mode imageMode
}

func newEncoder(w io.Writer, m hdr.Image) *encoder {
	return &encoder{
		w: w,
		m: m,
	}
}

//--------------------------------------//
// Header writer                        //
//--------------------------------------//

func (e *encoder) writeHeader() error {
	_, err := io.WriteString(e.w, header0+"\n")
	if err != nil {
		return err
	}

	d := e.m.Bounds().Size()
	size := fmt.Sprintf("%d %d\n", d.X, d.Y)
	_, err = io.WriteString(e.w, size)
	if err != nil {
		return err
	}

	_, err = io.WriteString(e.w, "-1.0\n")
	return err
}

// Encode writes the Image m to w in PFM format.
func Encode(w io.Writer, m hdr.Image) error {
	e := newEncoder(w, m)

	switch m.ColorModel() {
	case hdrcolor.RGBModel:
		e.mode = mColor
	default:
		return UnsupportedError("color space")
	}

	if err := e.writeHeader(); err != nil {
		return err
	}

	buff := bufio.NewWriter(w)

	// The pixels in each row ordered left to right and the rows ordered bottom to top
	for y := m.Bounds().Dy() - 1; y >= 0; y-- {
		for x := 0; x < m.Bounds().Dx(); x++ {
			r, g, b, _ := m.HDRAt(x, y).HDRRGBA()

			if _, err := buff.Write(format.ToBytes(binary.LittleEndian, r, g, b)); err != nil {
				return err
			}
		}
	}

	return buff.Flush()
}
