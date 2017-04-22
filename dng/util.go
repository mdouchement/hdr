package dng

import "fmt"

// A FormatError reports that the input is not a valid TIFF image.
type FormatError string

func (e FormatError) Error() string {
	return fmt.Sprintf("dng: invalid format: %s", string(e))
}

// An UnsupportedError reports that the input uses a valid but
// unimplemented feature.
type UnsupportedError string

func (e UnsupportedError) Error() string {
	return fmt.Sprintf("dng: unsupported feature: %s", string(e))
}

// An InternalError reports that an internal error was encountered.
type InternalError string

func (e InternalError) Error() string {
	return fmt.Sprintf("dng: internal error: %s", string(e))
}

// minInt returns the smaller of x or y.
func minInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func tag(t int) string {
	switch t {
	case tBitsPerSample:
		return "BitsPerSample"
	case tExtraSamples:
		return "BitsPerSample"
	case tPhotometricInterpretation:
		return "PhotometricInterpretation"
	case tCompression:
		return "Compression"
	case tPredictor:
		return "Predictor"
	case tStripOffsets:
		return "StripOffsets"
	case tStripByteCounts:
		return "StripByteCounts"
	case tRowsPerStrip:
		return "RowsPerStrip"
	case tTileWidth:
		return "TileWidth"
	case tTileLength:
		return "TileLength"
	case tTileOffsets:
		return "TileOffsets"
	case tTileByteCounts:
		return "TileByteCounts"
	case tImageLength:
		return "ImageLength"
	case tImageWidth:
		return "ImageWidth"
	case tDNGVersion:
		return "DNGVersion"
	case tDNGBackwardVersion:
		return "DNGBackwardVersion"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}
