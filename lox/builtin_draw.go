package lox

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// takes a filename, and a FloatArrayObject, and a boolean indicating whether the array contains RGB encoded data
func drawPNGBuiltIn(argCount int, arg_stackptr int, vm VMContext) Value {
	if argCount != 3 {
		vm.RunTimeError("Invalid argument count to draw_png.")
		return makeNilValue()
	}
	nameVal := vm.Stack(arg_stackptr)
	plotData := vm.Stack(arg_stackptr + 1)
	colourEncoded := vm.Stack(arg_stackptr + 2)

	if !nameVal.isStringObject() {
		vm.RunTimeError("First argument to draw_png must be a string filename")
		return makeNilValue()
	}

	if !plotData.isFloatArrayObject() {
		vm.RunTimeError("Second argument to draw_png must be a float array")
		return makeNilValue()
	}
	if !colourEncoded.isBool() {
		vm.RunTimeError("Third argument to draw_png must be a boolean")
	}

	fa := plotData.asFloatArray()
	if fa.value.width <= 0 || fa.value.height <= 0 {
		vm.RunTimeError("draw_png data must not be empty")
		return makeNilValue()
	}

	width := fa.value.width
	height := fa.value.height

	if !colourEncoded.Bool {
		img := image.NewGray(image.Rect(0, 0, width, height))

		var gray uint8

		for y := range height {
			for x := range width {
				val := fa.value.get(x, y)
				gray = uint8(min(val*255, 255))
				img.SetGray(x, y, color.Gray{Y: gray})
			}
		}
		file, _ := os.Create(nameVal.asString().get())
		defer file.Close()
		_ = png.Encode(file, img)
	} else {
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		for y := range height {
			for x := range width {
				val := fa.value.get(x, y)
				r, g, b := DecodeRGB(val)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
		file, _ := os.Create(nameVal.asString().get())
		defer file.Close()
		_ = png.Encode(file, img)
	}

	return makeNilValue()
}

// args:
// 2D float array for plotting
// height
// width
// max iterations
// x offset
// y offset
// scale
// 1D float array (RGB encoded) for colour mapping
func MandelArrayBuiltIn(argCount int, arg_stackptr int, vm VMContext) Value {

	if argCount != 8 {
		vm.RunTimeError("Invalid argument count to lox_mandel_array")
		return makeNilValue()
	}
	arrayVal := vm.Stack(arg_stackptr)
	hVal := vm.Stack(arg_stackptr + 1)
	wVal := vm.Stack(arg_stackptr + 2)
	maxIterVal := vm.Stack(arg_stackptr + 3)
	xoffsetVal := vm.Stack(arg_stackptr + 4)
	yoffsetVal := vm.Stack(arg_stackptr + 5)
	scaleVal := vm.Stack(arg_stackptr + 6)
	colourVal := vm.Stack(arg_stackptr + 7)

	if !(hVal.isInt() && wVal.isInt() && maxIterVal.isInt() && xoffsetVal.isFloat() &&
		yoffsetVal.isFloat() && arrayVal.isFloatArrayObject() && scaleVal.isFloat() &&
		colourVal.isFloatArrayObject()) {
		vm.RunTimeError("Invalid arguments to lox_mandel_array")
		return makeNilValue()
	}

	array := arrayVal.asFloatArray()
	height := hVal.Int
	width := wVal.Int
	maxIteration := maxIterVal.Int
	xOffset := xoffsetVal.Float
	yOffset := yoffsetVal.Float
	scale := scaleVal.Float
	colours := colourVal.asFloatArray()

	for row := 0; row < height; row = row + 1 {
		for col := 0; col < width; col = col + 1 {

			x0 := scale*(float64(col)-float64(width)/2)/float64(width) + xOffset
			y0 := scale*(float64(row)-float64(height)/2)/float64(height) + yOffset
			x, y := 0.0, 0.0
			iteration := 0

			for (x*x+y*y <= 4) && (iteration < maxIteration) {
				xtemp := x*x - y*y + x0
				y = 2*x*y + y0
				x = xtemp
				iteration++
			}

			var colour float64
			if iteration == maxIteration {
				colour = EncodeRGB(0, 0, 0)
			} else {
				index := iteration
				if index >= colours.value.width {
					index = colours.value.width - 1
				}
				colour = colours.value.get(index, 0)

			}
			array.value.set(row, col, colour)
		}
	}
	return makeNilValue()
}
