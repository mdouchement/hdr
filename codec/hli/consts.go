package hli

const (
	header = "HLi.v1"
)

const (
	// FormatRGBE for RGBE model
	FormatRGBE = "RGBE"
	// FormatXYZE for XYZE model
	FormatXYZE = "XYZE"
	// FormatRGB for RGB model
	FormatRGB = "RGB"
	// FormatXYZ for XYZ model
	FormatXYZ = "XYZ"
	// FormatLogLuv for LogLuv model
	FormatLogLuv = "LogLuv"

	// ColorModelYCoCgRE = "YCoCg-RE"
	// ColorModelYCoCgR  = "YCoCg-R"

	// RasterModeNormal for normal pixel positioning
	RasterModeNormal = "normal"
	// RasterModeSeparately for separately pixel's color positioning
	RasterModeSeparately = "separately"

	// CompressionGzip for gzip compression
	CompressionGzip = "gzip"
	// CompressionZstd for Zstandard compression
	CompressionZstd = "zstd"
)

// A Header handles all image properties.
type Header struct {
	Width       int               `cbor:"width"`
	Height      int               `cbor:"height"`
	Depth       int               `cbor:"depth"`
	Format      string            `cbor:"format"`
	RasterMode  string            `cbor:"raster_mode"`
	Compression string            `cbor:"compression"`
	Metadata    map[string]string `cbor:"metadata,omitempty"`
}

var (
	// Mode1 offers the better compression in RGBE/XYZE color model depending to
	// the provided hdr.Image implementation. (quantization steps: 1%)
	Mode1 = &Header{
		Depth:       32,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionGzip,
	}
	// Mode2 offers the better compression in XYZE that covers gamut. (quantization steps: 1%)
	Mode2 = &Header{
		Depth:       32,
		Format:      FormatXYZE,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionGzip,
	}
	// Mode3 offers a trade off in compression/quality in RGB. (quantization steps: 0.1%)
	Mode3 = &Header{
		Depth:       32,
		Format:      FormatRGB,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionGzip,
	}
	// Mode4 offers the better quality in XYZ that covers gamut. (quantization steps: 0.1%)
	Mode4 = &Header{
		Depth:       32,
		Format:      FormatXYZ,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionGzip,
	}
	// Mode5 offers the better compression and quality in LogLuv that covers gamut. (quantization steps: 0.1%)
	Mode5 = &Header{
		Depth:       32,
		Format:      FormatLogLuv,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionGzip,
	}
	// Mode6 offers the better and faster compression and quality in LogLuv that covers gamut. (quantization steps: 0.1%)
	Mode6 = &Header{
		Depth:       32,
		Format:      FormatLogLuv,
		RasterMode:  RasterModeSeparately,
		Compression: CompressionZstd,
	}
)
