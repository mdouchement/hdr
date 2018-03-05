package hdrcolor

// XyzToLms converts from CIE XYZ-space to LMS-space (using D65-LMS matrix).
func XyzToLms(x, y, z float64) (l, m, s float64) {
	l = 0.4002*x + 0.7075*y - 0.0807*z
	m = -0.228*x + 1.1500*y + 0.0612*z
	s = 0.0000*x + 0.0000*y + 0.9184*z
	return
}

// LmsToXyz converts from LMS-space (using D65-LMS matrix) to CIE XYZ-space.
func LmsToXyz(l, m, s float64) (x, y, z float64) {
	x = 1.8501*l - 1.1385*m + 0.2384*s
	y = 0.3668*l + 0.6438*m - 0.0107*s
	z = 0.0000*l + 0.0000*m + 1.0889*s
	return
}

// XyzToLmsMcat02 converts from CIE XYZ-space to LMS-space (using CIE CAT02 matrix).
func XyzToLmsMcat02(x, y, z float64) (l, m, s float64) {
	l = 0.7328*x + 0.4296*y - 0.1624*z
	m = -0.7036*x + 1.6974*y + 0.0061*z
	s = 0.0030*x + 0.0136*y + 0.9834*z
	return
}

// LmsMcat02ToXyz converts from LMS-space (using CIE CAT02 matrix) to CIE XYZ-space.
func LmsMcat02ToXyz(l, m, s float64) (x, y, z float64) {
	x = 1.0961*l - 0.2789*m + 0.1827*s
	y = 0.4544*l + 0.4736*m + 0.0721*s
	z = -0.0096*l - 0.0057*m + 1.0153*s
	return
}

// XyzToLmsMhpe converts from CIE XYZ-space to LMS-space (using Hunt-Pointer-Estevez matrix).
func XyzToLmsMhpe(x, y, z float64) (l, m, s float64) {
	l = 0.38971*x + 0.68898*y - 0.07868*z
	m = -0.22981*x + 1.18340*y + 0.04641*z
	s = 0.00000*x + 0.00000*y + 1.00000*z
	return
}

// LmsMhpeToXyz converts from LMS-space (using Hunt-Pointer-Estevez matrix) to CIE XYZ-space.
func LmsMhpeToXyz(l, m, s float64) (x, y, z float64) {
	x = 1.91020*l - 1.11212*m + 0.20191*s
	y = 0.37095*l + 0.62905*m - 0.00001*s
	z = 0.00000*l + 0.00000*m + 1.00000*s
	return
}

// LmsToIpt converts from LMS-space to IPT-space.
func LmsToIpt(l, m, s float64) (i, p, t float64) {
	i = 0.4000*l + 0.4000*m + 0.2000*s
	p = 4.4550*l - 4.8510*m + 0.3960*s
	t = 0.8056*l + 0.3572*m - 1.1628*s
	return
}

// IptToLms converts from IPT-space to LMS-space.
func IptToLms(i, p, t float64) (l, m, s float64) {
	l = 1*i + 0.0976*p + 0.2052*t
	m = 1*i - 0.1139*p + 0.1332*t
	s = 1*i + 0.0326*p - 0.6769*t
	return
}
