package builtin

import (
	"glox/src/core"
)

func RegisterAllFloatArrayMethods(o *FloatArrayObject) {

	o.RegisterMethod("width", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(o.Value.Width, true)
		},
	})
	o.RegisterMethod("height", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(o.Value.Height, true)
		},
	})
	o.RegisterMethod("get", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			x := xval.AsInt()
			y := yval.AsInt()
			return core.MakeFloatValue(o.Value.Get(x, y), false)
		},
	})
	o.RegisterMethod("set", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			fval := vm.Stack(arg_stackptr + 2)
			x := xval.AsInt()
			y := yval.AsInt()
			f := fval.AsFloat()
			o.Value.Set(x, y, f)
			return core.MakeNilValue()
		},
	})
	o.RegisterMethod("clear", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			fval := vm.Stack(arg_stackptr)
			f := fval.AsFloat()
			o.Value.Clear(f)
			return core.MakeNilValue()
		},
	})
}
