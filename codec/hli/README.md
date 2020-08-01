# HLI

The HLI is an HDR image format aimed to be simple as possible.

File extension `.hli`.


## General structure

`[magic number][1-byte header-size length][variable-length header-size][variable-length header][raster]`

### Magic number

The magic number of the the HLI is `HLi.v1` at the first line.

### Header

CBOR encoded data:
```go
type Header struct {
	Width       int               `cbor:"width"`
	Height      int               `cbor:"height"`
	Depth       int               `cbor:"depth"`
	Format      string            `cbor:"format"`
	RasterMode  string            `cbor:"raster_mode"`
	Compression string            `cbor:"compression"`
	Metadata    map[string]string `cbor:"metadata,omitempty"`
}
```
> See consts.go for possible values for each field.

- `metadata` is optional

Default options:
```go
Mode6 = &hli.Header{
    Depth:       32,
    Format:      hli.FormatLogLuv,
    RasterMode:  hli.RasterModeSeparately,
    Compression: hli.CompressionZstd,
}
```

### Raster

The raster contains all the pixels of the image.

#### Compression

The raster is compressed with the specified algorithm in the header (typically: `gzip`)

Supported compression:
- `gzip`
- `zstd` (default)

#### Format

- `RGBE` and `XYZE`

These formats are based on the Radiance RGBE encoding.
A pixel is stored on a 4-byte representation where three 8-bit mantissas shared a common 8-bit exponent.
It offers a very compact storage of 32-bit floating points.
The net result is a format that has an absolute accuracy of about 1%, covering a range of over 76 orders of magnitude.

- `RGB` and `XYZ`

These formats are based on the representation of a 32-bit floating points in bytes.
A pixel is stored on a 12-byte representation where a channel is coded on 4 bytes in little endian order.
It offers a great absolute accuracy.

- `LogLuv` (used as default format)

This format is based on the LogLuv Encoding for Full Gamut.
A pixel is stored on a 4-byte representation where a channel is coded on 4 bytes in.


```
# Original LogLuv representation

 1       15           8        8
|-+---------------|--------+--------|
 S       Le           ue       ve


# HLI LogLuv representation

    8        8        8        8
|--------+---------|--------+--------|
   SLe       le       ue       ve
```

It offers a great compression and with an absolute accuracy of about 0.3%, covering a range of over 38 orders of magnitude.


#### Raster mode

Uncompressed `Raster` is stored in several modes:

- `normal` mode

Each pixels' bytes are stored in contiguous order.

RGBE example:
```
rgbergbergbe
rgbergbergbe
```
> This is the same for XYZE.

RGB example:
```
rrrrggggbbbbrrrrggggbbbbrrrrggggbbbb
rrrrggggbbbbrrrrggggbbbbrrrrggggbbbb
```
> This is the same for XYZ.

LogLuv example:
```
SLeleueveSLeleueveSLeleueve
SLeleueveSLeleueveSLeleueve
```

- `separately` mode

The color channels are stored separately in order to improve the compression ratio.

RGBE example:
```
rrrgggbbbeee
rrrgggbbbeee
```
> This is the same for XYZE.

RGB example:
```
rrrrrrrrrrrrggggggggggggbbbbbbbbbbbb
rrrrrrrrrrrrggggggggggggbbbbbbbbbbbb
```
> This is the same for XYZ.

LogLuv example:
```
SLeSLeSLeleleleueueueveve
SLeSLeSLeleleleueueueveve
```
