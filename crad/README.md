# CRAD - Compressed Radiance RGBE/XYZE

File extension `.hdr2`


## General structure

```
#?COMPRESSED_RADIANCE
{"width":800,"height":600,"color_space":"RGBE","compression":"gzip"}
DATA
```

## Data structure

`DATA` is compressed with the specified algorithm in the JSON header (typically: `gzip`).

Uncompressed `DATA` are slices of width length, each slice stores channels separately.

For example, an uncompressed of RGBE 3x3p:
```
R{0,0}R{1,0}R{2,0}B{0,0}B{1,0}B{2,0}G{0,0}G{1,0}G{2,0}E{0,0}E{1,0}E{2,0}
R{0,1}R{1,1}R{2,1}B{0,1}B{1,1}B{2,1}G{0,1}G{1,1}G{2,1}E{0,1}E{1,1}E{2,1}
R{0,2}R{1,2}R{2,2}B{0,2}B{1,2}B{2,2}G{0,2}G{1,2}G{2,2}E{0,2}E{1,2}E{2,2}
```
> `channel{x,y}`

This is the same for XYZE.
