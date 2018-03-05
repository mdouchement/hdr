package hdr

import "github.com/mdouchement/hdr/hdrcolor"

//===============//
// LMS           //
//===============//

// A LMSCAT02w wrapper hollows to get pixels in LMS-space using CIE CAT02 matrix.
// https://en.wikipedia.org/wiki/CIECAM02#Chromatic_adaptation
type LMSCAT02w struct {
	Image
}

// NewLMSCAT02w instanciates a new LMSCAT02w wrapper.
func NewLMSCAT02w(m Image) *LMSCAT02w {
	return &LMSCAT02w{Image: m}
}

// HDRAt returns the pixel in LMS-space using CIE CAT02 matrix.
func (p *LMSCAT02w) HDRAt(x, y int) hdrcolor.Color {
	X, Y, Z, _ := p.Image.HDRAt(x, y).HDRXYZA()
	L, M, S := hdrcolor.XyzToLmsMcat02(X, Y, Z)
	return hdrcolor.RAW{P1: L, P2: M, P3: S}
}

// A LMSHPEw wrapper hollows to get pixels in LMS-space using Hunt-Pointer-Estevez matrix.
// https://en.wikipedia.org/wiki/LMS_color_space
type LMSHPEw struct {
	Image
}

// NewLMSHPEw instanciates a new LMSHPE wrapper.
func NewLMSHPEw(m Image) *LMSHPEw {
	return &LMSHPEw{Image: m}
}

// HDRAt returns the pixel in LMS-space using Hunt-Pointer-Estevez matrix.
func (p *LMSHPEw) HDRAt(x, y int) hdrcolor.Color {
	X, Y, Z, _ := p.Image.HDRAt(x, y).HDRXYZA()
	L, M, S := hdrcolor.XyzToLmsMhpe(X, Y, Z)
	return hdrcolor.RAW{P1: L, P2: M, P3: S}
}
