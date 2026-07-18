package builtin

import (
	"glox/src/core"
)

// DumpsBuiltIn serialises a plain-data Lox value (nil, bool, int, float,
// string, list, tuple, dict, vec2/3/4, arbitrarily nested) to a string
// holding the raw encoded bytes. Unsupported values (closures, classes,
// instances, files, native/graphics objects) or cyclic structures raise a
// catchable PickleError rather than crashing the interpreter.
func DumpsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to dumps.")
		return core.NIL_VALUE
	}

	val := vm.Stack(arg_stackptr)
	data, err := core.EncodeValue(val)
	if err != nil {
		vm.RunTimeErrorNamed("PickleError", "%v", err)
		return core.NIL_VALUE
	}

	return core.MakeStringObjectValue(string(data), false)
}

// LoadsBuiltIn deserialises a string produced by dumps back into a Lox
// value. Truncated or malformed input raises a catchable PickleError.
func LoadsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to loads.")
		return core.NIL_VALUE
	}

	dataVal := vm.Stack(arg_stackptr)
	if dataVal.Type != core.VAL_OBJ || dataVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to loads, expected string.")
		return core.NIL_VALUE
	}

	val, err := core.DecodeValue([]byte(dataVal.AsString().Get()))
	if err != nil {
		vm.RunTimeErrorNamed("PickleError", "%v", err)
		return core.NIL_VALUE
	}

	return val
}
