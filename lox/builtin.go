package lox

import (
	"fmt"
	"math"
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
	vm.defineBuiltIn("append", appendBuiltIn)
	vm.defineBuiltIn("float", floatBuiltIn)
	vm.defineBuiltIn("int", intBuiltIn)
	vm.defineBuiltIn("join", joinBuiltIn)
	vm.defineBuiltIn("keys", keysBuiltIn)
	vm.defineBuiltIn("lox_mandel", mandelBuiltIn)
	vm.defineBuiltIn("replace", replaceBuiltIn)
	vm.defineBuiltIn("open", openBuiltIn)
	vm.defineBuiltIn("close", closeBuiltIn)
	vm.defineBuiltIn("readln", readlnBuiltIn)
	vm.defineBuiltIn("write", writeBuiltIn)

	// lox built ins e.g Exception classes
	vm.loadBuiltInModule(exceptionSource)
	vm.loadBuiltInModule(eofErrorSource)
	vm.loadBuiltInModule(runTimeErrorSource)

}

func keysBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to keys.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]

	switch val.Type {
	case VAL_OBJ:
		if val.isDictObject() {
			d := val.asDict()
			return d.keys()
		}
	}
	vm.runTimeError("Argument to keys must be a dictionary.")
	return makeNilValue()
}

func joinBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	var err string = ""

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to join.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	err = "Argument 2 to join must be a string."

	switch val.Type {
	case VAL_OBJ:
		if val.isListObject() {
			err = "Argument 2 to join must be a string."
			l := val.asList()
			val2 := vm.stack[arg_stackptr+1]
			switch val2.Type {
			case VAL_OBJ:
				if val2.isStringObject() {
					rv, errj := l.join(val2.asString().get())
					if errj == nil {
						return rv
					} else {
						err = errj.Error()
					}
				}
			}
		}
	}
	vm.runTimeError(err)
	return makeNilValue()
}

func typeBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
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
	return makeObjectValue(makeStringObject(val_type), true)
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
	}
	vm.runTimeError("Argument must be number.")
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
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func clockBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeFloatValue(float64(elapsed.Seconds()), false)
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
		return val
	}
	vm.runTimeError("Argument 1 to append must be list.")
	return makeNilValue()
}

func mandelBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

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

	if iteration == maxIteration {
		return makeIntValue(0, false)
	}
	return makeIntValue(iteration, false)

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
		vm.runTimeError(fmt.Sprintf("%s", err))
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
	subvm.SetGlobals(vm.globals)
	DebugSuppress = true
	_, _ = subvm.Interpret(source)
	vm.SetGlobals(subvm.globals)
	DebugSuppress = false
}

// predefine an Exception class using Lox source
const exceptionSource = `class Exception {init(msg) {this.msg = msg;this.name = "Exception";  }toString() {return this.msg;}}`
const eofErrorSource = `class EOFError < Exception {init(msg) {this.msg = msg;this.name = "EOFError";  }toString() {return this.msg;}}`
const runTimeErrorSource = `class RunTimeError < Exception {init(msg) {this.msg = msg;this.name = "RunTimeError";  }toString() {return this.msg;}}`
