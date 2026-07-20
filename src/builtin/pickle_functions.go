package builtin

import (
	"glox/src/core"
)

// DumpsBuiltIn serialises a plain-data Lox value (nil, bool, int, float,
// string, list, tuple, dict, vec2/3/4, class instances, arbitrarily nested)
// to a string holding the raw encoded bytes. An instance is encoded as its
// class name plus its fields -- never its class's methods/code -- so
// loads() can only reconstruct it if a class of that name is already loaded
// in the decoding process (see LoadsBuiltIn). Unsupported values (closures,
// classes themselves, files, native/graphics objects) or cyclic structures
// raise a catchable PickleError rather than crashing the interpreter.
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
// value. Truncated or malformed input raises a catchable PickleError, as
// does an encoded instance whose class isn't loaded in this process --
// class lookup goes through vm.ResolveClass, which checks built-ins, then
// the calling frame's module scope (same resolution order used to look up
// an exception handler's class by name).
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

	val, err := core.DecodeValueResolvingClasses([]byte(dataVal.AsString().Get()), vm.ResolveClass)
	if err != nil {
		vm.RunTimeErrorNamed("PickleError", "%v", err)
		return core.NIL_VALUE
	}

	return val
}
