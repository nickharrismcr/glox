package lox

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// takes a filename, and a FloatArrayObject Value
func drawPNGBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {
	if argCount != 2 {
		vm.runTimeError("Invalid argument count to draw_png.")
		return makeNilValue()
	}
	nameVal := vm.stack[arg_stackptr]
	plotData := vm.stack[arg_stackptr+1]

	if !nameVal.isStringObject() {
		vm.runTimeError("First argument to draw_png must be a string filename")
		return makeNilValue()
	}

	if !plotData.isFloatArrayObject() {
		vm.runTimeError("Second argument to draw_png must be a float array")
		return makeNilValue()
	}

	fa := plotData.asFloatArray()
	if fa.Value.Width <= 0 || fa.Value.Height <= 0 {
		vm.runTimeError("draw_png data must not be empty")
		return makeNilValue()
	}

	width := fa.Value.Width
	height := fa.Value.Height

	img := image.NewGray(image.Rect(0, 0, width, height))

	var gray uint8
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := fa.Value.Get(x, y)
			gray = uint8(min(val*255, 255))
			img.SetGray(x, y, color.Gray{Y: gray})
		}
	}

	file, _ := os.Create(nameVal.asString().get())
	defer file.Close()
	_ = png.Encode(file, img)
	return makeNilValue()
}

func MandelBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 5 {
		vm.runTimeError("Invalid argument count to lox_mandel.")
		return makeNilValue()
	}
	ii := vm.stack[arg_stackptr]
	jj := vm.stack[arg_stackptr+1]
	h := vm.stack[arg_stackptr+2]
	w := vm.stack[arg_stackptr+3]
	max := vm.stack[arg_stackptr+4]

	if ii.Type != VAL_INT || jj.Type != VAL_INT || h.Type != VAL_INT || w.Type != VAL_INT || max.Type != VAL_INT {
		vm.runTimeError("Invalid arguments to lox_mandel")
		return makeNilValue()
	}

	i := ii.Int
	j := jj.Int
	height := h.Int
	width := w.Int
	maxIteration := max.Int

	x0 := 4.0*(float64(i)-float64(height)/2)/float64(height) - 1.0
	y0 := 4.0 * (float64(j) - float64(width)/2) / float64(width)
	x, y := 0.0, 0.0
	iteration := 0

	for (x*x+y*y <= 4) && (iteration < maxIteration) {
		xtemp := x*x - y*y + x0
		y = 2*x*y + y0
		x = xtemp
		iteration++
	}

	return makeIntValue(iteration, false)

}

func MandelArrayBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 4 {
		vm.runTimeError("Invalid argument count to lox_mandel_array")
		return makeNilValue()
	}
	arrayVal := vm.stack[arg_stackptr]
	hVal := vm.stack[arg_stackptr+1]
	wVal := vm.stack[arg_stackptr+2]
	maxIterVal := vm.stack[arg_stackptr+3]

	if hVal.Type != VAL_INT || wVal.Type != VAL_INT || maxIterVal.Type != VAL_INT || !arrayVal.isFloatArrayObject() {
		vm.runTimeError("Invalid arguments to lox_mandel_array")
		return makeNilValue()
	}

	array := arrayVal.asFloatArray()
	height := hVal.Int
	width := wVal.Int
	maxIteration := maxIterVal.Int
	var brightness float64

	for row := 0; row < height; row = row + 1 {
		for col := 0; col < width; col = col + 1 {

			y0 := 4.0*(float64(row)-float64(height)/2)/float64(height) - 1.0
			x0 := 4.0 * (float64(col) - float64(width)/2) / float64(width)
			x, y := 0.0, 0.0
			iteration := 0

			for (x*x+y*y <= 4) && (iteration < maxIteration) {
				xtemp := x*x - y*y + y0
				y = 2*x*y + x0
				x = xtemp
				iteration++
			}

			if iteration == maxIteration {
				brightness = 0
			} else {
				brightness = float64(iteration) / float64(maxIteration)
				brightness = math.Sqrt(brightness) // smoother contrast
			}
			array.Value.Set(row, col, brightness)
		}
	}
	return makeNilValue()
}
