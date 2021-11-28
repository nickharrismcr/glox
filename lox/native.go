package lox

import (
	"fmt"
	"math"
	"time"
)

func (vm *VM) defineNativeFunctions() {

	vm.defineNative("args", argsNative)
	vm.defineNative("clock", clockNative)
	vm.defineNative("str", strNative)
	vm.defineNative("len", lenNative)
	vm.defineNative("sin", sinNative)
	vm.defineNative("cos", cosNative)
	vm.defineNative("append", appendNative)
	vm.defineNative("float", floatNative)
	vm.defineNative("int", intNative)
}

func argsNative(argCount int, arg_stackptr int, vm *VM) Value {

	argvList := []Value{}
	for _, a := range vm.args {
		argvList = append(argvList, makeObjectValue(makeStringObject(a), true))
	}
	list := makeListObject(argvList)
	return makeObjectValue(list, false)
}

func floatNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.(type) {
	case FloatValue:
		return arg
	case IntValue:
		return makeFloatValue(float64(asInt(arg)), false)
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func intNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.(type) {
	case IntValue:
		return arg
	case FloatValue:
		return makeIntValue(int(asFloat(arg)), false)
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func clockNative(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeFloatValue(float64(elapsed.Seconds()), false)
}

// str(x)
func strNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg := arg.(type) {

	case FloatValue:
		s := arg.String()
		so := makeStringObject(s)
		return makeObjectValue(so, false)

	case IntValue:
		s := arg.String()
		so := makeStringObject(s)
		return makeObjectValue(so, false)

	case NilValue:
		so := makeStringObject("nil")
		return makeObjectValue(so, false)

	case ObjectValue:
		o := arg
		switch o.value.getType() {
		case OBJECT_STRING:
			return arg
		case OBJECT_FUNCTION:
			return makeObjectValue(makeStringObject("<func>"), false)
		case OBJECT_LIST:
			return makeObjectValue(makeStringObject(o.String()), false)
		case OBJECT_CLASS:
			return makeObjectValue(makeStringObject(o.String()), false)
		case OBJECT_INSTANCE:
			return makeObjectValue(makeStringObject(o.String()), false)
		}

	case BooleanValue:
		s := fmt.Sprintf("%t", arg.get())
		so := makeStringObject(s)
		return makeObjectValue(so, false)
	}

	return makeObjectValue(makeStringObject(""), false)
}

// len( string )
func lenNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Invalid argument count to len.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	vobj, ok := val.(ObjectValue)
	if !ok {
		vm.runTimeError("Invalid argument type to len.")
		return makeNilValue()
	}
	switch vobj.get().getType() {
	case OBJECT_STRING:
		s := vobj.get().(StringObject).get()
		return makeIntValue(len(s), false)
	case OBJECT_LIST:
		l := vobj.get().(*ListObject).get()
		return makeIntValue(len(l), false)
	}
	vm.runTimeError("Invalid argument type to len.")
	return makeNilValue()
}

// sin(number)
func sinNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Invalid argument count to sin.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	vn, ok := vnum.(FloatValue)
	if !ok {
		vm.runTimeError("Invalid argument type to sin.")
		return makeNilValue()
	}
	n := vn.get()
	return makeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Invalid argument count to cos.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	vn, ok := vnum.(FloatValue)
	if !ok {
		vm.runTimeError("Invalid argument type to cos.")
		return makeNilValue()
	}
	n := vn.get()
	return makeFloatValue(math.Cos(n), false)
}

// append(obj,value)
func appendNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 2 {
		vm.runTimeError("Invalid argument count to append.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	vobj, ok := val.(ObjectValue)
	if !ok {
		vm.runTimeError("Argument 1 to append must be list.")
		return makeNilValue()
	}
	val2 := vm.stack[arg_stackptr+1]
	switch vobj.get().getType() {

	case OBJECT_LIST:
		l := vobj.get().(*ListObject)
		l.append(val2)
		return val
	}
	vm.runTimeError("Argument 1 to append must be list.")
	return makeNilValue()
}
