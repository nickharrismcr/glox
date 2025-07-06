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

	// Calculate optimal number of blocks based on image size and CPU cores
	// For larger images, use smaller blocks for better parallelism
	// Target: 32x32 to 64x64 pixels per block for high-resolution images
	minBlockSize := 32
	if width*height < 160000 { // For smaller images (< 400x400)
		minBlockSize = 64
	}
	maxBlocks := 512 // Increased for better parallelism on large images

	// Calculate blocks based on image dimensions
	blocksX := max(1, min(width/minBlockSize, int(math.Sqrt(float64(maxBlocks)))))
	blocksY := max(1, min(height/minBlockSize, int(math.Sqrt(float64(maxBlocks)))))

	blockHeight := (height + blocksY - 1) / blocksY
	blockWidth := (width + blocksX - 1) / blocksX

	// Precompute color table for all possible iterations
	colorTable := precomputeColorTable(maxIteration)

	var wg sync.WaitGroup
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			startRow := by * blockHeight
			endRow := min((by+1)*blockHeight, height)
			startCol := bx * blockWidth
			endCol := min((bx+1)*blockWidth, width)
			if startRow >= endRow || startCol >= endCol {
				continue
			}
			wg.Add(1)
			go func(sr, er, sc, ec int) {
				defer wg.Done()
				mandelbrotCalcBlock(sr, er, sc, ec, width, height, maxIteration, scale, xOffset, yOffset, array, colorTable)
			}(startRow, endRow, startCol, endCol)
		}
	}
	wg.Wait()
	return core.NIL_VALUE
}

// mandelbrotCalcBlock calculates a block of the mandelbrot set and stores the result in the provided FloatArrayObject
// blocks are calculated in parallel using goroutines, with precomputed color table for fast lookup
func mandelbrotCalcBlock(startRow, endRow, startCol, endCol, width, height, maxIteration int, scale, xOffset, yOffset float64, array *FloatArrayObject, colorTable []float64) {
	// Use the larger dimension as reference to maintain square aspect ratio
	maxDim := max(width, height)

	// Use float64 for better precision in deep zooms
	maxDimFloat := float64(maxDim)
	heightFloat := float64(height)
	widthFloat := float64(width)

	// Precompute constants outside loops
	scaleOverMaxDim := scale / maxDimFloat
	halfHeight := heightFloat / 2.0
	halfWidth := widthFloat / 2.0

	// Direct access to the underlying slice for better performance
	data := array.Value.Data
	arrayWidth := array.Value.Width

	for row := startRow; row < endRow; row++ {
		y0 := scaleOverMaxDim*(float64(row)-halfHeight) + yOffset
		for col := startCol; col < endCol; col++ {
			x0 := scaleOverMaxDim*(float64(col)-halfWidth) + xOffset

			x, y := 0.0, 0.0
			iteration := 0

			// Early bailout for known regions (effectiveness depends on zoom level)
			// Temporarily disabled to debug black circle issue
			/*
				if !isDeepZoom && (isInMainCardioid(x0, y0) || isInPeriod2Bulb(x0, y0)) {
					// Points in cardioid or bulb are definitely in the set
					iteration = maxIteration
				} else if isMediumZoom && isInMiniBulb(x0, y0) {
					// Check smaller bulbs at medium zoom
					iteration = maxIteration
				} else {
			*/
			{
				// Basic Mandelbrot calculation - no periodicity checking
				for (x*x+y*y <= 4.0) && (iteration < maxIteration) {
					xtemp := x*x - y*y + x0
					y = 2.0*x*y + y0
					x = xtemp
					iteration++
				}
			}

			// Direct slice access instead of array.Value.Set() for better performance
			// Note: FloatArray uses y*Width+x indexing, so row*arrayWidth+col is correct
			data[row*arrayWidth+col] = colorTable[iteration]
		}
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

	// Calculate optimal number of blocks based on image size and CPU cores
	// For larger images, use smaller blocks for better parallelism
	// Target: 32x32 to 64x64 pixels per block for high-resolution images
	minBlockSize := 32
	if width*height < 160000 { // For smaller images (< 400x400)
		minBlockSize = 64
	}
	maxBlocks := 512 // Increased for better parallelism on large images

	// Calculate blocks based on image dimensions
	blocksX := max(1, min(width/minBlockSize, int(math.Sqrt(float64(maxBlocks)))))
	blocksY := max(1, min(height/minBlockSize, int(math.Sqrt(float64(maxBlocks)))))

	blockHeight := (height + blocksY - 1) / blocksY
	blockWidth := (width + blocksX - 1) / blocksX

	// Precompute color table for all possible iterations
	colorTable := precomputeColorTable(maxIteration)

	var wg sync.WaitGroup
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			startRow := by * blockHeight
			endRow := min((by+1)*blockHeight, height)
			startCol := bx * blockWidth
			endCol := min((bx+1)*blockWidth, width)
			if startRow >= endRow || startCol >= endCol {
				continue
			}
			wg.Add(1)
			go func(sr, er, sc, ec int) {
				defer wg.Done()
				juliaCalcBlock(sr, er, sc, ec, width, height, maxIteration, scale, cx, cy, xOffset, yOffset, array, colorTable)
			}(startRow, endRow, startCol, endCol)
		}
	}
	wg.Wait()
	return core.NIL_VALUE
}

// juliaCalcBlock calculates a block of the julia set and stores the result in the provided FloatArrayObject
// blocks are calculated in parallel using goroutines, with precomputed color table for fast lookup
func juliaCalcBlock(startRow, endRow, startCol, endCol, width, height, maxIteration int, scale, cx, cy, xOffset, yOffset float64, array *FloatArrayObject, colorTable []float64) {
	// Use the larger dimension as reference to maintain square aspect ratio
	maxDim := max(width, height)

	// Convert to float32 for better performance
	scale32 := float32(scale)
	cx32 := float32(cx)
	cy32 := float32(cy)
	xOffset32 := float32(xOffset)
	yOffset32 := float32(yOffset)
	maxDim32 := float32(maxDim)
	height32 := float32(height)
	width32 := float32(width)

	// Precompute constants outside loops
	scaleOverMaxDim := scale32 / maxDim32
	halfHeight := height32 / 2
	halfWidth := width32 / 2

	// Direct access to the underlying slice for better performance
	data := array.Value.Data
	arrayWidth := array.Value.Width

	for row := startRow; row < endRow; row++ {
		y0 := scaleOverMaxDim*(float32(row)-halfHeight) + yOffset32
		for col := startCol; col < endCol; col++ {
			x0 := scaleOverMaxDim*(float32(col)-halfWidth) + xOffset32
			zx, zy := x0, y0
			iteration := 0

			// Basic Julia set calculation - no periodicity checking needed
			for (zx*zx+zy*zy <= 4.0) && (iteration < maxIteration) {
				xtemp := zx*zx - zy*zy + cx32
				zy = 2*zx*zy + cy32
				zx = xtemp
				iteration++
			}

			// Direct slice access instead of array.Value.Set() for better performance
			data[row*arrayWidth+col] = colorTable[iteration]
		}
	}
}

// precomputeColorTable creates a lookup table for fractal colors to avoid repeated calculations
func precomputeColorTable(maxIteration int) []float64 {
	colorTable := make([]float64, maxIteration+1)
	maxIter32 := float32(maxIteration)

	for i := 0; i <= maxIteration; i++ {
		if i == maxIteration {
			colorTable[i] = util.EncodeRGB(0, 0, 0)
		} else {
			s := float32(i) / maxIter32
			v := 1.0 - float32(math.Pow(math.Cos(math.Pi*float64(s)), 2.0))
			lum := 90 - (50 * v)
			smooth := float32(i) + 1 - 0.0 // simplified without zn calculation
			hue := float32(math.Mod(float64(smooth), 360))
			chroma := float32(70.0)
			r, g, b := util.HCLToRGB255(float64(hue), float64(chroma), float64(lum))
			colorTable[i] = util.EncodeRGB(int(r), int(g), int(b))
		}
	}
	return colorTable
}

// isInMainCardioid checks if a point is inside the main cardioid of the Mandelbrot set
// Points inside the main cardioid are guaranteed to be in the set
func isInMainCardioid(x, y float64) bool {
	// Standard cardioid test for the main body of the Mandelbrot set
	// Let q = (x-1/4)^2 + y^2
	// The cardioid test is: q*(q + (x-1/4)) < 1/4*y^2
	dx := x - 0.25
	q := dx*dx + y*y
	return q*(q+dx) < 0.25*y*y
}

// isInPeriod2Bulb checks if a point is inside the period-2 bulb
// Points inside this bulb are guaranteed to be in the set
func isInPeriod2Bulb(x, y float64) bool {
	// Period-2 bulb is centered at (-1, 0) with radius 0.25
	dx := x + 1.0
	dy := y
	return dx*dx+dy*dy < 0.0625 // 0.25^2
}

// isInMiniBulb checks if a point is inside one of the smaller bulbs
// This is more effective at medium zoom levels
func isInMiniBulb(x, y float64) bool {
	// Check period-3 bulb (approximate)
	dx := x + 1.25
	dy := y
	if dx*dx+dy*dy < 0.01 { // Small bulb near (-1.25, 0)
		return true
	}

	// Check other small bulbs along the real axis
	dx = x + 0.75
	if dx*dx+dy*dy < 0.001 { // Very small bulb near (-0.75, 0)
		return true
	}

	return false
}
