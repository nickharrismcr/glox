package lox

import (
	"fmt"
	"math"
	"time"
)

func (vm *VM) defineNatives() {

	vm.defineNative("clock", clockNative)
	vm.defineNative("str", strNative)
	vm.defineNative("len", lenNative)
	vm.defineNative("sin", sinNative)
	vm.defineNative("cos", cosNative)
}

func clockNative(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeNumberValue(float64(elapsed.Seconds()), false)
}

// str(x)
func strNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.(type) {

	case NumberValue:
		s := arg.(NumberValue).String()
		so := makeStringObject(s)
		return makeObjectValue(so, false)

	case NilValue:
		so := makeStringObject("nil")
		return makeObjectValue(so, false)

	case ObjectValue:
		o := arg.(ObjectValue)
		switch o.value.getType() {
		case OBJECT_STRING:
			return arg
		case OBJECT_FUNCTION:
			return makeObjectValue(makeStringObject("<func>"), false)
		case OBJECT_LIST:
			return makeObjectValue(makeStringObject(o.String()), false)
		}

	case BooleanValue:
		s := fmt.Sprintf("%t", arg.(BooleanValue).get())
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
		return makeNumberValue(float64(len(s)), false)
	case OBJECT_LIST:
		l := vobj.get().(*ListObject).get()
		return makeNumberValue(float64(len(l)), false)
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

	vn, ok := vnum.(NumberValue)
	if !ok {
		vm.runTimeError("Invalid argument type to sin.")
		return makeNilValue()
	}
	n := vn.get()
	return makeNumberValue(math.Sin(n), false)
}

// cos(number)
func cosNative(argcount int, arg_stackptr int, vm *VM) Value {

	if argcount != 1 {
		vm.runTimeError("Invalid argument count to cos.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	vn, ok := vnum.(NumberValue)
	if !ok {
		vm.runTimeError("Invalid argument type to cos.")
		return makeNilValue()
	}
	n := vn.get()
	return makeNumberValue(math.Cos(n), false)
}
