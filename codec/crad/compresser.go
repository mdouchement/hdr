package crad

import (
	"compress/gzip"
	"io"
)

type compresserWriter interface {
	io.WriteCloser
	Flush() error
}

func newCompresserWriter(w io.Writer, h *Header) (c compresserWriter) {
	if h.Compression == CompressionGzip {
		c = gzip.NewWriter(w)
	}

	return
}

func newCompresserReader(r io.Reader, h *Header) (c io.ReadCloser) {
	var err error
	if h.Compression == CompressionGzip {
		c, err = gzip.NewReader(r)
	}
	if err != nil {
		panic(err) // Should never occurred
	}

	return
}
