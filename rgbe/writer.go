package rgbe

import (
	"bufio"
	"fmt"
	"io"

	"github.com/mdouchement/hdr"
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

	switch e.mode {
	case mRGBE:
		_, err = io.WriteString(e.w, "FORMAT=32-bit_rle_rgbe\n")
	case mXYZE:
		_, err = io.WriteString(e.w, "FORMAT=32-bit_rle_xyze\n")
	}
	if err != nil {
		return err
	}

	d := e.m.Bounds().Size()
	size := fmt.Sprintf("\n-Y %d +X %d\n", d.Y, d.X)
	_, err = io.WriteString(e.w, size)

	return err
}

//--------------------------------------//
// Pixels writer                        //
//--------------------------------------//

func (e *encoder) encode() error {
	w := bufio.NewWriter(e.w)
	ar := newAR(e.m)

	d := e.m.Bounds().Size()

	var err error
	for y := 0; y < d.Y; y++ {
		for x := 0; x < d.X; x++ {
			_, err = w.Write(floatsToBytes(ar.at(x, y)))

			if err != nil {
				return err
			}
		}
	}

	return w.Flush()
}

// Encode writes the Image m to w in RGBE format.
func Encode(w io.Writer, m hdr.Image) error {
	e := newEncoder(w, m)

	switch m.(type) {
	case *hdr.RGB:
		e.mode = mRGBE
	case *hdr.XYZ:
		e.mode = mXYZE
	default:
		return UnsupportedError("color space")
	}

	if err := e.writeHeader(); err != nil {
		return err
	}

	return e.encode()
}
