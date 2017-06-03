package format

import "math"

// XYZToLogLuv converts XYZ floats 64 bits to LogLuv bytes.
//
//	 1       15           8        8
//	|-+---------------|--------+--------|
//	 S     Le+le          ue       ve
func XYZToLogLuv(x, y, z float64) []byte {
	l := YToSLe(y)

	absY := math.Abs(y)
	s := x + 15*absY + 3*z
	ubis := 4 * x / s
	vbis := 9 * absY / s

	SLe, le := Uint16ToBytes(l)

	var ue byte
	if ubis > 0 {
		ue = byte(math.Floor(410 * ubis))
	}

	var ve byte
	if vbis > 0 {
		ve = byte(math.Floor(410 * vbis))
	}

	return []byte{SLe, le, ue, ve}
}

// LogLuvToXYZ converts LogLuv bytes to XYZ floats 64 bits.
//
//	 1       15           8        8
//	|-+---------------|--------+--------|
//	 S     Le+le          ue       ve
func LogLuvToXYZ(SLe, le, ue, ve byte) (x, y, z float64) {
	l := BytesToUint16(SLe, le)
	y = SLeToY(l)

	ubis := (float64(ue) + 0.5) / 410
	vbis := (float64(ve) + 0.5) / 410

	s := 6*ubis - 16*vbis + 12
	lx := 9 * ubis / s
	ly := 4 * vbis / s

	absY := math.Abs(y)
	x = lx / ly * absY
	z = (1 - lx - ly) / ly * absY

	return
}

// YToSLe converts Y to Le with S.
//
//	 1       15
//	|-+---------------|
//	 S       Le
func YToSLe(y float64) uint16 {
	le := uint16(math.Floor(256 * (math.Log2(math.Abs(y)) + 64)))

	if y < 0.0 {
		return 0x8000 | le // Add S for negative Y
	}
	return le
}

// SLeToY converts Le with S to Y.
//
//	 1       15
//	|-+---------------|
//	 S       Le
func SLeToY(sle uint16) float64 {
	le := sle & 0x7FFF // Drop S from le
	y := math.Exp2((float64(le)+0.5)/256 - 64)

	if 0x8000&sle != 0 { // Inspect S
		return -y
	}
	return y
}

// Uint16ToBytes splits v into 2 bytes (big endian order).
func Uint16ToBytes(v uint16) (byte, byte) {
	return byte(v >> 8), byte(v & 0xff)
}

// BytesToUint16 merges a & b (big endian order) into an uint16 value.
func BytesToUint16(a, b byte) uint16 {
	return uint16(a)<<8 | uint16(b)
}
