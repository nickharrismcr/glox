package builtin

import (
	"glox/src/core"
	"math"
	"math/rand"
)

// Color utility builtin functions for colour_utils module

func ColourUtilsFadeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 4 {
		vm.RunTimeError("fade expects 4 arguments (r, g, b, alpha)")
		return core.NIL_VALUE
	}

	rVal := vm.Stack(arg_stackptr)
	gVal := vm.Stack(arg_stackptr + 1)
	bVal := vm.Stack(arg_stackptr + 2)
	alphaVal := vm.Stack(arg_stackptr + 3)

	if !rVal.IsNumber() || !gVal.IsNumber() || !bVal.IsNumber() || !alphaVal.IsNumber() {
		vm.RunTimeError("fade arguments must be numbers")
		return core.NIL_VALUE
	}

	r := rVal.AsFloat()
	g := gVal.AsFloat()
	b := bVal.AsFloat()
	alpha := alphaVal.AsFloat()

	// Clamp inputs to valid ranges
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	// Apply alpha to each component
	newR := int(r * alpha)
	newG := int(g * alpha)
	newB := int(b * alpha)

	// Return vec4 directly
	return core.MakeVec4Value(float64(newR), float64(newG), float64(newB), 255.0, false)
}

func ColourUtilsTintBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 6 {
		vm.RunTimeError("tint expects 6 arguments (r1, g1, b1, r2, g2, b2)")
		return core.NIL_VALUE
	}

	r1Val := vm.Stack(arg_stackptr)
	g1Val := vm.Stack(arg_stackptr + 1)
	b1Val := vm.Stack(arg_stackptr + 2)
	r2Val := vm.Stack(arg_stackptr + 3)
	g2Val := vm.Stack(arg_stackptr + 4)
	b2Val := vm.Stack(arg_stackptr + 5)

	if !r1Val.IsNumber() || !g1Val.IsNumber() || !b1Val.IsNumber() ||
		!r2Val.IsNumber() || !g2Val.IsNumber() || !b2Val.IsNumber() {
		vm.RunTimeError("tint arguments must be numbers")
		return core.NIL_VALUE
	}

	r1 := r1Val.AsFloat()
	g1 := g1Val.AsFloat()
	b1 := b1Val.AsFloat()
	r2 := r2Val.AsFloat()
	g2 := g2Val.AsFloat()
	b2 := b2Val.AsFloat()

	// Clamp inputs to 0-255
	if r1 < 0 {
		r1 = 0
	}
	if r1 > 255 {
		r1 = 255
	}
	if g1 < 0 {
		g1 = 0
	}
	if g1 > 255 {
		g1 = 255
	}
	if b1 < 0 {
		b1 = 0
	}
	if b1 > 255 {
		b1 = 255
	}
	if r2 < 0 {
		r2 = 0
	}
	if r2 > 255 {
		r2 = 255
	}
	if g2 < 0 {
		g2 = 0
	}
	if g2 > 255 {
		g2 = 255
	}
	if b2 < 0 {
		b2 = 0
	}
	if b2 > 255 {
		b2 = 255
	}
	// Apply tint by multiplying components
	newR := int((r1 * r2) / 255.0)
	newG := int((g1 * g2) / 255.0)
	newB := int((b1 * b2) / 255.0)

	// Return vec4 directly
	return core.MakeVec4Value(float64(newR), float64(newG), float64(newB), 255.0, false)
}

func ColourUtilsBrightnessBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 4 {
		vm.RunTimeError("brightness expects 4 arguments (r, g, b, factor)")
		return core.NIL_VALUE
	}

	rVal := vm.Stack(arg_stackptr)
	gVal := vm.Stack(arg_stackptr + 1)
	bVal := vm.Stack(arg_stackptr + 2)
	factorVal := vm.Stack(arg_stackptr + 3)

	if !rVal.IsNumber() || !gVal.IsNumber() || !bVal.IsNumber() || !factorVal.IsNumber() {
		vm.RunTimeError("brightness arguments must be numbers")
		return core.NIL_VALUE
	}

	r := rVal.AsFloat()
	g := gVal.AsFloat()
	b := bVal.AsFloat()
	factor := factorVal.AsFloat()

	// Clamp inputs to 0-255
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}
	// Apply brightness factor and clamp to 0-255
	newR := int(math.Min(255, math.Max(0, r*factor)))
	newG := int(math.Min(255, math.Max(0, g*factor)))
	newB := int(math.Min(255, math.Max(0, b*factor)))

	// Return vec4 directly
	return core.MakeVec4Value(float64(newR), float64(newG), float64(newB), 255.0, false)
}

func ColourUtilsLerpBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 7 {
		vm.RunTimeError("lerp expects 7 arguments (r1, g1, b1, r2, g2, b2, amount)")
		return core.NIL_VALUE
	}

	r1Val := vm.Stack(arg_stackptr)
	g1Val := vm.Stack(arg_stackptr + 1)
	b1Val := vm.Stack(arg_stackptr + 2)
	r2Val := vm.Stack(arg_stackptr + 3)
	g2Val := vm.Stack(arg_stackptr + 4)
	b2Val := vm.Stack(arg_stackptr + 5)
	amountVal := vm.Stack(arg_stackptr + 6)

	if !r1Val.IsNumber() || !g1Val.IsNumber() || !b1Val.IsNumber() ||
		!r2Val.IsNumber() || !g2Val.IsNumber() || !b2Val.IsNumber() || !amountVal.IsNumber() {
		vm.RunTimeError("lerp arguments must be numbers")
		return core.NIL_VALUE
	}

	r1 := r1Val.AsFloat()
	g1 := g1Val.AsFloat()
	b1 := b1Val.AsFloat()
	r2 := r2Val.AsFloat()
	g2 := g2Val.AsFloat()
	b2 := b2Val.AsFloat()
	amount := amountVal.AsFloat()

	// Clamp amount between 0 and 1
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}

	// Clamp RGB values to 0-255
	if r1 < 0 {
		r1 = 0
	}
	if r1 > 255 {
		r1 = 255
	}
	if g1 < 0 {
		g1 = 0
	}
	if g1 > 255 {
		g1 = 255
	}
	if b1 < 0 {
		b1 = 0
	}
	if b1 > 255 {
		b1 = 255
	}
	if r2 < 0 {
		r2 = 0
	}
	if r2 > 255 {
		r2 = 255
	}
	if g2 < 0 {
		g2 = 0
	}
	if g2 > 255 {
		g2 = 255
	}
	if b2 < 0 {
		b2 = 0
	}
	if b2 > 255 {
		b2 = 255
	}
	// Linear interpolation
	newR := int(r1 + (r2-r1)*amount)
	newG := int(g1 + (g2-g1)*amount)
	newB := int(b1 + (b2-b1)*amount)

	// Return vec4 directly
	return core.MakeVec4Value(float64(newR), float64(newG), float64(newB), 255.0, false)
}

func ColourUtilsHSVToRGBBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 {
		vm.RunTimeError("hsv_to_rgb expects 3 arguments (h, s, v)")
		return core.NIL_VALUE
	}

	hVal := vm.Stack(arg_stackptr)
	sVal := vm.Stack(arg_stackptr + 1)
	vVal := vm.Stack(arg_stackptr + 2)

	if !hVal.IsNumber() || !sVal.IsNumber() || !vVal.IsNumber() {
		vm.RunTimeError("hsv_to_rgb arguments must be numbers")
		return core.NIL_VALUE
	}

	h := hVal.AsFloat()
	s := sVal.AsFloat()
	v := vVal.AsFloat()

	// Normalize inputs
	h = math.Mod(h, 360.0) // Hue wraps around
	if s < 0 {
		s = 0
	}
	if s > 1 {
		s = 1
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	// HSV to RGB conversion
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64

	if h >= 0 && h < 60 {
		r, g, b = c, x, 0
	} else if h >= 60 && h < 120 {
		r, g, b = x, c, 0
	} else if h >= 120 && h < 180 {
		r, g, b = 0, c, x
	} else if h >= 180 && h < 240 {
		r, g, b = 0, x, c
	} else if h >= 240 && h < 300 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	// Convert to 0-255 range
	newR := int((r + m) * 255)
	newG := int((g + m) * 255)
	newB := int((b + m) * 255)

	// Return vec4 directly
	return core.MakeVec4Value(float64(newR), float64(newG), float64(newB), 255.0, false)
}

func ColourUtilsRandomBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("random expects 0 arguments")
		return core.NIL_VALUE
	}

	// Generate random RGB components
	r := rand.Intn(256)
	g := rand.Intn(256)
	b := rand.Intn(256)

	// Return vec4 directly
	return core.MakeVec4Value(float64(r), float64(g), float64(b), 255.0, false)
}
