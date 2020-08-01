package hli

import (
	"compress/gzip"
	"io"

	"github.com/klauspost/compress/zstd"
)

type compresserWriter interface {
	io.WriteCloser
	Flush() error
}

func newCompresserWriter(w io.Writer, h *Header) (c compresserWriter) {
	var err error

	switch h.Compression {
	case CompressionGzip:
		c = gzip.NewWriter(w)
	case CompressionZstd:
		c, err = zstd.NewWriter(w)
	}

	if err != nil {
		panic(err) // Should never occurred
	}

	return
}

func newCompresserReader(r io.Reader, h *Header) (c io.ReadCloser) {
	var err error

	switch h.Compression {
	case CompressionGzip:
		c, err = gzip.NewReader(r)
	case CompressionZstd:
		c, err = newzstdreader(r)
	}

	if err != nil {
		panic(err) // Should never occurred
	}

	return
}

type zstdreader struct {
	*zstd.Decoder
}

func newzstdreader(r io.Reader) (*zstdreader, error) {
	c, err := zstd.NewReader(r)
	return &zstdreader{c}, err
}

func (r *zstdreader) Close() error {
	r.Decoder.Close()
	return nil
}
