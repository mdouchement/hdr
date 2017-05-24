# CRAD

The CRAD is an HDR image format aimed to be simple as possible.

File extension `.crad`.


## General structure

```
#?CRAD\n
{"width":3272,"height":1280,"depth":32,"format":"RGBE","raster_mode":"normal","compression":"gzip"}\n
+----------------+
|                |
|     Raster     |
|                |
+----------------+
```

### Magic number

The magic number of the the CRAD is `#?CRAD` at the first line.

### Header

The header is a one-line JSON string at the second line.

### Raster

The raster contains all the pixels of the image.

#### Compression

The raster is compressed with the specified algorithm in the JSON header (typically: `gzip`)

#### Format

- RGBE and XYZE

These formats are based on the Radiance RGBE enconding.
A pixel is stored on a 4-byte representation where three 8-bit mantissas shared a common 8-bit exponent.
It offers a very compact storage of 32-bit floating points.
The net result is a format that has an absolute accuracy of about 1%, covering a range of over 76 orders of magnitude.

- RGB and XYZ

These formats are based on the representation of a 32-bit floating points in bytes.
A pixel is stored on a 12-byte representation where a channel is coded on 4 bytes in little endian order.
It offers a great absolute accuracy.


#### Raster mode

Uncompressed `Raster` is stored in several modes:

- Normal mode

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

- Separately mode

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
rrrrrrrrrrrrggggggggggggbbbbbbbbbbbb
```
> This is the same for XYZ.
