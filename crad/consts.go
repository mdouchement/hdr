package crad

const (
	header = "#?CRAD"
)

const (
	// ColorModelRGBE for RGBE model
	ColorModelRGBE = "RGBE"
	// ColorModelXYZE for XYZE model
	ColorModelXYZE = "XYZE"
	// ColorModelRGB for RGB model
	ColorModelRGB = "RGB"
	// ColorModelXYZ for XYZ model
	ColorModelXYZ = "XYZ"

	// ColorModelYCoCgRE = "YCoCg-RE"
	// ColorModelYCoCgR  = "YCoCg-R"

	// RasterModeNormal for normal pixel positioning
	RasterModeNormal = "normal"
	// RasterModeSeparately for separately pixel' color positioning
	RasterModeSeparately = "separately"

	// CompressionGzip for gzip compression
	CompressionGzip = "gzip"
)

// A Header handles all image properties.
type Header struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Depth       int    `json:"depth"`
	ColorModel  string `json:"color_model"`
	RasterMode  string `json:"raster_mode"`
	Compression string `json:"compression"`
}
