package crad

import (
	"bytes"
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

func readUntil(r io.Reader, delimiter byte) (string, error) {
	buf := &bytes.Buffer{}
	p := make([]byte, 1)

	for {
		if _, err := r.Read(p); err != nil {
			return "", err
		}

		if p[0] != delimiter {
			buf.Write(p)
		} else {
			return buf.String(), nil
		}
	}
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
