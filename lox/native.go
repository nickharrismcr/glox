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

// substr( string s, int start, int len )
func substrNative(argcount int, arg_stackptr int, vm *VM) Value {
	if argcount != 3 {
		vm.runTimeError("Invalid argument count to substr.")
		return makeNilValue()
	}
	vstring_ := vm.stack[arg_stackptr]
	vstart := vm.stack[arg_stackptr+1]
	vlen := vm.stack[arg_stackptr+2]

	vo_string, ok := vstring_.(ObjectValue)
	if !ok {
		vm.runTimeError("Invalid argument 1 type to substr.")
		return makeNilValue()
	}

	vn_start, ok := vstart.(NumberValue)
	if !ok {
		vm.runTimeError("Invalid argument 2 type to substr.")
		return makeNilValue()
	}

	vn_len, ok := vlen.(NumberValue)
	if !ok {
		vm.runTimeError("Invalid argument 3 type to substr.")
		return makeNilValue()
	}

	string_ := vo_string.String()
	start := int(vn_start.Get())
	if start < 1 || start > len(string_) {
		vm.runTimeError("substr() start out of bounds.")
		return makeNilValue()
	}

	length := int(vn_len.Get())
	if (start+length)-1 > len(string_) {
		length = (len(string_) - start) + 1
	}
	return makeObjectValue(MakeStringObject(string_[start-1:(start+length)-1]), false)
}

// len( string )
func lenNative(argcount int, arg_stackptr int, vm *VM) Value {
	if argcount != 1 {
		vm.runTimeError("Invalid argument count to len.")
		return makeNilValue()
	}
	vstring_ := vm.stack[arg_stackptr]

	vo_string, ok := vstring_.(ObjectValue)
	if !ok {
		vm.runTimeError("Invalid argument type to len.")
		return makeNilValue()
	}
	s := vo_string.String()
	return makeNumberValue(float64(len(s)), false)
}
