package lox

import (
	"fmt"
	"glox/src/builtin"
	"glox/src/core"
	"glox/src/util"
	"math"
	"math/rand"
	"os"
	"time"
)

func (vm *VM) defineBuiltIns() {

	core.Log(core.INFO, "Defining built-in functions")
	// native functions
	vm.defineBuiltIn("args", argsBuiltIn)
	vm.defineBuiltIn("clock", clockBuiltIn)
	vm.defineBuiltIn("type", typeBuiltIn)
	vm.defineBuiltIn("len", lenBuiltIn)
	vm.defineBuiltIn("sin", sinBuiltIn)
	vm.defineBuiltIn("cos", cosBuiltIn)
	vm.defineBuiltIn("sqrt", sqrtBuiltIn)
	vm.defineBuiltIn("append", appendBuiltIn)
	vm.defineBuiltIn("float", floatBuiltIn)
	vm.defineBuiltIn("int", intBuiltIn)
	vm.defineBuiltIn("lox_mandel_array", builtin.MandelArrayBuiltIn)
	vm.defineBuiltIn("replace", replaceBuiltIn)
	vm.defineBuiltIn("open", openBuiltIn)
	vm.defineBuiltIn("close", closeBuiltIn)
	vm.defineBuiltIn("readln", readlnBuiltIn)
	vm.defineBuiltIn("write", writeBuiltIn)
	vm.defineBuiltIn("rand", randBuiltIn)
	vm.defineBuiltIn("draw_png", builtin.DrawPNGBuiltIn)
	vm.defineBuiltIn("float_array", builtin.FloatArrayBuiltin)
	vm.defineBuiltIn("encode_rgb", encodeRGBABuiltIn)
	vm.defineBuiltIn("decode_rgb", decodeRGBABuiltIn)
	vm.defineBuiltIn("window", builtin.WindowBuiltIn)
	vm.defineBuiltIn("texture", builtin.TextureBuiltIn)
	vm.defineBuiltIn("render_texture", builtin.RenderTextureBuiltIn)
	vm.defineBuiltIn("image", builtin.ImageBuiltIn)
	vm.defineBuiltIn("sleep", sleepBuiltIn)
	vm.defineBuiltIn("range", rangeBuiltIn)
	vm.defineBuiltIn("vec2", Vec2BuiltIn)
	vm.defineBuiltIn("vec3", Vec3BuiltIn)
	vm.defineBuiltIn("vec4", Vec4BuiltIn)

	// lox built ins e.g Exception classes
	vm.loadBuiltInModule(exceptionSource, "exception")

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

	var values []core.Value
	for i := start; i < end; i += step {
		values = append(values, core.MakeIntValue(i, false))
	}

	list := core.MakeListObject(values, false)
	return core.MakeObjectValue(list, false)
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

func (vm *VM) loadBuiltInModule(source string, moduleName string) {
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
