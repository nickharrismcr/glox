package lox

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

func (vm *VM) defineBuiltIns() {

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
	vm.defineBuiltIn("float_array", makeFloatArrayBuiltIn)
	vm.defineBuiltIn("encode_rgb", encodeRGBABuiltIn)
	vm.defineBuiltIn("decode_rgb", decodeRGBABuiltIn)

	// lox built ins e.g Exception classes
	vm.loadBuiltInModule(exceptionSource)
	vm.loadBuiltInModule(eofErrorSource)
	vm.loadBuiltInModule(runTimeErrorSource)

}

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

func encodeRGBABuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 3 {
		vm.runTimeError("encodeRGB expects 3 arguments")
		return makeNilValue()
	}
	rVal := vm.stack[arg_stackptr]
	gVal := vm.stack[arg_stackptr+1]
	bVal := vm.stack[arg_stackptr+2]
	if !rVal.isInt() || !gVal.isInt() || !bVal.isInt() {
		vm.runTimeError("encodeRGB arguments must be integers")
		return makeNilValue()
	}
	r := rVal.Int
	g := gVal.Int
	b := bVal.Int
	color := EncodeRGB(r, g, b)
	return makeFloatValue(color, false)
}

func decodeRGBABuiltIn(argCount int, arg_stackptr int, vm *VM) Value {
	if argCount != 1 {
		vm.runTimeError("decodeRGB expects 1 float argument")
		return makeNilValue()
	}
	fVal := vm.stack[arg_stackptr]

	if !fVal.isFloat() {
		vm.runTimeError("decodeRGB argument must be a float")
		return makeNilValue()
	}
	f := fVal.Float
	r, g, b := DecodeRGB(f)
	rVal := makeIntValue(int(r), false)
	gVal := makeIntValue(int(g), false)
	bVal := makeIntValue(int(b), false)
	ro := makeListObject([]Value{rVal, gVal, bVal}, true)
	return makeObjectValue(ro, false)
}

func makeFloatArrayBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	widthval := vm.stack[arg_stackptr]
	heightval := vm.stack[arg_stackptr+1]
	if argCount != 2 {
		vm.runTimeError("Invalid argument count to float_array.")
		return makeNilValue()
	}
	if !widthval.isInt() || !heightval.isInt() {
		vm.runTimeError("float_array arguments must be integers")
		return makeNilValue()
	}
	width := widthval.Int
	height := heightval.Int
	floatArrObj := makeFloatArrayObject(width, height)
	return makeObjectValue(floatArrObj, false)
}

func typeName(val Value) string {
	var val_type string
	switch val.Type {
	case VAL_INT:
		val_type = "int"
	case VAL_FLOAT:
		val_type = "float"
	case VAL_BOOL:
		val_type = "boolean"
	case VAL_OBJ:
		switch val.Obj.getType() {
		case OBJECT_STRING:
			val_type = "string"
		case OBJECT_FUNCTION:
			val_type = "function"
		case OBJECT_CLOSURE:
			val_type = "closure"
		case OBJECT_LIST:
			val_type = "list"
		case OBJECT_NATIVE:
			val_type = "builtin"
		case OBJECT_DICT:
			val_type = "dict"
		case OBJECT_CLASS:
			val_type = "class"
		case OBJECT_INSTANCE:
			val_type = "instance"
		case OBJECT_MODULE:
			val_type = "module"
		case OBJECT_FILE:
			val_type = "file"
		}
	case VAL_NIL:
		val_type = "nil"
	}
	return val_type
}

func typeBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	name := typeName(val)

	return makeObjectValue(makeStringObject(name), true)
}

func argsBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	argvList := []Value{}
	for _, a := range vm.args {
		argvList = append(argvList, makeObjectValue(makeStringObject(a), true))
	}
	list := makeListObject(argvList, false)
	return makeObjectValue(list, false)
}

func floatBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.Type {
	case VAL_FLOAT:
		return arg
	case VAL_INT:
		return makeFloatValue(float64(arg.Int), false)
	case VAL_OBJ:
		if arg.Obj.getType() == OBJECT_STRING {
			f, ok := arg.asString().parseFloat()
			if !ok {
				vm.runTimeError("Could not parse string into float.")
				return makeNilValue()
			}
			return makeFloatValue(f, false)
		}
	}
	vm.runTimeError("Argument must be number or valid string")
	return makeNilValue()
}

func intBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.Type {
	case VAL_INT:
		return arg
	case VAL_FLOAT:
		return makeIntValue(int(arg.Float), false)
	case VAL_OBJ:
		if arg.Obj.getType() == OBJECT_STRING {
			i, ok := arg.asString().parseInt()
			if !ok {
				vm.runTimeError("Could not parse string into int.")
				return makeNilValue()
			}
			return makeIntValue(i, false)
		}
	}
	vm.runTimeError("Argument must be number or valid string.")
	return makeNilValue()
}

func clockBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeFloatValue(float64(elapsed.Seconds()), false)
}

func randBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	return makeFloatValue(rand.Float64(), false)
}

// len( string )
func lenBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to len.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	if val.Type != VAL_OBJ {
		vm.runTimeError("Invalid argument type to len.")
		return makeNilValue()
	}
	switch val.Obj.getType() {
	case OBJECT_STRING:
		s := val.asString().get()
		return makeIntValue(len(s), false)
	case OBJECT_LIST:
		l := val.asList().get()
		return makeIntValue(len(l), false)
	}
	vm.runTimeError("Invalid argument type to len.")
	return makeNilValue()
}

// sin(number)
func sinBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to sin.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	if vnum.Type != VAL_FLOAT {
		vm.runTimeError("Invalid argument type to sin.")
		return makeNilValue()
	}
	n := vnum.Float
	return makeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to cos.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	if vnum.Type != VAL_FLOAT {

		vm.runTimeError("Invalid argument type to cos.")
		return makeNilValue()
	}
	n := vnum.Float
	return makeFloatValue(math.Cos(n), false)
}

func sqrtBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to sqrt.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	if vnum.Type != VAL_FLOAT {

		vm.runTimeError("Invalid argument type to sqrt.")
		return makeNilValue()
	}
	n := vnum.Float
	return makeFloatValue(math.Sqrt(n), false)
}

// append(obj,value)
func appendBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to append.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	if val.Type != VAL_OBJ {
		vm.runTimeError("Argument 1 to append must be list.")
		return makeNilValue()
	}
	val2 := vm.stack[arg_stackptr+1]
	switch val.Obj.getType() {

	case OBJECT_LIST:
		l := val.asList()
		if l.tuple {
			vm.runTimeError("Tuples are immutable")
			return makeNilValue()
		}
		l.append(val2)
		return makeObjectValue(l, false)
	}
	vm.runTimeError("Argument 1 to append must be list.")
	return makeNilValue()
}

// replace( string|list )
func replaceBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 3 {
		vm.runTimeError("Invalid argument count to replace.")
		return makeNilValue()
	}
	target := vm.stack[arg_stackptr]
	from := vm.stack[arg_stackptr+1]
	to := vm.stack[arg_stackptr+2]

	if target.Type != VAL_OBJ || target.Obj.getType() != OBJECT_STRING {
		vm.runTimeError("Invalid argument type to replace.")
		return makeNilValue()
	}

	s := target.asString()
	return s.replace(from, to)
}

// return a FileObject
func openBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to open.")
		return makeNilValue()
	}
	path := vm.stack[arg_stackptr]
	mode := vm.stack[arg_stackptr+1]

	if path.Type != VAL_OBJ || path.Obj.getType() != OBJECT_STRING ||
		mode.Type != VAL_OBJ || mode.Obj.getType() != OBJECT_STRING {
		vm.runTimeError("Invalid argument type to open.")
		return makeNilValue()
	}

	s_path := path.asString().get()
	s_mode := mode.asString().get()
	fp, err := openFile(s_path, s_mode)
	if err != nil {
		vm.runTimeError("%v", err)
		return makeNilValue()
	}
	file := makeObjectValue(makeFileObject(fp), true)
	return file

}

func closeBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to close.")
		return makeNilValue()
	}
	fov := vm.stack[arg_stackptr]

	if fov.Type != VAL_OBJ || fov.Obj.getType() != OBJECT_FILE {
		vm.runTimeError("Invalid argument type to close.")
		return makeNilValue()
	}

	fo := fov.Obj.(*FileObject)
	fo.close()
	return makeBooleanValue(true, false)
}

func readlnBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to readln.")
		return makeNilValue()
	}
	fov := vm.stack[arg_stackptr]

	if fov.Type != VAL_OBJ || fov.Obj.getType() != OBJECT_FILE {
		vm.runTimeError("Invalid argument type to readln.")
		return makeNilValue()
	}

	fo := fov.Obj.(*FileObject)
	if fo.closed {
		vm.runTimeError("readln attempted on closed file.")
		return makeNilValue()
	}

	rv := fo.readLine()
	if rv.Type == VAL_NIL {
		vm.raiseExceptionByName("EOFError", "End of file reached")
		return makeBooleanValue(true, false)
	}
	return rv
}

func writeBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to writeln.")
		return makeNilValue()
	}
	fov := vm.stack[arg_stackptr]
	str := vm.stack[arg_stackptr+1]

	if fov.Type != VAL_OBJ || fov.Obj.getType() != OBJECT_FILE {
		vm.runTimeError("Invalid argument type to writeln.")
		return makeNilValue()
	}
	if str.Type != VAL_OBJ || str.Obj.getType() != OBJECT_STRING {
		vm.runTimeError("Invalid argument type to writeln.")
		return makeNilValue()
	}

	fo := fov.Obj.(*FileObject)
	if fo.closed {
		vm.runTimeError("writeln attempted on closed file.")
		return makeNilValue()
	}

	fo.write(str)
	return makeBooleanValue(true, false)
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

func (vm *VM) loadBuiltInModule(source string) {
	subvm := NewVM("", false)
	subvm.environments.vars = vm.environments.vars
	DebugSuppress = true
	_, _ = subvm.Interpret(source)
	vm.updateEnvironment(*subvm.environments)
	DebugSuppress = false
}

// predefine an Exception class using Lox source
const exceptionSource = `class Exception {init(msg) {this.msg = msg;this.name = "Exception";  }toString() {return this.msg;}}`
const eofErrorSource = `class EOFError < Exception {init(msg) {this.msg = msg;this.name = "EOFError";  }toString() {return this.msg;}}`
const runTimeErrorSource = `class RunTimeError < Exception {init(msg) {this.msg = msg;this.name = "RunTimeError";  }toString() {return this.msg;}}`
