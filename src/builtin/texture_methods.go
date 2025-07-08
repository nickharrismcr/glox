package builtin

import (
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllTextureMethods(o *TextureObject) {

	o.RegisterMethod("width", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(int(o.Data.Width), true)
		},
	})
	o.RegisterMethod("frame_width", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeIntValue(int(o.Data.FrameWidth), true)
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
				return core.NIL_VALUE
			}
			o.Data.TicksPerFrame = ticksVal.Int
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("unload", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			// Unload the texture from GPU memory
			rl.UnloadTexture(o.Data.Texture)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("set_wrap_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount < 1 {
				vm.RunTimeError("set_wrap_mode requires one argument")
				return core.NIL_VALUE
			}
			wrapModeVal := vm.Stack(arg_stackptr)
			if !wrapModeVal.IsNumber() {
				vm.RunTimeError("set_wrap_mode requires a number argument for wrap mode")
				return core.NIL_VALUE
			}
			wrapMode := rl.TextureWrapMode(wrapModeVal.Int)
			rl.SetTextureWrap(o.Data.Texture, wrapMode)
			return core.NIL_VALUE
		},
	})
}
