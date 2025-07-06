package builtin

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"

	"glox/src/core"
	"glox/src/util"
)

// takes a filename, and a FloatArrayObject, and a boolean indicating whether the array contains RGB encoded data
func DrawPNGBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 {
		vm.RunTimeError("Invalid argument count to draw_png.")
		return core.NIL_VALUE
	}
	nameVal := vm.Stack(arg_stackptr)
	plotData := vm.Stack(arg_stackptr + 1)
	colourEncoded := vm.Stack(arg_stackptr + 2)

	if !nameVal.IsStringObject() {
		vm.RunTimeError("First argument to draw_png must be a string filename")
		return core.NIL_VALUE
	}

	if !IsFloatArrayObject(plotData) {
		vm.RunTimeError("Second argument to draw_png must be a float array")
		return core.NIL_VALUE
	}
	if !colourEncoded.IsBool() {
		vm.RunTimeError("Third argument to draw_png must be a boolean")
	}

	fa := AsFloatArray(plotData)
	if fa.Value.Width <= 0 || fa.Value.Height <= 0 {
		vm.RunTimeError("draw_png data must not be empty")
		return core.NIL_VALUE
	}

	width := fa.Value.Width
	height := fa.Value.Height

	if !colourEncoded.Bool {
		img := image.NewGray(image.Rect(0, 0, width, height))

		var gray uint8

		for y := range height {
			for x := range width {
				val := fa.Value.Get(x, y)
				gray = uint8(min(val*255, 255))
				img.SetGray(x, y, color.Gray{Y: gray})
			}
		}
		file, _ := os.Create(nameVal.AsString().Get())
		defer file.Close()
		_ = png.Encode(file, img)
	} else {
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		for y := range height {
			for x := range width {
				val := fa.Value.Get(x, y)
				r, g, b := util.DecodeRGB(val)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
		file, _ := os.Create(nameVal.AsString().Get())
		defer file.Close()
		_ = png.Encode(file, img)
	}

	return core.NIL_VALUE
}

// args:
// 2D float array for plotting
// height
// width
// max iterations
// x offset
// y offset
// scale

// creates a goroutine for each row of the array, which calculates the mandelbrot set for that row
func MandelArrayBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 7 {
		vm.RunTimeError("Invalid argument count to lox_mandel_array")
		return core.NIL_VALUE
	}
	arrayVal := vm.Stack(arg_stackptr)
	hVal := vm.Stack(arg_stackptr + 1)
	wVal := vm.Stack(arg_stackptr + 2)
	maxIterVal := vm.Stack(arg_stackptr + 3)
	xoffsetVal := vm.Stack(arg_stackptr + 4)
	yoffsetVal := vm.Stack(arg_stackptr + 5)
	scaleVal := vm.Stack(arg_stackptr + 6)

	if !IsFloatArrayObject(arrayVal) {
		vm.RunTimeError("First argument to lox_mandel_array must be a float array")
		return core.NIL_VALUE
	}
	if !hVal.IsInt() {
		vm.RunTimeError("Second argument to lox_mandel_array (height) must be an integer")
		return core.NIL_VALUE
	}
	if !wVal.IsInt() {
		vm.RunTimeError("Third argument to lox_mandel_array (width) must be an integer")
		return core.NIL_VALUE
	}
	if !maxIterVal.IsInt() {
		vm.RunTimeError("Fourth argument to lox_mandel_array (max iterations) must be an integer")
		return core.NIL_VALUE
	}
	if !xoffsetVal.IsFloat() {
		vm.RunTimeError("Fifth argument to lox_mandel_array (x offset) must be a float")
		return core.NIL_VALUE
	}
	if !yoffsetVal.IsFloat() {
		vm.RunTimeError("Sixth argument to lox_mandel_array (y offset) must be a float")
		return core.NIL_VALUE
	}
	if !scaleVal.IsFloat() {
		vm.RunTimeError("Seventh argument to lox_mandel_array (scale) must be a float")
		return core.NIL_VALUE
	}

	array := AsFloatArray(arrayVal)
	height := hVal.Int
	width := wVal.Int
	maxIteration := maxIterVal.Int
	xOffset := xoffsetVal.Float
	yOffset := yoffsetVal.Float
	scale := scaleVal.Float

	var wg sync.WaitGroup
	for row := range height {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			mandelbrotCalcRow(row, width, height, maxIteration, scale, xOffset, yOffset, array)
		}(row)
	}
	wg.Wait()
	return core.NIL_VALUE
}

// mandelbrotCalcRow calculates a single row of the mandelbrot set and stores the result in the provided FloatArrayObject
// rows are calculated in parallel using goroutines
func mandelbrotCalcRow(row, width, height, maxIteration int, scale, xOffset, yOffset float64, array *FloatArrayObject) {

	const periodLength = 20

	// Fix aspect ratio distortion by using consistent coordinate system
	// Use the larger dimension as reference to maintain square aspect ratio
	maxDim := max(width, height)

	y0 := scale*(float64(row)-float64(height)/2)/float64(maxDim) + yOffset

	for col := 0; col < width; col++ {

		var xold float64 = -1.0
		var yold float64 = -1.0
		var period int = 0
		x0 := scale*(float64(col)-float64(width)/2)/float64(maxDim) + xOffset

		x, y := 0.0, 0.0
		iteration := 0

		for (x*x+y*y <= 4) && (iteration < maxIteration) {
			// optimisation : check for periodicity
			if x == xold && y == yold {
				iteration = maxIteration
				goto bailout
			}
			period++
			if period > periodLength {
				period = 0
				xold = x
				yold = y
			}

			xtemp := x*x - y*y + x0
			y = 2*x*y + y0
			x = xtemp
			iteration++
		}
	bailout:
		var colour float64
		if iteration == maxIteration {
			colour = util.EncodeRGB(0, 0, 0)
		} else {

			//https://en.m.wikipedia.org/wiki/Plotting_algorithms_for_the_Mandelbrot_set

			s := float64(iteration) / float64(maxIteration)
			v := 1.0 - math.Pow(math.Cos(math.Pi*s), 2.0)
			lum := 90 - (50 * v)
			// chroma := 28+(75-(75*v))
			// hue := int32(math.Pow(360*s, 1.5)) % 360

			zn := math.Sqrt(x*x + y*y)
			nu := math.Log2(math.Log2(zn))
			smooth := float64(iteration) + 1 - nu

			// Use a fixed scale for hue cycling, e.g. 20 or 30
			hue := math.Mod(smooth, 360)
			chroma := 70.0
			//lum := 80.0

			r, g, b := util.HCLToRGB255(float64(hue), float64(chroma), float64(lum))
			colour = util.EncodeRGB(int(r), int(g), int(b))
		}
		array.Value.Set(row, col, colour)
	}
}

// args:
// 2D float array for plotting
// height
// width
// max iterations
// cx (real part of Julia constant)
// cy (imaginary part of Julia constant)
// scale

// creates a goroutine for each row of the array, which calculates the julia set for that row
// args:
// 2D float array for plotting
// height
// width
// max iterations
// cx (real part of Julia constant)
// cy (imaginary part of Julia constant)
// scale
// xoffset (center x)
// yoffset (center y)
func JuliaArrayBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 9 {
		vm.RunTimeError("Invalid argument count to lox_julia_array (expected 9)")
		return core.NIL_VALUE
	}
	arrayVal := vm.Stack(arg_stackptr)
	wVal := vm.Stack(arg_stackptr + 1)
	hVal := vm.Stack(arg_stackptr + 2)
	maxIterVal := vm.Stack(arg_stackptr + 3)
	cxVal := vm.Stack(arg_stackptr + 4)
	cyVal := vm.Stack(arg_stackptr + 5)
	scaleVal := vm.Stack(arg_stackptr + 6)
	xoffsetVal := vm.Stack(arg_stackptr + 7)
	yoffsetVal := vm.Stack(arg_stackptr + 8)

	if !IsFloatArrayObject(arrayVal) {
		vm.RunTimeError("First argument to lox_julia_array must be a float array")
		return core.NIL_VALUE
	}
	if !wVal.IsInt() {
		vm.RunTimeError("Second argument to lox_julia_array (width) must be an integer")
		return core.NIL_VALUE
	}
	if !hVal.IsInt() {
		vm.RunTimeError("Third argument to lox_julia_array (height) must be an integer")
		return core.NIL_VALUE
	}

	if !maxIterVal.IsInt() {
		vm.RunTimeError("Fourth argument to lox_julia_array (max iterations) must be an integer")
		return core.NIL_VALUE
	}
	if !cxVal.IsFloat() {
		vm.RunTimeError("Fifth argument to lox_julia_array (cx) must be a float")
		return core.NIL_VALUE
	}
	if !cyVal.IsFloat() {
		vm.RunTimeError("Sixth argument to lox_julia_array (cy) must be a float")
		return core.NIL_VALUE
	}
	if !scaleVal.IsFloat() {
		vm.RunTimeError("Seventh argument to lox_julia_array (scale) must be a float")
		return core.NIL_VALUE
	}
	if !xoffsetVal.IsFloat() {
		vm.RunTimeError("Eighth argument to lox_julia_array (xoffset) must be a float")
		return core.NIL_VALUE
	}
	if !yoffsetVal.IsFloat() {
		vm.RunTimeError("Ninth argument to lox_julia_array (yoffset) must be a float")
		return core.NIL_VALUE
	}

	array := AsFloatArray(arrayVal)
	height := hVal.Int
	width := wVal.Int
	maxIteration := maxIterVal.Int
	cx := cxVal.Float
	cy := cyVal.Float
	scale := scaleVal.Float
	xOffset := xoffsetVal.Float
	yOffset := yoffsetVal.Float

	var wg sync.WaitGroup
	for row := range height {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			juliaCalcRow(row, width, height, maxIteration, scale, cx, cy, xOffset, yOffset, array)
		}(row)
	}
	wg.Wait()
	return core.NIL_VALUE
}

// juliaCalcRow calculates a single row of the julia set and stores the result in the provided FloatArrayObject
// rows are calculated in parallel using goroutines
func juliaCalcRow(row, width, height, maxIteration int, scale, cx, cy, xOffset, yOffset float64, array *FloatArrayObject) {
	// Use the larger dimension as reference to maintain square aspect ratio
	maxDim := max(width, height)
	y0 := scale*(float64(row)-float64(height)/2)/float64(maxDim) + yOffset
	for col := 0; col < width; col++ {
		x0 := scale*(float64(col)-float64(width)/2)/float64(maxDim) + xOffset
		zx, zy := x0, y0
		iteration := 0
		for (zx*zx+zy*zy <= 4) && (iteration < maxIteration) {
			xtemp := zx*zx - zy*zy + cx
			zy = 2*zx*zy + cy
			zx = xtemp
			iteration++
		}
		var colour float64
		if iteration == maxIteration {
			colour = util.EncodeRGB(0, 0, 0)
		} else {
			s := float64(iteration) / float64(maxIteration)
			v := 1.0 - math.Pow(math.Cos(math.Pi*s), 2.0)
			lum := 90 - (50 * v)
			zn := math.Sqrt(zx*zx + zy*zy)
			nu := math.Log2(math.Log2(zn))
			smooth := float64(iteration) + 1 - nu
			hue := math.Mod(smooth, 360)
			chroma := 70.0
			r, g, b := util.HCLToRGB255(float64(hue), float64(chroma), float64(lum))
			colour = util.EncodeRGB(int(r), int(g), int(b))
		}
		array.Value.Set(col, row, colour)
	}
}
