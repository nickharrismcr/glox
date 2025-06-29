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
			if argCount != 5 {
				vm.RunTimeError("line expects 5 arguments: x1, y1, x2, y2, color")
				return core.NIL_VALUE
			}

			x1Val := vm.Stack(arg_stackptr)
			y1Val := vm.Stack(arg_stackptr + 1)
			x2Val := vm.Stack(arg_stackptr + 2)
			y2Val := vm.Stack(arg_stackptr + 3)
			colVal := vm.Stack(arg_stackptr + 4)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for line color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x1 := int32(x1Val.AsFloat())
			y1 := int32(y1Val.AsFloat())
			x2 := int32(x2Val.AsFloat())
			y2 := int32(y2Val.AsFloat())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawLine(x1, y1, x2, y2, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("line_ex", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 6 {
				vm.RunTimeError("line_ex expects 6 arguments: x1, y1, x2, y2, thickness, color")
				return core.NIL_VALUE
			}

			x1Val := vm.Stack(arg_stackptr)
			y1Val := vm.Stack(arg_stackptr + 1)
			x2Val := vm.Stack(arg_stackptr + 2)
			y2Val := vm.Stack(arg_stackptr + 3)
			thickVal := vm.Stack(arg_stackptr + 4)
			colVal := vm.Stack(arg_stackptr + 5)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for line color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			x1 := float32(x1Val.AsFloat())
			y1 := float32(y1Val.AsFloat())
			x2 := float32(x2Val.AsFloat())
			y2 := float32(y2Val.AsFloat())
			thickness := float32(thickVal.AsFloat())

			rlv1 := rl.Vector2{X: x1, Y: y1}
			rlv2 := rl.Vector2{X: x2, Y: y2}
			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawLineEx(rlv1, rlv2, thickness, rl.NewColor(r, g, b, a))
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
			colVal := vm.Stack(arg_stackptr + 4)
			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for rectangle color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			w := int32(wval.AsInt())
			h := int32(hval.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawRectangle(x, y, w, h, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("circle_fill", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("circle_fill expects 4 arguments: x, y, radius, color")
				return core.NIL_VALUE
			}

			xVal := vm.Stack(arg_stackptr)
			yVal := vm.Stack(arg_stackptr + 1)
			radVal := vm.Stack(arg_stackptr + 2)
			colVal := vm.Stack(arg_stackptr + 3)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for circle color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())
			rad := float32(radVal.AsFloat())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawCircle(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("pixel", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("pixel expects 3 arguments: x, y, color")
				return core.NIL_VALUE
			}

			xVal := vm.Stack(arg_stackptr)
			yVal := vm.Stack(arg_stackptr + 1)
			colVal := vm.Stack(arg_stackptr + 2)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for pixel color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawPixel(x, y, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("circle", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("circle expects 4 arguments: x, y, radius, color")
				return core.NIL_VALUE
			}

			xVal := vm.Stack(arg_stackptr)
			yVal := vm.Stack(arg_stackptr + 1)
			radVal := vm.Stack(arg_stackptr + 2)
			colVal := vm.Stack(arg_stackptr + 3)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for circle color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())
			rad := float32(radVal.AsFloat())

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
