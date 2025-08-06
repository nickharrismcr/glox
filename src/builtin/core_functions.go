package builtin

import (
	"fmt"
	"glox/src/core"
	"glox/src/debug"
	"math/rand"
	"time"
)

// Core utility functions

func TypeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Single argument expected.")
		return core.NIL_VALUE
	}
	val := vm.Stack(arg_stackptr)
	name := typeName(val)
	return core.MakeStringObjectValue(name, true)
}

func ArgsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	argvList := []core.Value{}
	for _, a := range vm.Args() {
		argvList = append(argvList, core.MakeStringObjectValue(a, true))
	}
	list := core.MakeListObject(argvList, false)
	return core.MakeObjectValue(list, false)
}

func ClockBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	elapsed := time.Since(vm.StartTime())
	return core.MakeFloatValue(float64(elapsed.Seconds()), false)
}

func FloatBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func IntBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func RandBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	return core.MakeFloatValue(rand.Float64(), false)
}

func LenBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func AppendBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func ReplaceBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func FormatBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func SleepBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func RangeBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

// Vector creation functions
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

// Debug/inspection functions
func DumpFrameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
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

func GetFrameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	return debug.FrameDictValue(vm)
}
