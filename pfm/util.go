package pfm

import (
	"bytes"
	"io"
)

const (
	header0 = "PF" // Color
	header1 = "pf" // Grayscale
)

// byte order of each channel
type endianness int

const (
	eLittleEndian endianness = iota
	eBigEndian
)

// imageMode represents the mode of the image.
type imageMode int

const (
	mColor imageMode = iota
	mGrayscale
)

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

// A FormatError reports that the input is not a valid PFM image.
type FormatError string

func (e FormatError) Error() string {
	return "pfm: invalid format: " + string(e)
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return "pfm: unsupported feature: " + string(e)
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return "pfm: internal error: " + string(e)
}
