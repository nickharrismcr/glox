package lox

import (
	"fmt"
	"glox/src/core"
	debug "glox/src/loxdebug"
	"glox/src/util"
	"math"
	"math/rand"
	"os"
	"time"
)

func (vm *VM) defineBuiltIns() {

	debug.Debug("Defining built-in functions")
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
	vm.defineBuiltIn("lox_mandel_array", MandelArrayBuiltIn)
	vm.defineBuiltIn("replace", replaceBuiltIn)
	vm.defineBuiltIn("open", openBuiltIn)
	vm.defineBuiltIn("close", closeBuiltIn)
	vm.defineBuiltIn("readln", readlnBuiltIn)
	vm.defineBuiltIn("write", writeBuiltIn)
	vm.defineBuiltIn("rand", randBuiltIn)
	vm.defineBuiltIn("draw_png", drawPNGBuiltIn)
	vm.defineBuiltIn("float_array", MakeFloatArrayBuiltIn)
	vm.defineBuiltIn("encode_rgb", encodeRGBABuiltIn)
	vm.defineBuiltIn("decode_rgb", decodeRGBABuiltIn)
	vm.defineBuiltIn("window", graphicsBuiltIn)
	vm.defineBuiltIn("sleep", sleepBuiltIn)

	// lox built ins e.g Exception classes
	vm.loadBuiltInModule(exceptionSource, "exception")

}

func graphicsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("graphics expects 2 arguments")
		return core.MakeNilValue()
	}
	wVal := vm.Stack(arg_stackptr)
	hVal := vm.Stack(arg_stackptr + 1)
	if !wVal.IsInt() || !hVal.IsInt() {
		vm.RunTimeError("graphics arguments must be integers")
		return core.MakeNilValue()
	}
	o := core.MakeGraphicsObject(wVal.Int, hVal.Int)
	return core.MakeObjectValue(o, true)
}

func sleepBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("sleep expects 1 argument")
		return core.MakeNilValue()
	}
	tVal := vm.Stack(arg_stackptr)
	if !tVal.IsNumber() {
		vm.RunTimeError("sleep argument must be number")
		return core.MakeNilValue()
	}
	var dur time.Duration
	if tVal.IsInt() {
		dur = time.Duration(tVal.AsInt()) * time.Second
	}
	if tVal.IsFloat() {
		dur = time.Duration(tVal.AsFloat()) * time.Second
	}
	time.Sleep(dur)
	return core.MakeNilValue()
}

func encodeRGBABuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 3 {
		vm.RunTimeError("encode_rgb expects 3 arguments")
		return core.MakeNilValue()
	}
	rVal := vm.Stack(arg_stackptr)
	gVal := vm.Stack(arg_stackptr + 1)
	bVal := vm.Stack(arg_stackptr + 2)
	if !rVal.IsInt() || !gVal.IsInt() || !bVal.IsInt() {
		vm.RunTimeError("encode_rgb arguments must be integers")
		return core.MakeNilValue()
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
		return core.MakeNilValue()
	}
	fVal := vm.Stack(arg_stackptr)

	if !fVal.IsFloat() {
		vm.RunTimeError("decode_rgb argument must be a float")
		return core.MakeNilValue()
	}
	f := fVal.Float
	r, g, b := util.DecodeRGB(f)
	rVal := core.MakeIntValue(int(r), false)
	gVal := core.MakeIntValue(int(g), false)
	bVal := core.MakeIntValue(int(b), false)
	ro := core.MakeListObject([]core.Value{rVal, gVal, bVal}, true)
	return core.MakeObjectValue(ro, false)
}

func MakeFloatArrayBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	widthval := vm.Stack(arg_stackptr)
	heightval := vm.Stack(arg_stackptr + 1)
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to float_array.")
		return core.MakeNilValue()
	}
	if !widthval.IsInt() || !heightval.IsInt() {
		vm.RunTimeError("float_array arguments must be integers")
		return core.MakeNilValue()
	}
	width := widthval.Int
	height := heightval.Int
	floatArrObj := core.MakeFloatArrayObject(width, height)
	return core.MakeObjectValue(floatArrObj, false)
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
		case core.OBJECT_GRAPHICS:
			val_type = "graphics"
		case core.OBJECT_FLOAT_ARRAY:
			val_type = "float_array"
		}
	case core.VAL_NIL:
		val_type = "nil"
	}
	return val_type
}

func typeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.MakeNilValue()
	}
	val := vm.Stack(arg_stackptr)
	name := typeName(val)

	return core.MakeObjectValue(core.MakeStringObject(name), true)
}

func argsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	argvList := []core.Value{}
	for _, a := range vm.Args() {
		argvList = append(argvList, core.MakeObjectValue(core.MakeStringObject(a), true))
	}
	list := core.MakeListObject(argvList, false)
	return core.MakeObjectValue(list, false)
}

func floatBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.MakeNilValue()
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
				return core.MakeNilValue()
			}
			return core.MakeFloatValue(f, false)
		}
	}
	vm.RunTimeError("Argument must be number or valid string")
	return core.MakeNilValue()
}

func intBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.MakeNilValue()
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
				return core.MakeNilValue()
			}
			return core.MakeIntValue(i, false)
		}
	}
	vm.RunTimeError("Argument must be number or valid string.")
	return core.MakeNilValue()
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
		return core.MakeNilValue()
	}
	val := vm.Stack(arg_stackptr)
	if val.Type != core.VAL_OBJ {
		vm.RunTimeError("Invalid argument type to len.")
		return core.MakeNilValue()
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
	return core.MakeNilValue()
}

// sin(number)
func sinBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sin.")
		return core.MakeNilValue()
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to sin.")
		return core.MakeNilValue()
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to cos.")
		return core.MakeNilValue()
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to cos.")
		return core.MakeNilValue()
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Cos(n), false)
}

func sqrtBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sqrt.")
		return core.MakeNilValue()
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {

		vm.RunTimeError("Invalid argument type to sqrt.")
		return core.MakeNilValue()
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sqrt(n), false)
}

// append(obj,value)
func appendBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to append.")
		return core.MakeNilValue()
	}
	val := vm.Stack(arg_stackptr)
	if val.Type != core.VAL_OBJ {
		vm.RunTimeError("Argument 1 to append must be list.")
		return core.MakeNilValue()
	}
	val2 := vm.Stack(arg_stackptr + 1)
	switch val.Obj.GetType() {

	case core.OBJECT_LIST:
		l := val.AsList()
		if l.Tuple {
			vm.RunTimeError("Tuples are immutable")
			return core.MakeNilValue()
		}
		l.Append(val2)
		return core.MakeObjectValue(l, false)
	}
	vm.RunTimeError("Argument 1 to append must be list.")
	return core.MakeNilValue()
}

// replace( string|list )
func replaceBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 3 {
		vm.RunTimeError("Invalid argument count to replace.")
		return core.MakeNilValue()
	}
	target := vm.Stack(arg_stackptr)
	from := vm.Stack(arg_stackptr + 1)
	to := vm.Stack(arg_stackptr + 2)

	if target.Type != core.VAL_OBJ || target.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to replace.")
		return core.MakeNilValue()
	}

	s := target.AsString()
	return s.Replace(from, to)
}

// return a FileObject
func openBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to open.")
		return core.MakeNilValue()
	}
	path := vm.Stack(arg_stackptr)
	mode := vm.Stack(arg_stackptr + 1)

	if path.Type != core.VAL_OBJ || path.Obj.GetType() != core.OBJECT_STRING ||
		mode.Type != core.VAL_OBJ || mode.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to open.")
		return core.MakeNilValue()
	}

	s_path := path.AsString().Get()
	s_mode := mode.AsString().Get()
	fp, err := openFile(s_path, s_mode)
	if err != nil {
		vm.RunTimeError("%v", err)
		return core.MakeNilValue()
	}
	file := core.MakeObjectValue(core.MakeFileObject(fp), true)
	return file

}

func closeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to close.")
		return core.MakeNilValue()
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to close.")
		return core.MakeNilValue()
	}

	fo := fov.Obj.(*core.FileObject)
	fo.Close()
	return core.MakeBooleanValue(true, false)
}

func readlnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to readln.")
		return core.MakeNilValue()
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to readln.")
		return core.MakeNilValue()
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("readln attempted on closed file.")
		return core.MakeNilValue()
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
		return core.MakeNilValue()
	}
	fov := vm.Stack(arg_stackptr)
	str := vm.Stack(arg_stackptr + 1)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.MakeNilValue()
	}
	if str.Type != core.VAL_OBJ || str.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.MakeNilValue()
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("writeln attempted on closed file.")
		return core.MakeNilValue()
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
	debug.Debug("Loading built-in module ")
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
