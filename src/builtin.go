package lox

import (
	"fmt"
	"glox/src/builtin"
	"glox/src/core"
	"glox/src/debug"
	"glox/src/util"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

func defineBuiltIn(vm *VM, module string, name string, fn core.BuiltInFn) {
	// Add the built-in to the specified module namespace (environment)

	if module != "" {
		addBuiltInModuleFunction(vm, module, name, fn)
	} else {
		vm.builtIns[core.InternName(name)] = core.MakeObjectValue(core.MakeBuiltInObject(fn), false)
	}

}

func DefineBuiltIns(vm *VM) {

	makeBuiltInModule(vm, "sys")
	makeBuiltInModule(vm, "inspect")
	makeBuiltInModule(vm, "colour_utils")
	makeBuiltInModule(vm, "os")

	core.Log(core.INFO, "Defining built-in functions")

	defineBuiltIn(vm, "sys", "args", argsBuiltIn)
	defineBuiltIn(vm, "sys", "clock", clockBuiltIn)
	defineBuiltIn(vm, "", "type", typeBuiltIn)
	defineBuiltIn(vm, "", "len", lenBuiltIn)
	defineBuiltIn(vm, "", "_sin", sinBuiltIn)
	defineBuiltIn(vm, "", "_cos", cosBuiltIn)
	defineBuiltIn(vm, "", "_tan", tanBuiltIn)
	defineBuiltIn(vm, "", "_sqrt", sqrtBuiltIn)
	defineBuiltIn(vm, "", "_pow", powBuiltIn)
	defineBuiltIn(vm, "", "append", appendBuiltIn)
	defineBuiltIn(vm, "", "float", floatBuiltIn)
	defineBuiltIn(vm, "", "int", intBuiltIn)
	defineBuiltIn(vm, "", "lox_mandel_array", builtin.MandelArrayBuiltIn)
	defineBuiltIn(vm, "", "lox_julia_array", builtin.JuliaArrayBuiltIn)
	defineBuiltIn(vm, "", "draw_png", builtin.DrawPNGBuiltIn)
	defineBuiltIn(vm, "", "replace", replaceBuiltIn)
	defineBuiltIn(vm, "", "format", formatBuiltIn)
	defineBuiltIn(vm, "", "sleep", sleepBuiltIn)
	defineBuiltIn(vm, "", "range", rangeBuiltIn)
	defineBuiltIn(vm, "", "rand", randBuiltIn)
	defineBuiltIn(vm, "", "_atan2", atan2BuiltIn)
	defineBuiltIn(vm, "", "encode_rgba", encodeRGBABuiltIn)
	defineBuiltIn(vm, "", "decode_rgba", decodeRGBABuiltIn)
	defineBuiltIn(vm, "", "vec2", Vec2BuiltIn)
	defineBuiltIn(vm, "", "vec3", Vec3BuiltIn)
	defineBuiltIn(vm, "", "vec4", Vec4BuiltIn)
	defineBuiltIn(vm, "", "window", builtin.WindowBuiltIn)
	defineBuiltIn(vm, "", "image", builtin.ImageBuiltIn)
	defineBuiltIn(vm, "", "texture", builtin.TextureBuiltIn)
	defineBuiltIn(vm, "", "render_texture", builtin.RenderTextureBuiltIn)
	defineBuiltIn(vm, "", "shader", builtin.ShaderBuiltIn)
	defineBuiltIn(vm, "", "camera", builtin.CameraBuiltIn)
	defineBuiltIn(vm, "", "batch", builtin.BatchBuiltIn)
	defineBuiltIn(vm, "", "batch_instanced", builtin.BatchInstancedBuiltIn)
	defineBuiltIn(vm, "", "float_array", builtin.FloatArrayBuiltin)
	defineBuiltIn(vm, "inspect", "dump_frame", dumpFrameBuiltIn)
	defineBuiltIn(vm, "inspect", "get_frame", getFrameBuiltIn)

	// os module functions
	defineBuiltIn(vm, "os", "open", openBuiltIn)
	defineBuiltIn(vm, "os", "close", closeBuiltIn)
	defineBuiltIn(vm, "os", "readln", readlnBuiltIn)
	defineBuiltIn(vm, "os", "write", writeBuiltIn)
	defineBuiltIn(vm, "os", "listdir", listdirBuiltIn)
	defineBuiltIn(vm, "os", "isdir", isdirBuiltIn)
	defineBuiltIn(vm, "os", "isfile", isfileBuiltIn)
	defineBuiltIn(vm, "os", "exists", existsBuiltIn)
	defineBuiltIn(vm, "os", "mkdir", mkdirBuiltIn)
	defineBuiltIn(vm, "os", "rmdir", rmdirBuiltIn)
	defineBuiltIn(vm, "os", "remove", removeBuiltIn)
	defineBuiltIn(vm, "os", "getcwd", getcwdBuiltIn)
	defineBuiltIn(vm, "os", "chdir", chdirBuiltIn)
	defineBuiltIn(vm, "os", "join", joinBuiltIn)
	defineBuiltIn(vm, "os", "dirname", dirnameBuiltIn)
	defineBuiltIn(vm, "os", "basename", basenameBuiltIn)
	defineBuiltIn(vm, "os", "splitext", splitextBuiltIn)

	// Color utility functions (colour_utils module)
	defineBuiltIn(vm, "colour_utils", "fade", colourUtilsFadeBuiltIn)
	defineBuiltIn(vm, "colour_utils", "tint", colourUtilsTintBuiltIn)
	defineBuiltIn(vm, "colour_utils", "brightness", colourUtilsBrightnessBuiltIn)
	defineBuiltIn(vm, "colour_utils", "lerp", colourUtilsLerpBuiltIn)
	defineBuiltIn(vm, "colour_utils", "hsv_to_rgb", colourUtilsHSVToRGBBuiltIn)
	defineBuiltIn(vm, "colour_utils", "random", colourUtilsRandomBuiltIn)

	// lox built ins e.g Exception classes
	loadBuiltInFromSource(vm, exceptionSource, "exception")

	// Do NOT inject sys into the global environment here.
	// It must be imported by client code to be available.
}

// Color utility builtin functions for colour_utils module

// fade(r, g, b, alpha) - Apply alpha to RGB values
func colourUtilsFadeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// tint(r1, g1, b1, r2, g2, b2) - Tint a color with another color
func colourUtilsTintBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// brightness(r, g, b, factor) - Adjust brightness of RGB values
func colourUtilsBrightnessBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// lerp(r1, g1, b1, r2, g2, b2, amount) - Linear interpolation between two colors
func colourUtilsLerpBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// hsv_to_rgb(h, s, v) - Convert HSV to RGB color
func colourUtilsHSVToRGBBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// random() - Generate a random color
func colourUtilsRandomBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func dumpFrameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	frame := vm.Frame()

	ip := frame.Ip
	funcName := frame.Closure.Function.Name.Get()
	if funcName == "" {
		funcName = "<script>"
	}
	fmt.Println("=====================================================")
	fmt.Printf("Frame: %s (ip=%d)\n", funcName, ip)
	fmt.Printf("Stack: \n%s\n", vm.ShowStack())
	fmt.Printf("Globals: %s\n", debug.ShowGlobals(vm.GetGlobals()))
	fmt.Println("=====================================================")

	// Optionally print upvalues, source line, etc.
	return core.NIL_VALUE
}

func getFrameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	return debug.FrameDictValue(vm)
}

func Vec2BuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("vec2 expects 2 arguments (x,y)")
		return core.NIL_VALUE
	}
	xVal := vm.Stack(arg_stackptr)
	yVal := vm.Stack(arg_stackptr + 1)

	if !xVal.IsNumber() || !yVal.IsNumber() {
		vm.RunTimeError("vec2 arguments must be numbers")
		return core.NIL_VALUE
	}

	return core.MakeVec2Value(float64(xVal.AsFloat()), float64(yVal.AsFloat()), false)
}

func Vec3BuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 {
		vm.RunTimeError("vec3 expects 3 arguments (x,y,z)")
		return core.NIL_VALUE
	}
	xVal := vm.Stack(arg_stackptr)
	yVal := vm.Stack(arg_stackptr + 1)
	zVal := vm.Stack(arg_stackptr + 2)

	if !xVal.IsNumber() || !yVal.IsNumber() || !zVal.IsNumber() {
		vm.RunTimeError("vec3 arguments must be numbers")
		return core.NIL_VALUE
	}

	return core.MakeVec3Value(float64(xVal.AsFloat()), float64(yVal.AsFloat()), float64(zVal.AsFloat()), false)
}

func Vec4BuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 4 {
		vm.RunTimeError("vec4 expects 4 arguments (x,y,z,w)")
		return core.NIL_VALUE
	}
	xVal := vm.Stack(arg_stackptr)
	yVal := vm.Stack(arg_stackptr + 1)
	zVal := vm.Stack(arg_stackptr + 2)
	wVal := vm.Stack(arg_stackptr + 3)

	if !xVal.IsNumber() || !yVal.IsNumber() || !zVal.IsNumber() || !wVal.IsNumber() {
		vm.RunTimeError("vec4 arguments must be numbers")
		return core.NIL_VALUE
	}

	return core.MakeVec4Value(float64(xVal.AsFloat()), float64(yVal.AsFloat()), float64(zVal.AsFloat()), float64(wVal.AsFloat()), false)
}

func rangeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount < 1 || argCount > 3 {
		vm.RunTimeError("range expects 1 to 3 arguments")
		return core.NIL_VALUE
	}

	start := 0
	end := 0
	step := 1

	switch argCount {
	case 1:
		end = vm.Stack(arg_stackptr).AsInt()
	case 2:
		start = vm.Stack(arg_stackptr).AsInt()
		end = vm.Stack(arg_stackptr + 1).AsInt()
	case 3:
		start = vm.Stack(arg_stackptr).AsInt()
		end = vm.Stack(arg_stackptr + 1).AsInt()
		step = vm.Stack(arg_stackptr + 2).AsInt()
	}

	if step == 0 {
		vm.RunTimeError("step cannot be zero")
		return core.NIL_VALUE
	}

	iter := core.MakeIntIteratorObject(start, end, step)
	return core.MakeObjectValue(iter, false)
}

func sleepBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("sleep expects 1 argument")
		return core.NIL_VALUE
	}
	tVal := vm.Stack(arg_stackptr)
	if !tVal.IsNumber() {
		vm.RunTimeError("sleep argument must be number")
		return core.NIL_VALUE
	}
	var dur time.Duration
	if tVal.IsInt() {
		dur = time.Duration(tVal.AsInt()) * time.Second
	}
	if tVal.IsFloat() {
		dur = time.Duration(tVal.AsFloat()) * time.Second
	}
	time.Sleep(dur)
	return core.NIL_VALUE
}

func encodeRGBABuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 3 {
		vm.RunTimeError("encode_rgb expects 3 arguments")
		return core.NIL_VALUE
	}
	rVal := vm.Stack(arg_stackptr)
	gVal := vm.Stack(arg_stackptr + 1)
	bVal := vm.Stack(arg_stackptr + 2)
	if !rVal.IsInt() || !gVal.IsInt() || !bVal.IsInt() {
		vm.RunTimeError("encode_rgb arguments must be integers")
		return core.NIL_VALUE
	}
	r := rVal.Int
	g := gVal.Int
	b := bVal.Int
	color := util.EncodeRGB(r, g, b)
	return core.MakeFloatValue(color, false)
}

func decodeRGBABuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("decode_rgb expects 1 float argument")
		return core.NIL_VALUE
	}
	fVal := vm.Stack(arg_stackptr)

	if !fVal.IsFloat() {
		vm.RunTimeError("decode_rgb argument must be a float")
		return core.NIL_VALUE
	}
	f := fVal.Float
	r, g, b := util.DecodeRGB(f)
	rVal := core.MakeIntValue(int(r), false)
	gVal := core.MakeIntValue(int(g), false)
	bVal := core.MakeIntValue(int(b), false)
	ro := core.MakeListObject([]core.Value{rVal, gVal, bVal}, true)
	return core.MakeObjectValue(ro, false)
}

func typeName(val core.Value) string {
	var val_type string
	switch val.Type {
	case core.VAL_INT:
		val_type = "int"
	case core.VAL_FLOAT:
		val_type = "float"
	case core.VAL_BOOL:
		val_type = "boolean"
	case core.VAL_OBJ:
		switch val.Obj.GetType() {
		case core.OBJECT_STRING:
			val_type = "string"
		case core.OBJECT_FUNCTION:
			val_type = "function"
		case core.OBJECT_CLOSURE:
			val_type = "closure"
		case core.OBJECT_LIST:
			val_type = "list"
		case core.OBJECT_NATIVE:
			val_type = "builtin"
		case core.OBJECT_DICT:
			val_type = "dict"
		case core.OBJECT_CLASS:
			val_type = "class"
		case core.OBJECT_INSTANCE:
			val_type = "instance"
		case core.OBJECT_MODULE:
			val_type = "module"
		case core.OBJECT_FILE:
			val_type = "file"

		}
	case core.VAL_NIL:
		val_type = "nil"
	}
	return val_type
}

func typeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.NIL_VALUE
	}
	val := vm.Stack(arg_stackptr)
	name := typeName(val)

	return core.MakeStringObjectValue(name, true)
}

func argsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	argvList := []core.Value{}
	for _, a := range vm.Args() {
		argvList = append(argvList, core.MakeStringObjectValue(a, true))
	}
	list := core.MakeListObject(argvList, false)
	return core.MakeObjectValue(list, false)
}

func floatBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.NIL_VALUE
	}
	arg := vm.Stack(arg_stackptr)

	switch arg.Type {
	case core.VAL_FLOAT:
		return arg
	case core.VAL_INT:
		return core.MakeFloatValue(float64(arg.Int), false)
	case core.VAL_OBJ:
		if arg.Obj.GetType() == core.OBJECT_STRING {
			f, ok := arg.AsString().ParseFloat()
			if !ok {
				vm.RunTimeError("Could not parse string into float.")
				return core.NIL_VALUE
			}
			return core.MakeFloatValue(f, false)
		}
	}
	vm.RunTimeError("Argument must be number or valid string")
	return core.NIL_VALUE
}

func intBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.NIL_VALUE
	}
	arg := vm.Stack(arg_stackptr)

	switch arg.Type {
	case core.VAL_INT:
		return arg
	case core.VAL_FLOAT:
		return core.MakeIntValue(int(arg.Float), false)
	case core.VAL_OBJ:
		if arg.Obj.GetType() == core.OBJECT_STRING {
			i, ok := arg.AsString().ParseInt()
			if !ok {
				vm.RunTimeError("Could not parse string into int.")
				return core.NIL_VALUE
			}
			return core.MakeIntValue(i, false)
		}
	}
	vm.RunTimeError("Argument must be number or valid string.")
	return core.NIL_VALUE
}

func clockBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	elapsed := time.Since(vm.StartTime())
	return core.MakeFloatValue(float64(elapsed.Seconds()), false)
}

func randBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	return core.MakeFloatValue(rand.Float64(), false)
}

// len( string )
func lenBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to len.")
		return core.NIL_VALUE
	}
	val := vm.Stack(arg_stackptr)
	if val.Type != core.VAL_OBJ {
		vm.RunTimeError("Invalid argument type to len.")
		return core.NIL_VALUE
	}
	switch val.Obj.GetType() {
	case core.OBJECT_STRING:
		s := val.AsString().Get()
		return core.MakeIntValue(len(s), false)
	case core.OBJECT_LIST:
		l := val.AsList().Get()
		return core.MakeIntValue(len(l), false)
	}
	vm.RunTimeError("Invalid argument type to len.")
	return core.NIL_VALUE
}

// sin(number)
func powBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to pow.")
		return core.NIL_VALUE
	}
	vbase := vm.Stack(arg_stackptr)
	vexp := vm.Stack(arg_stackptr + 1)

	if vbase.Type != core.VAL_FLOAT || vexp.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to pow.")
		return core.NIL_VALUE
	}
	n := vbase.Float
	return core.MakeFloatValue(math.Pow(n, vexp.Float), false)
}

// sin(number)
func sinBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sin.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to sin.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to cos.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to cos.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Cos(n), false)
}

// cos(number)
func tanBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to tan.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to tan.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Tan(n), false)
}

func sqrtBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sqrt.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to sqrt.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sqrt(n), false)
}

func atan2BuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to atan2.")
		return core.NIL_VALUE
	}
	vnum1 := vm.Stack(arg_stackptr)
	vnum2 := vm.Stack(arg_stackptr + 1)

	if vnum1.Type != core.VAL_FLOAT || vnum2.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to atan2.")
		return core.NIL_VALUE
	}
	n1 := vnum1.Float
	n2 := vnum2.Float
	return core.MakeFloatValue(math.Atan2(n1, n2), false)
}

// append(obj,value)
func appendBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to append.")
		return core.NIL_VALUE
	}
	val := vm.Stack(arg_stackptr)
	if val.Type != core.VAL_OBJ {
		vm.RunTimeError("Argument 1 to append must be list.")
		return core.NIL_VALUE
	}
	val2 := vm.Stack(arg_stackptr + 1)
	switch val.Obj.GetType() {

	case core.OBJECT_LIST:
		l := val.AsList()
		if l.Tuple {
			vm.RunTimeError("Tuples are immutable")
			return core.NIL_VALUE
		}
		l.Append(val2)
		return core.MakeObjectValue(l, false)
	}
	vm.RunTimeError("Argument 1 to append must be list.")
	return core.NIL_VALUE
}

// replace( string|list )
func replaceBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 3 {
		vm.RunTimeError("Invalid argument count to replace.")
		return core.NIL_VALUE
	}
	target := vm.Stack(arg_stackptr)
	from := vm.Stack(arg_stackptr + 1)
	to := vm.Stack(arg_stackptr + 2)

	if target.Type != core.VAL_OBJ || target.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to replace.")
		return core.NIL_VALUE
	}

	s := target.AsString()
	return s.Replace(from, to)
}

// format(template, ...args) - Printf-style formatting
func formatBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount < 1 {
		vm.RunTimeError("format expects at least 1 argument")
		return core.NIL_VALUE
	}

	templateVal := vm.Stack(arg_stackptr)
	if templateVal.Type != core.VAL_OBJ || templateVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("format template must be a string")
		return core.NIL_VALUE
	}

	template := templateVal.AsString().Get()

	// Convert Lox values to Go interfaces for fmt.Sprintf
	goArgs := make([]interface{}, argCount-1)
	for i := 1; i < argCount; i++ {
		arg := vm.Stack(arg_stackptr + i)
		goArgs[i-1] = valueToGoInterface(arg)
	}

	result := fmt.Sprintf(template, goArgs...)
	return core.MakeStringObjectValue(result, true)
}

// Helper function to convert Lox value to Go interface for fmt.Sprintf
func valueToGoInterface(val core.Value) interface{} {
	switch val.Type {
	case core.VAL_INT:
		return val.Int
	case core.VAL_FLOAT:
		return val.Float
	case core.VAL_BOOL:
		return val.Bool
	case core.VAL_NIL:
		return nil
	case core.VAL_OBJ:
		if val.Obj.GetType() == core.OBJECT_STRING {
			return val.AsString().Get()
		}
		return fmt.Sprintf("%v", val.Obj)
	}
	return fmt.Sprintf("%v", val)
}

// return a FileObject
func openBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to open.")
		return core.NIL_VALUE
	}
	path := vm.Stack(arg_stackptr)
	mode := vm.Stack(arg_stackptr + 1)

	if path.Type != core.VAL_OBJ || path.Obj.GetType() != core.OBJECT_STRING ||
		mode.Type != core.VAL_OBJ || mode.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to open.")
		return core.NIL_VALUE
	}

	s_path := path.AsString().Get()
	s_mode := mode.AsString().Get()
	fp, err := openFile(s_path, s_mode)
	if err != nil {
		vm.RunTimeError("%v", err)
		return core.NIL_VALUE
	}
	file := core.MakeObjectValue(core.MakeFileObject(fp), true)
	return file

}

func closeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to close.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to close.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	fo.Close()
	return core.MakeBooleanValue(true, false)
}

func readlnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to readln.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to readln.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("readln attempted on closed file.")
		return core.NIL_VALUE
	}

	rv := fo.ReadLine()
	if rv.Type == core.VAL_NIL {
		vm.RaiseExceptionByName("EOFError", "End of file reached")
		return core.MakeBooleanValue(true, false)
	}
	return rv
}

func writeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to writeln.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)
	str := vm.Stack(arg_stackptr + 1)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.NIL_VALUE
	}
	if str.Type != core.VAL_OBJ || str.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("writeln attempted on closed file.")
		return core.NIL_VALUE
	}

	fo.Write(str)
	return core.MakeBooleanValue(true, false)
}

func openFile(path string, mode string) (*os.File, error) {
	switch mode {
	case "r":
		return os.Open(path) // Read-only
	case "w":
		return os.Create(path) // Write (truncate if exists)
	case "a":
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Append
	default:
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}
}

func makeBuiltInModule(vm *VM, moduleName string) {

	env := core.NewEnvironment(moduleName)
	module := core.MakeModuleObject(moduleName, *env)
	vm.builtInModules[core.InternName(moduleName)] = module
	core.LogFmtLn(core.INFO, "Created built-in module %s", moduleName)

}

func addBuiltInModuleFunction(vm *VM, moduleName string, name string, fn core.BuiltInFn) {
	// Add a function to a built-in module
	module := vm.builtInModules[core.InternName(moduleName)]
	fo := core.MakeBuiltInObject(fn)
	module.Environment.Vars[core.InternName(name)] = core.MakeObjectValue(fo, false)
}

// load built-in functions from source code
func loadBuiltInFromSource(vm *VM, source string, moduleName string) {
	core.Log(core.INFO, "Loading built-in module ")
	subvm := NewVM("", false)
	//	DebugSuppress = true
	_, _ = subvm.Interpret(source, moduleName)
	for k, v := range subvm.frames[0].Closure.Function.Environment.Vars {
		vm.builtIns[k] = v
	}

	core.DebugSuppress = false
}

// predefine an Exception class using Lox source
const exceptionSource = `class Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "Exception";
	}
	toString() {
	    return this.msg;
	}
}
class EOFError < Exception {
     init(msg) {
	    this.msg = msg;
		this.name = "EOFError";
	}
	toString() {
	    return this.msg;
	}
}
class RunTimeError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "RunTimeError";
	}
	toString() {
		return this.msg;
	}
}
`

// ============================================================================
// OS Module Functions - Directory and File Operations
// ============================================================================

// listdir(path) -> list - List directory contents
func listdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to listdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to listdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	entries, err := os.ReadDir(path)
	if err != nil {
		vm.RunTimeError("Failed to read directory '%s': %v", path, err)
		return core.NIL_VALUE
	}
	// Create a list of filenames
	var items []core.Value
	for _, entry := range entries {
		filename := core.MakeStringObjectValue(entry.Name(), false)
		items = append(items, filename)
	}
	list := core.MakeListObject(items, false)

	return core.MakeObjectValue(list, false)
}

// isdir(path) -> bool - Check if path is a directory
func isdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to isdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to isdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	info, err := os.Stat(path)
	if err != nil {
		return core.MakeBooleanValue(false, false)
	}

	return core.MakeBooleanValue(info.IsDir(), false)
}

// isfile(path) -> bool - Check if path is a regular file
func isfileBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to isfile.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to isfile, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	info, err := os.Stat(path)
	if err != nil {
		return core.MakeBooleanValue(false, false)
	}

	return core.MakeBooleanValue(!info.IsDir(), false)
}

// exists(path) -> bool - Check if path exists
func existsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to exists.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to exists, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	_, err := os.Stat(path)
	return core.MakeBooleanValue(err == nil, false)
}

// mkdir(path) -> bool - Create directory
func mkdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to mkdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to mkdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		vm.RunTimeError("Failed to create directory '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

// rmdir(path) -> bool - Remove empty directory
func rmdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to rmdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to rmdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Remove(path)
	if err != nil {
		vm.RunTimeError("Failed to remove directory '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

// remove(path) -> bool - Remove file
func removeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to remove.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to remove, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Remove(path)
	if err != nil {
		vm.RunTimeError("Failed to remove file '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

// getcwd() -> string - Get current working directory
func getcwdBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("Invalid argument count to getcwd.")
		return core.NIL_VALUE
	}

	cwd, err := os.Getwd()
	if err != nil {
		vm.RunTimeError("Failed to get current directory: %v", err)
		return core.NIL_VALUE
	}

	return core.MakeStringObjectValue(cwd, false)
}

// chdir(path) -> bool - Change current working directory
func chdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to chdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to chdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Chdir(path)
	if err != nil {
		vm.RunTimeError("Failed to change directory to '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

// ============================================================================
// Path Manipulation Functions
// ============================================================================

// join(paths...) -> string - Join path components
func joinBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount < 1 {
		vm.RunTimeError("join requires at least one argument.")
		return core.NIL_VALUE
	}

	var paths []string
	for i := 0; i < argCount; i++ {
		pathVal := vm.Stack(arg_stackptr + i)
		if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
			vm.RunTimeError("Invalid argument type to join, expected string.")
			return core.NIL_VALUE
		}
		paths = append(paths, pathVal.AsString().Get())
	}

	result := paths[0]
	for i := 1; i < len(paths); i++ {
		if strings.HasSuffix(result, "/") || strings.HasSuffix(result, "\\") {
			result = result + paths[i]
		} else {
			result = result + "/" + paths[i]
		}
	}

	return core.MakeStringObjectValue(result, false)
}

// dirname(path) -> string - Get directory part of path
func dirnameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to dirname.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to dirname, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Simple dirname implementation
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return core.MakeStringObjectValue(".", false)
	}
	if lastSlash == 0 {
		return core.MakeStringObjectValue("/", false)
	}

	return core.MakeStringObjectValue(path[:lastSlash], false)
}

// basename(path) -> string - Get basename part of path
func basenameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to basename.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to basename, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Simple basename implementation
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return core.MakeStringObjectValue(path, false)
	}

	return core.MakeStringObjectValue(path[lastSlash+1:], false)
}

// splitext(path) -> [name, ext] - Split filename and extension
func splitextBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to splitext.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to splitext, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Find the last dot
	lastDot := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			lastDot = i
			break
		}
		if path[i] == '/' || path[i] == '\\' {
			break // Stop at directory separator
		}
	}

	var name, ext string
	if lastDot == -1 {
		name = path
		ext = ""
	} else {
		name = path[:lastDot]
		ext = path[lastDot:]
	}

	nameVal := core.MakeStringObjectValue(name, false)
	extVal := core.MakeStringObjectValue(ext, false)
	items := []core.Value{nameVal, extVal}
	list := core.MakeListObject(items, false)

	return core.MakeObjectValue(list, false)
}
