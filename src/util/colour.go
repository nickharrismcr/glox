package util

import (
	"math"
)

func EncodeRGB(r, g, b int) float64 {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		panic("RGB values must be between 0 and 255")
	}
	return float64(uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

func DecodeRGB(color float64) (uint8, uint8, uint8) {
	v := uint32(color)
	r := uint8((v >> 16) & 0xFF)
	g := uint8((v >> 8) & 0xFF)
	b := uint8(v & 0xFF)
	return r, g, b
}

// covert HCL to RGB 255
func HCLToRGB255(h, c, l float64) (uint8, uint8, uint8) {
	// Convert HCL to Lab
	hr := h * math.Pi / 180.0
	a := math.Cos(hr) * c
	b := math.Sin(hr) * c

	// Lab to XYZ
	y := (l + 16.0) / 116.0
	x := a/500.0 + y
	z := y - b/200.0

	refX, refY, refZ := 95.047, 100.000, 108.883

	x3 := math.Pow(x, 3)
	y3 := math.Pow(y, 3)
	z3 := math.Pow(z, 3)

	if y3 > 0.008856 {
		y = y3
	} else {
		y = (y - 16.0/116.0) / 7.787
	}
	if x3 > 0.008856 {
		x = x3
	} else {
		x = (x - 16.0/116.0) / 7.787
	}
	if z3 > 0.008856 {
		z = z3
	} else {
		z = (z - 16.0/116.0) / 7.787
	}

	x = x * refX
	y = y * refY
	z = z * refZ

	// XYZ to linear RGB
	r := x*0.032406 + y*-0.015372 + z*-0.004986
	g := x*-0.009689 + y*0.018758 + z*0.000415
	b = x*0.000557 + y*-0.002040 + z*0.010570

	// Linear RGB to sRGB
	r = xyzToSRGB(r)
	g = xyzToSRGB(g)
	b = xyzToSRGB(b)

	// Clamp and scale to [0,255]
	r8 := uint8(clamp(r, 0, 1) * 255)
	g8 := uint8(clamp(g, 0, 1) * 255)
	b8 := uint8(clamp(b, 0, 1) * 255)

	return r8, g8, b8
}

func xyzToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
