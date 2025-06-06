package lox

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// takes a filename, and a FloatArrayObject, and a boolean indicating whether the array contains RGB encoded data
func drawPNGBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {
	if argCount != 3 {
		vm.runTimeError("Invalid argument count to draw_png.")
		return makeNilValue()
	}
	nameVal := vm.stack[arg_stackptr]
	plotData := vm.stack[arg_stackptr+1]
	colourEncoded := vm.stack[arg_stackptr+2]

	if !nameVal.isStringObject() {
		vm.runTimeError("First argument to draw_png must be a string filename")
		return makeNilValue()
	}

	if !plotData.isFloatArrayObject() {
		vm.runTimeError("Second argument to draw_png must be a float array")
		return makeNilValue()
	}
	if !colourEncoded.isBool() {
		vm.runTimeError("Third argument to draw_png must be a boolean")
	}

	fa := plotData.asFloatArray()
	if fa.value.width <= 0 || fa.value.height <= 0 {
		vm.runTimeError("draw_png data must not be empty")
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

func MandelArrayBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 7 {
		vm.runTimeError("Invalid argument count to lox_mandel_array")
		return makeNilValue()
	}
	arrayVal := vm.stack[arg_stackptr]
	hVal := vm.stack[arg_stackptr+1]
	wVal := vm.stack[arg_stackptr+2]
	maxIterVal := vm.stack[arg_stackptr+3]
	xoffsetVal := vm.stack[arg_stackptr+4]
	yoffsetVal := vm.stack[arg_stackptr+5]
	scaleVal := vm.stack[arg_stackptr+6]

	if !(hVal.isInt() && wVal.isInt() && maxIterVal.isInt() && xoffsetVal.isFloat() &&
		yoffsetVal.isFloat() && arrayVal.isFloatArrayObject() && scaleVal.isFloat()) {
		vm.runTimeError("Invalid arguments to lox_mandel_array")
		return makeNilValue()
	}

	array := arrayVal.asFloatArray()
	height := hVal.Int
	width := wVal.Int
	maxIteration := maxIterVal.Int
	xOffset := xoffsetVal.Float
	yOffset := yoffsetVal.Float
	scale := scaleVal.Float

	var brightness float64

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

			if iteration == maxIteration {
				brightness = 0
			} else {
				brightness = float64(iteration) / float64(maxIteration)
				brightness = math.Sqrt(brightness) // smoother contrast
			}
			array.value.set(row, col, brightness)
		}
	}
	return makeNilValue()
}
