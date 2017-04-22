// Resources:
// http://wwwimages.adobe.com/content/dam/Adobe/en/products/photoshop/pdfs/dng_spec_1.4.0.0.pdf
// https://github.com/golang/image/tree/master/tiff
//
// https://github.com/google/tiff
// https://github.com/google/tiff/blob/master/dng/dng.go
// https://github.com/google/cameraraw/blob/master/devices/canon/ifd.go (extra IDF)
//
// https://github.com/chai2010/tiff (float)

// http://www.awaresystems.be/imaging/tiff.html

package dng

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/mdouchement/hdr/hdrcolor"
)

type decoder struct {
	r         io.ReaderAt
	byteOrder binary.ByteOrder
	config    image.Config
	mode      imageMode
	bpp       uint
	features  map[int][]uint
	palette   []color.Color

	buf   []byte
	off   int    // Current offset in buf.
	v     uint32 // Buffer value for reading with arbitrary bit depths.
	nbits uint   // Remaining number of bits in v.
}

// readBits reads n bits from the internal buffer starting at the current offset.
func (d *decoder) readBits(n uint) uint32 {
	for d.nbits < n {
		d.v <<= 8
		d.v |= uint32(d.buf[d.off])
		d.off++
		d.nbits += 8
	}
	d.nbits -= n
	rv := d.v >> d.nbits
	d.v &^= rv << d.nbits
	return rv
}

// firstVal returns the first uint of the features entry with the given tag,
// or 0 if the tag does not exist.
func (d *decoder) firstVal(tag int) uint {
	f := d.features[tag]
	if len(f) == 0 {
		return 0
	}
	return f[0]
}

// flushBits discards the unread bits in the buffer used by readBits.
// It is used at the end of a line.
func (d *decoder) flushBits() {
	d.v = 0
	d.nbits = 0
}

//------------------------//
// Header parser          //
//------------------------//

// parseIFD decides whether the the IFD entry in p is "interesting" and
// stows away the data in the decoder.
func (d *decoder) parseIFD(p []byte) error {
	tag := d.byteOrder.Uint16(p[0:2])
	switch tag {
	case tBitsPerSample,
		tExtraSamples,
		tPhotometricInterpretation,
		tCompression,
		tPredictor,
		tStripOffsets,
		tStripByteCounts,
		tRowsPerStrip,
		tTileWidth,
		tTileLength,
		tTileOffsets,
		tTileByteCounts,
		tImageLength,
		tImageWidth,
		tDNGVersion,
		tDNGBackwardVersion:
		val, err := d.ifdUint(p)
		if err != nil {
			return err
		}
		d.features[int(tag)] = val
	case tColorMap:
		// FIXME could be dropped or updated to DNG.
		val, err := d.ifdUint(p)
		if err != nil {
			return err
		}
		numcolors := len(val) / 3
		if len(val)%3 != 0 || numcolors <= 0 || numcolors > 256 {
			return FormatError("bad ColorMap length")
		}
		d.palette = make([]color.Color, numcolors)
		for i := 0; i < numcolors; i++ {
			d.palette[i] = color.RGBA64{
				uint16(val[i]),
				uint16(val[i+numcolors]),
				uint16(val[i+2*numcolors]),
				0xffff,
			}
		}
	case tSampleFormat:
		// Page 27 of the spec: If the SampleFormat is present and
		// the value is not 1 [= unsigned integer data], a Baseline
		// TIFF reader that cannot handle the SampleFormat value
		// must terminate the import process gracefully.
		val, err := d.ifdUint(p)
		if err != nil {
			return err
		}
		fmt.Println("SampleFormat:", val) // tSampleFormat == 3 only when bpp == 32
		for _, v := range val {
			if v != 1 {
				return UnsupportedError("sample format")
			}
		}
	default:
		// val, err := d.ifdUint(p)
		// if err == nil {
		// 	fmt.Println(tag, "-", val)
		// }
		fmt.Println(tag, "-", p)
	}
	return nil
}

// ifdUint decodes the IFD entry in p, which must be of the Byte, Short
// or Long type, and returns the decoded uint values.
func (d *decoder) ifdUint(p []byte) (u []uint, err error) {
	var raw []byte
	datatype := d.byteOrder.Uint16(p[2:4])
	count := d.byteOrder.Uint32(p[4:8])
	if datalen := lengths[datatype] * count; datalen > 4 {
		// The IFD contains a pointer to the real value.
		raw = make([]byte, datalen)
		_, err = d.r.ReadAt(raw, int64(d.byteOrder.Uint32(p[8:12])))
	} else {
		raw = p[8 : 8+datalen]
	}
	if err != nil {
		return nil, err
	}

	u = make([]uint, count)
	switch datatype {
	case dtByte:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(raw[i])
		}
	case dtShort:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(d.byteOrder.Uint16(raw[2*i : 2*(i+1)]))
		}
	case dtLong:
		for i := uint32(0); i < count; i++ {
			u[i] = uint(d.byteOrder.Uint32(raw[4*i : 4*(i+1)]))
		}
	default:
		return nil, UnsupportedError("data type")
	}
	return u, nil
}

//------------------------//
// Pixels parser          //
//------------------------//

//------------------------//
// Reader                 //
//------------------------//

func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r:        newReaderAt(r),
		features: make(map[int][]uint),
	}

	p := make([]byte, 8)
	if _, err := d.r.ReadAt(p, 0); err != nil {
		return nil, err
	}
	switch string(p[0:4]) {
	case leHeader:
		d.byteOrder = binary.LittleEndian
	case beHeader:
		d.byteOrder = binary.BigEndian
	default:
		return nil, FormatError("malformed header")
	}

	ifdOffset := int64(d.byteOrder.Uint32(p[4:8]))

	// The first two bytes contain the number of entries (12 bytes each).
	if _, err := d.r.ReadAt(p[0:2], ifdOffset); err != nil {
		return nil, err
	}
	numItems := int(d.byteOrder.Uint16(p[0:2]))

	// All IFD entries are read in one chunk.
	p = make([]byte, ifdLen*numItems)
	if _, err := d.r.ReadAt(p, ifdOffset+2); err != nil {
		return nil, err
	}

	for i := 0; i < len(p); i += ifdLen {
		if err := d.parseIFD(p[i : i+ifdLen]); err != nil {
			return nil, err
		}
	}

	fmt.Println("HERE")
	for idf, v := range d.features {
		fmt.Printf("%s: %d\n", tag(idf), v)
	}

	d.config.Width = int(d.firstVal(tImageWidth))
	d.config.Height = int(d.firstVal(tImageLength))

	if _, ok := d.features[tBitsPerSample]; !ok {
		return nil, FormatError("BitsPerSample tag missing")
	}
	d.bpp = d.firstVal(tBitsPerSample)
	fmt.Println("BPP:", d.bpp)

	// Determine the image mode.
	switch d.firstVal(tPhotometricInterpretation) {
	case pWhiteIsZero:
		fallthrough
	case pBlackIsZero:
		fallthrough
	case pRGB:
		fallthrough
	case pPaletted:
		fallthrough
	case pTransMask:
		// All LDR modes are droped.
		return nil, UnsupportedError("color model, use Golang's lib for LDR images")
	case pCMYK:
		d.mode = m
		d.config.ColorModel = hdrcolor.RGB
	default:
		return nil, UnsupportedError("color model")
	}

	return d, nil
}

// DecodeConfig returns the color model and dimensions of a DNG image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	d, err := newDecoder(r)
	if err != nil {
		return image.Config{}, err
	}
	return d.config, nil
}

// Decode reads a DNG image from r and returns an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	d, err := newDecoder(r)
	if err != nil {
		return nil, err
	}

	for idf, v := range d.features {
		fmt.Printf("%s: %d", tag(idf), v)
	}

	return nil, nil
}

func init() {
	image.RegisterFormat("dng", leHeader, Decode, DecodeConfig)
	image.RegisterFormat("dng", beHeader, Decode, DecodeConfig)
}
