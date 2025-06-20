package builtin

import (
	"glox/src/core"
	"glox/src/util"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllGraphicsMethods(o *GraphicsObject) {
	o.RegisterMethod("init", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.SetTraceLogLevel(rl.LogNone)
			rl.InitWindow(o.Value.Width, o.Value.Height, "GLOX")
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("begin", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.BeginDrawing()
			rl.BeginBlendMode(rl.BlendAdditive)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("begin_blend_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			modeVal := vm.Stack(arg_stackptr)
			o.Value.SetBlendMode(modeVal.AsString().Get())
			rl.BeginBlendMode(o.Value.Blend_mode)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("end_blend_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.EndBlendMode()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("end", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.DrawFPS(10, 10)
			rl.EndDrawing()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("clear", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rval := vm.Stack(arg_stackptr)
			gval := vm.Stack(arg_stackptr + 1)
			bval := vm.Stack(arg_stackptr + 2)
			aval := vm.Stack(arg_stackptr + 3)
			r := rval.AsInt()
			g := gval.AsInt()
			b := bval.AsInt()
			a := aval.AsInt()
			rl.ClearBackground(rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawLine(x1, y1, x2, y2, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawRectangle(x, y, w, h, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawCircle(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawPixel(x, y, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawCircleLines(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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

			rl.DrawText(s, x, y, 10, rl.White)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_array", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			arrVal := vm.Stack(arg_stackptr)
			arrobj := AsFloatArray(arrVal)
			arr := arrobj.Value

			for x := range arr.Width {
				for y := range arr.Height {
					f := arr.Get(x, y)
					r, g, b := util.DecodeRGB(f)
					col := rl.NewColor(r, g, b, 255)
					rl.DrawPixel(int32(x), int32(y), col)
				}
			}

			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("should_close", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			return core.MakeBooleanValue(rl.WindowShouldClose(), true)
		},
	})
	o.RegisterMethod("close", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.CloseWindow()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_texture", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			textureVal := vm.Stack(arg_stackptr)
			xval := vm.Stack(arg_stackptr + 1)
			yval := vm.Stack(arg_stackptr + 2)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())

			to := textureVal.Obj.(*TextureObject)
			rect := to.Data.GetFrameRect()
			rl.DrawTextureRec(to.Data.Texture, rect, rl.Vector2{X: float32(x), Y: float32(y)}, rl.White)
			to.Data.Animate()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_texture_rect", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			textureVal := vm.Stack(arg_stackptr)
			xval := vm.Stack(arg_stackptr + 1)
			yval := vm.Stack(arg_stackptr + 2)
			rectx0Val := vm.Stack(arg_stackptr + 3)
			recty0Val := vm.Stack(arg_stackptr + 4)
			rectWVal := vm.Stack(arg_stackptr + 5)
			rectHVal := vm.Stack(arg_stackptr + 6)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			rectX0 := int32(rectx0Val.AsInt())
			rectY0 := int32(recty0Val.AsInt())
			rectW := int32(rectWVal.AsInt())
			rectH := int32(rectHVal.AsInt())

			to := textureVal.Obj.(*TextureObject)
			rect := rl.Rectangle{
				X:      float32(rectX0),
				Y:      float32(rectY0),
				Width:  float32(rectW),
				Height: float32(rectH),
			}

			rl.DrawTextureRec(to.Data.Texture, rect, rl.Vector2{X: float32(x), Y: float32(y)}, rl.White)
			to.Data.Animate()
			return core.NIL_VALUE
		},
	})

}
