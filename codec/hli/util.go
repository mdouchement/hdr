package hli

import (
	"io"
)

// func csRGBToYCoCg(r, g, b float64) (y, co, cg float64) {
// 	co = r - b
// 	t := b + (co / 2)
// 	cg = g - t
// 	y = t + (cg / 2)
//
// 	return
// }
//
// func csYCoCgToRGB(y, co, cg float64) (r, g, b float64) {
// 	t := y - (cg / 2)
// 	g = cg + t
// 	b = t - (co / 2)
// 	r = co + b
//
// 	return
// }

func readN(r io.Reader, n int) ([]byte, error) {
	p := make([]byte, n)
	_, err := io.ReadFull(r, p)
	return p, err
}

// A FormatError reports that the input is not a valid CRAD image.
type FormatError string

func (e FormatError) Error() string {
	return "crad: invalid format: " + string(e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return "crad: unsupported feature: " + string(e)
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return "crad: internal error: " + string(e)
}

//

// converts the given l to its BigEndian representation.
func lengthToBytes(l int) []byte {
	ul := uint32(l)

	if ul <= 0xFF { // uint8
		return []byte{byte(ul)}
	}

	if ul <= 0xFFFF { // uint16
		return []byte{
			byte(ul >> 8),
			byte(ul),
		}
	}

	if l <= 0xFFFFFF { // uint24
		return []byte{
			byte(ul >> 16),
			byte(ul >> 8),
			byte(ul),
		}
	}

	// 0xFFFFFFFF => uint32
	return []byte{
		byte(ul >> 24),
		byte(ul >> 16),
		byte(ul >> 8),
		byte(ul),
	}
}

// converts the given BigEndian representation to its integer value.
func bytesToLength(b []byte) int {
	if len(b) == 0 {
		panic("bad size")
	}

	if len(b) == 1 { // uint8
		return int(b[0])
	}

	if len(b) == 2 { // uint16
		return int(b[1]) | int(b[0])<<8
	}

	if len(b) == 3 { // uint24
		return int(b[2]) | int(b[1])<<8 | int(b[0])<<16
	}

	// uint32
	return int(b[3]) | int(b[2])<<8 | int(b[1])<<16 | int(b[0])<<24
}
