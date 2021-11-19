package lox

import (
	"fmt"
	"time"
)

func clockNative(argCount int, arg_stackptr int, vm *VM) Value {
	elapsed := time.Since(vm.starttime)
	return makeNumberValue(float64(elapsed.Seconds()), false)
}

func strNative(argcount int, arg_stackptr int, vm *VM) Value {
	if argcount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]
	switch arg.(type) {
	case NumberValue:
		s := fmt.Sprintf("%f", arg.(NumberValue).Get())
		so := MakeStringObject(s)
		return makeObjectValue(so, false)
	case ObjectValue:
		o := arg.(ObjectValue)
		switch o.value.getType() {
		case OBJECT_STRING:
			return arg
		case OBJECT_FUNCTION:
			return makeObjectValue(MakeStringObject("<func>"), false)
		}
	case BooleanValue:
		s := fmt.Sprintf("%t", arg.(BooleanValue).Get())
		so := MakeStringObject(s)
		return makeObjectValue(so, false)
	}
	return makeObjectValue(MakeStringObject(""), false)
}
