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
	switch h.Compression {
	case CompressionGzip:
		c = gzip.NewWriter(w)
	}

	return
}

func newCompresserReader(r io.Reader, h *Header) (c io.ReadCloser) {
	var err error
	switch h.Compression {
	case CompressionGzip:
		c, err = gzip.NewReader(r)
	}
	check(err)

	return
}
