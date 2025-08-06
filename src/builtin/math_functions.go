package builtin

import (
	"glox/src/core"
	"math"
)

// Math functions

func SinBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sin.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to sin.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sin(n), false)
}

func CosBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to cos.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to cos.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Cos(n), false)
}

func TanBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to tan.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to tan.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Tan(n), false)
}

func SqrtBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to sqrt.")
		return core.NIL_VALUE
	}
	vnum := vm.Stack(arg_stackptr)

	if vnum.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to sqrt.")
		return core.NIL_VALUE
	}
	n := vnum.Float
	return core.MakeFloatValue(math.Sqrt(n), false)
}

func PowBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to pow.")
		return core.NIL_VALUE
	}
	vbase := vm.Stack(arg_stackptr)
	vexp := vm.Stack(arg_stackptr + 1)

	if vbase.Type != core.VAL_FLOAT || vexp.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to pow.")
		return core.NIL_VALUE
	}
	n := vbase.Float
	return core.MakeFloatValue(math.Pow(n, vexp.Float), false)
}

func Atan2BuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to atan2.")
		return core.NIL_VALUE
	}
	vnum1 := vm.Stack(arg_stackptr)
	vnum2 := vm.Stack(arg_stackptr + 1)

	if vnum1.Type != core.VAL_FLOAT || vnum2.Type != core.VAL_FLOAT {
		vm.RunTimeError("Invalid argument type to atan2.")
		return core.NIL_VALUE
	}
	n1 := vnum1.Float
	n2 := vnum2.Float
	return core.MakeFloatValue(math.Atan2(n1, n2), false)
}
