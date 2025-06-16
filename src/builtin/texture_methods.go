package builtin

import (
	"glox/src/core"
)

func RegisterAllTextureMethods(o *TextureObject) {

	o.RegisterMethod("width", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(int(o.Data.Width), true)
		},
	})
	o.RegisterMethod("height", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(int(o.Data.Height), true)
		},
	})
	o.RegisterMethod("animate", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount < 1 {
				vm.RunTimeError("animate requires at least one argument")

			}
			ticksVal := vm.Stack(arg_stackptr)
			if !ticksVal.IsNumber() {
				vm.RunTimeError("animate requires a number argument for ticks per frame")
				return core.MakeNilValue()
			}
			o.Data.TicksPerFrame = ticksVal.Int
			return core.MakeNilValue()
		},
	})
}
