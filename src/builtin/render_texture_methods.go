package builtin

import (
	"glox/src/core"
	"glox/src/util"

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

	o.RegisterMethod("triangle", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 7 {
				vm.RunTimeError("triangle expects 7 arguments: x1, y1, x2, y2, x3, y3, color")
				return core.NIL_VALUE
			}
			x1val := vm.Stack(arg_stackptr)
			y1val := vm.Stack(arg_stackptr + 1)
			x2val := vm.Stack(arg_stackptr + 2)
			y2val := vm.Stack(arg_stackptr + 3)
			x3val := vm.Stack(arg_stackptr + 4)
			y3val := vm.Stack(arg_stackptr + 5)
			colVal := vm.Stack(arg_stackptr + 6)
			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for rectangle color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := int32(v4obj.X)
			g := int32(v4obj.Y)
			b := int32(v4obj.Z)
			a := int32(v4obj.W)

			x1 := float32(x1val.AsInt())
			y1 := float32(y1val.AsInt())
			x2 := float32(x2val.AsInt())
			y2 := float32(y2val.AsInt())
			x3 := float32(x3val.AsInt())
			y3 := float32(y3val.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawTriangle(rl.Vector2{X: x1, Y: y1}, rl.Vector2{X: x2, Y: y2}, rl.Vector2{X: x3, Y: y3}, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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
	// draw a texture to the render texture
	o.RegisterMethod("draw_texture", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("texture expects 5 arguments: texture, x, y ")
				return core.NIL_VALUE
			}

			texVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)

			if texVal.Type != core.VAL_OBJ {
				vm.RunTimeError("Expected texture in parameter 1 for draw")
				return core.NIL_VALUE
			}
			to, ok := texVal.Obj.(*TextureObject)
			if !ok {
				vm.RunTimeError("Expected texture in parameter 1 for draw")
				return core.NIL_VALUE
			}
			texture := to.Data.Texture

			x := int32(xVal.AsInt())
			y := int32(yVal.AsInt())

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawTexture(texture, x, y, rl.White)
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("draw_texture_pro", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 13 {
				vm.RunTimeError("draw_texture_pro expects 13 arguments: texture, src_x, src_y, src_w, src_h, dest_x, dest_y, dest_w, dest_h, origin_x, origin_y, rotation, color")
				return core.NIL_VALUE
			}
			textureVal := vm.Stack(arg_stackptr)
			srcXVal := vm.Stack(arg_stackptr + 1)
			srcYVal := vm.Stack(arg_stackptr + 2)
			srcWVal := vm.Stack(arg_stackptr + 3)
			srcHVal := vm.Stack(arg_stackptr + 4)
			destXVal := vm.Stack(arg_stackptr + 5)
			destYVal := vm.Stack(arg_stackptr + 6)
			destWVal := vm.Stack(arg_stackptr + 7)
			destHVal := vm.Stack(arg_stackptr + 8)
			originXVal := vm.Stack(arg_stackptr + 9)
			originYVal := vm.Stack(arg_stackptr + 10)
			rotval := vm.Stack(arg_stackptr + 11)
			colVal := vm.Stack(arg_stackptr + 12)

			to, ok := textureVal.Obj.(*TextureObject)
			if !ok {
				vm.RunTimeError("Expected TextureObject for draw_texture_pro")
				return core.NIL_VALUE
			}
			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for texture color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

			srcX := float32(srcXVal.AsFloat())
			srcY := float32(srcYVal.AsFloat())
			srcW := float32(srcWVal.AsFloat())
			srcH := float32(srcHVal.AsFloat())
			destX := float32(destXVal.AsFloat())
			destY := float32(destYVal.AsFloat())
			destW := float32(destWVal.AsFloat())
			destH := float32(destHVal.AsFloat())
			originX := float32(originXVal.AsFloat())
			originY := float32(originYVal.AsFloat())
			rot := float32(rotval.AsFloat())

			srcRect := rl.Rectangle{
				X:      srcX,
				Y:      srcY,
				Width:  srcW,
				Height: srcH,
			}
			destRect := rl.Rectangle{
				X:      destX,
				Y:      destY,
				Width:  destW,
				Height: destH,
			}
			origin := rl.Vector2{
				X: originX,
				Y: originY,
			}

			rl.BeginTextureMode(o.Data.RenderTexture)
			rl.DrawTexturePro(to.Data.Texture, srcRect, destRect, origin, rot, tint)
			rl.EndTextureMode()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("draw_array", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			arrVal := vm.Stack(arg_stackptr)
			arrobj := AsFloatArray(arrVal)
			arr := arrobj.Value

			rl.BeginTextureMode(o.Data.RenderTexture)
			for x := range arr.Width {
				for y := range arr.Height {
					f := arr.Get(x, y)
					r, g, b := util.DecodeRGB(f)
					col := rl.NewColor(r, g, b, 255)
					rl.DrawPixel(int32(x), int32(y), col)
				}
			}
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
