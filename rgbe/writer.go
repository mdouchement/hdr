package rgbe

import (
	"bufio"
	"fmt"
	"io"

	"github.com/mdouchement/hdr"
)

// RLEWrites allows to write image file with run-length encoding.
var RLEWrites = true

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

func (e *encoder) encodeRLE() error {
	w := bufio.NewWriter(e.w)
	ar := newAR(e.m)
	d := e.m.Bounds().Size()

	// RLE header
	header := make([]byte, 4)
	header[0] = 2
	header[1] = 2
	header[2] = byte(d.X >> 8)
	header[3] = byte(d.X & 0xFF)

	scanline := make([]byte, d.X*4)

	var err error
	for y := 0; y < d.Y; y++ {
		// Prepare RLE treatment for each channel.
		for x := 0; x < d.X; x++ {
			pixel := floatsToBytes(ar.at(x, y))
			scanline[x] = pixel[0]       // R or X
			scanline[x+d.X] = pixel[1]   // G or Y
			scanline[x+2*d.X] = pixel[2] // B or Z
			scanline[x+3*d.X] = pixel[3] // Exposure
		}

		// Append data to the file
		_, err = w.Write(header)
		if err != nil {
			return err
		}

		for c := 0; c < 4; c++ {
			// Apply RLE for each channel
			offset := c * d.X
			err = e.writeRLE(w, scanline[offset:offset+d.X])
			if err != nil {
				return err
			}
		}
	}

	return w.Flush()
}

func (e *encoder) writeRLE(w *bufio.Writer, scanline []byte) error {
	l := 0
	eor := len(scanline) // End of RLE

	// One run matches to n successive same byte
	for l < eor {
		offset := 0 // Start offset of the n successive same byte
		index := 0  // Bytes iterator
		n := 0      // The run length

		for n <= 4 && index < 128 && l+index < eor {
			offset = index
			n = 0

			// Count successive bytes
			for n < 127 && offset+n < 128 && l+index < eor && scanline[l+offset] == scanline[l+index] {
				index++
				n++
			}
		}
		// fmt.Printf("l: %d, o: %d, n: %d, p: %d\n", l, offset, n, peek)

		// For an efficent RLE, we need more than 4 bytes.
		// Under 5 bytes, it is more efficient to write the bytes.
		if n > 4 {
			// Write a short-run (all the read bytes before the found run)
			if offset > 0 {
				// Write short-run size
				if err := w.WriteByte(byte(offset)); err != nil {
					return err
				}
				// Write short-run
				if _, err := w.Write(scanline[l : l+offset]); err != nil {
					return err
				}
			}

			// Write a run
			// fmt.Printf("l: %d, o: %d, n: %d, p: %d - %d\n", l, offset, n, index, scanline[offset])
			if _, err := w.Write([]byte{byte(128 + n), scanline[l+offset]}); err != nil {
				return err
			}
		} else {
			// Write a non run with size of index (128 or the remain bytes of the end of scanline)
			//
			// Write non-run size
			if err := w.WriteByte(byte(index)); err != nil {
				return err
			}
			// Write non-run
			if _, err := w.Write(scanline[l : l+index]); err != nil {
				return err
			}
		}

		l += index
	}

	if l != eor {
		return FormatError("difference in size while writing RLE scanline")
	}

	return nil
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

	if RLEWrites {
		return e.encodeRLE()
	}

	return e.encode()
}
