package builtin

import (
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllRenderTextureMethods(o *RenderTextureObject) {

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

	o.RegisterMethod("clear", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			v4val := vm.Stack(arg_stackptr)
			if v4val.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vector4")
			}
			v4 := v4val.Obj.(*core.Vec4Object)
			rval := v4.X
			gval := v4.Y
			bval := v4.Z
			aval := v4.W

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.ClearBackground(rl.NewColor(uint8(rval), uint8(gval), uint8(bval), uint8(aval)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("line", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			x1val := vm.Stack(arg_stackptr)
			y1val := vm.Stack(arg_stackptr + 1)
			x2val := vm.Stack(arg_stackptr + 2)
			y2val := vm.Stack(arg_stackptr + 3)
			rval := vm.Stack(arg_stackptr + 4)
			gval := vm.Stack(arg_stackptr + 5)
			bval := vm.Stack(arg_stackptr + 6)
			aval := vm.Stack(arg_stackptr + 7)

			x1 := int32(x1val.AsInt())
			y1 := int32(y1val.AsInt())
			x2 := int32(x2val.AsInt())
			y2 := int32(y2val.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawLine(x1, y1, x2, y2, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("rectangle", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			wval := vm.Stack(arg_stackptr + 2)
			hval := vm.Stack(arg_stackptr + 3)
			rval := vm.Stack(arg_stackptr + 4)
			gval := vm.Stack(arg_stackptr + 5)
			bval := vm.Stack(arg_stackptr + 6)
			aval := vm.Stack(arg_stackptr + 7)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			w := int32(wval.AsInt())
			h := int32(hval.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawRectangle(x, y, w, h, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("circle_fill", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			radVal := vm.Stack(arg_stackptr + 2)
			rval := vm.Stack(arg_stackptr + 3)
			gval := vm.Stack(arg_stackptr + 4)
			bval := vm.Stack(arg_stackptr + 5)
			aval := vm.Stack(arg_stackptr + 6)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			rad := float32(radVal.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawCircle(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("pixel", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			rval := vm.Stack(arg_stackptr + 2)
			gval := vm.Stack(arg_stackptr + 3)
			bval := vm.Stack(arg_stackptr + 4)
			aval := vm.Stack(arg_stackptr + 5)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawPixel(x, y, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("circle", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			radVal := vm.Stack(arg_stackptr + 2)
			rval := vm.Stack(arg_stackptr + 3)
			gval := vm.Stack(arg_stackptr + 4)
			bval := vm.Stack(arg_stackptr + 5)
			aval := vm.Stack(arg_stackptr + 6)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			rad := float32(radVal.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawCircleLines(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("text", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			sval := vm.Stack(arg_stackptr + 2)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			s := sval.AsString().Get()

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawText(s, x, y, 10, rl.White)
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})

}
