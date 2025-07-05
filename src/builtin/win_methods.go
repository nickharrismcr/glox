package builtin

import (
	"glox/src/core"
	"glox/src/util"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllWindowMethods(o *WindowObject) {
	o.RegisterMethod("init", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.SetTraceLogLevel(rl.LogNone)
			rl.SetConfigFlags(rl.FlagVsyncHint) // Enable VSync hint
			rl.InitWindow(o.Value.Width, o.Value.Height, "GLOX")
			rl.SetTargetFPS(60)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("begin", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.BeginDrawing()
			rl.BeginBlendMode(rl.BlendAlpha)
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("begin_blend_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

			modeVal := vm.Stack(arg_stackptr)
			o.Value.SetBlendMode(modeVal.Int)
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
	o.RegisterMethod("toggle_fullscreen", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if !rl.IsWindowFullscreen() {
				monitor := rl.GetCurrentMonitor()
				width := rl.GetMonitorWidth(monitor)
				height := rl.GetMonitorHeight(monitor)

				// Set window size to monitor size
				rl.SetWindowSize(width, height)
			}
			rl.ToggleFullscreen()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("get_screen_width", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			width := rl.GetScreenWidth()
			return core.MakeFloatValue(float64(width), true)
		},
	})
	o.RegisterMethod("get_screen_height", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			height := rl.GetScreenHeight()
			return core.MakeFloatValue(float64(height), true)
		},
	})

	o.RegisterMethod("clear", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			v4val := vm.Stack(arg_stackptr)
			if v4val.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for clear color")
				return core.NIL_VALUE
			}
			v4obj := v4val.Obj.(*core.Vec4Object)
			r := v4obj.X
			g := v4obj.Y
			b := v4obj.Z
			a := v4obj.W
			rl.ClearBackground(rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			x1 := int32(x1Val.AsFloat())
			y1 := int32(y1Val.AsFloat())
			x2 := int32(x2Val.AsFloat())
			y2 := int32(y2Val.AsFloat())

			rl.DrawLine(x1, y1, x2, y2, rl.NewColor(r, g, b, a))
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

			rl.DrawLineEx(rlv1, rlv2, thickness, rl.NewColor(r, g, b, a))
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("triangle", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 7 {
				vm.RunTimeError("triangle expects 7 arguments: x1, y1, x2, y2, x3, y3, color")
				return core.NIL_VALUE
			}

			x1Val := vm.Stack(arg_stackptr)
			y1Val := vm.Stack(arg_stackptr + 1)
			x2Val := vm.Stack(arg_stackptr + 2)
			y2Val := vm.Stack(arg_stackptr + 3)
			x3Val := vm.Stack(arg_stackptr + 4)
			y3Val := vm.Stack(arg_stackptr + 5)
			colVal := vm.Stack(arg_stackptr + 6)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for triangle color")
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
			x3 := float32(x3Val.AsFloat())
			y3 := float32(y3Val.AsFloat())

			rlv1 := rl.Vector2{X: x1, Y: y1}
			rlv2 := rl.Vector2{X: x2, Y: y2}
			rlv3 := rl.Vector2{X: x3, Y: y3}

			rl.DrawTriangle(rlv1, rlv2, rlv3, rl.NewColor(r, g, b, a))
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

			rl.DrawRectangle(x, y, w, h, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
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
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())
			radius := float32(radVal.AsFloat())

			rl.DrawCircle(x, y, radius, rl.NewColor(r, g, b, a))
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
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())

			rl.DrawPixel(x, y, rl.NewColor(r, g, b, a))
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
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())
			radius := float32(radVal.AsFloat())

			rl.DrawCircleLines(x, y, radius, rl.NewColor(r, g, b, a))
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("text", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 5 {
				vm.RunTimeError("text expects 5 arguments: text, x, y, size, color")
				return core.NIL_VALUE
			}

			textVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)
			sizeVal := vm.Stack(arg_stackptr + 3)
			colVal := vm.Stack(arg_stackptr + 4)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for text color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)

			text := textVal.AsString().Get()
			x := int32(xVal.AsFloat())
			y := int32(yVal.AsFloat())
			size := int32(sizeVal.AsFloat())

			rl.DrawText(text, x, y, size, rl.NewColor(r, g, b, a))
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
	o.RegisterMethod("set_target_fps", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("set_target_fps expects 1 argument (fps)")
				return core.NIL_VALUE
			}
			fpsVal := vm.Stack(arg_stackptr)
			if !fpsVal.IsInt() {
				vm.RunTimeError("set_target_fps argument must be an integer")
				return core.NIL_VALUE
			}
			rl.SetTargetFPS(int32(fpsVal.Int))
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_texture", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("draw_texture expects 4 arguments: texture, x, y, color")
				return core.NIL_VALUE
			}

			textureVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)
			colVal := vm.Stack(arg_stackptr + 3)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for texture color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

			x := float32(xVal.AsFloat())
			y := float32(yVal.AsFloat())

			to := textureVal.Obj.(*TextureObject)
			rect := to.Data.GetFrameRect()
			rl.DrawTextureRec(to.Data.Texture, rect, rl.Vector2{X: x, Y: y}, tint)
			to.Data.Animate()
			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_render_texture", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("draw_render_texture expects 4 arguments: render_texture, x, y, color")
				return core.NIL_VALUE
			}

			renderTextureVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)
			colVal := vm.Stack(arg_stackptr + 3)

			if renderTextureVal.Type != core.VAL_OBJ {
				vm.RunTimeError("Expected RenderTexture object")
				return core.NIL_VALUE
			}

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for render texture color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

			x := float32(xVal.AsFloat())
			y := float32(yVal.AsFloat())

			to := renderTextureVal.Obj.(*RenderTextureObject)
			target := to.Data.RenderTexture.Texture
			rl.DrawTextureRec(target, rl.Rectangle{X: 0, Y: 0, Width: float32(target.Width), Height: float32(-target.Height)}, rl.Vector2{X: x, Y: y}, tint)

			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_render_texture_ex", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 6 {
				vm.RunTimeError("draw_render_texture_ex expects 6 arguments: render_texture, x, y, rotation, scale, color")
				return core.NIL_VALUE
			}

			renderTextureVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)
			rotVal := vm.Stack(arg_stackptr + 3)
			scaleVal := vm.Stack(arg_stackptr + 4)
			colVal := vm.Stack(arg_stackptr + 5)

			if renderTextureVal.Type != core.VAL_OBJ {
				vm.RunTimeError("Expected RenderTexture object")
				return core.NIL_VALUE
			}

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for render texture color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

			x := float32(xVal.AsFloat())
			y := float32(yVal.AsFloat())
			rot := float32(rotVal.AsFloat())
			scale := float32(scaleVal.AsFloat())

			to := renderTextureVal.Obj.(*RenderTextureObject)
			target := to.Data.RenderTexture.Texture

			rl.DrawTextureEx(target, rl.Vector2{X: x, Y: y}, rot, scale, tint)

			return core.NIL_VALUE
		},
	})
	o.RegisterMethod("draw_texture_rect", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 8 {
				vm.RunTimeError("draw_texture_rect expects 8 arguments: texture, x, y, src_x, src_y, src_w, src_h, color")
				return core.NIL_VALUE
			}

			textureVal := vm.Stack(arg_stackptr)
			xVal := vm.Stack(arg_stackptr + 1)
			yVal := vm.Stack(arg_stackptr + 2)
			srcXVal := vm.Stack(arg_stackptr + 3)
			srcYVal := vm.Stack(arg_stackptr + 4)
			srcWVal := vm.Stack(arg_stackptr + 5)
			srcHVal := vm.Stack(arg_stackptr + 6)
			colVal := vm.Stack(arg_stackptr + 7)

			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for texture color")
				return core.NIL_VALUE
			}

			v4obj := colVal.Obj.(*core.Vec4Object)
			tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

			x := float32(xVal.AsFloat())
			y := float32(yVal.AsFloat())
			srcX := float32(srcXVal.AsFloat())
			srcY := float32(srcYVal.AsFloat())
			srcW := float32(srcWVal.AsFloat())
			srcH := float32(srcHVal.AsFloat())

			to := textureVal.Obj.(*TextureObject)
			rect := rl.Rectangle{
				X:      srcX,
				Y:      srcY,
				Width:  srcW,
				Height: srcH,
			}

			rl.DrawTextureRec(to.Data.Texture, rect, rl.Vector2{X: x, Y: y}, tint)
			to.Data.Animate()
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

			var to *TextureObject
			var rto *RenderTextureObject
			var texture rl.Texture2D

			to, ok := textureVal.Obj.(*TextureObject)
			if !ok {
				rto, ok = textureVal.Obj.(*RenderTextureObject)
				if !ok {
					vm.RunTimeError("Expected TextureObject for draw_texture_pro")
					return core.NIL_VALUE
				} else {
					texture = rto.Data.RenderTexture.Texture
				}
			} else {
				texture = to.Data.Texture
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

			rl.DrawTexturePro(texture, srcRect, destRect, origin, rot, tint)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("key_down", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("key_down takes one win.KEY_XXX argument.")
				return core.NIL_VALUE
			}
			keyVal := vm.Stack(arg_stackptr)

			isDown := rl.IsKeyDown(int32(keyVal.Int))
			return core.MakeBooleanValue(isDown, true)
		},
	})
	// arg should be an rl.KeyCode looked up in the constants
	o.RegisterMethod("key_pressed", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("key_pressed takes one win.KEY_XXX argument.")
				return core.NIL_VALUE
			}
			keyVal := vm.Stack(arg_stackptr)

			isPressed := rl.IsKeyPressed(int32(keyVal.Int))
			return core.MakeBooleanValue(isPressed, true)
		},
	})

	// 3D Mode methods
	o.RegisterMethod("begin_3d", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("begin_3d expects 1 argument (camera)")
				return core.NIL_VALUE
			}
			cameraVal := vm.Stack(arg_stackptr)
			if cameraVal.Type != core.VAL_OBJ {
				vm.RunTimeError("Expected camera object")
				return core.NIL_VALUE
			}

			// Type assertion to check if it's a CameraObject
			cameraObj, ok := cameraVal.Obj.(*CameraObject)
			if !ok {
				vm.RunTimeError("Expected camera object")
				return core.NIL_VALUE
			}

			rl.BeginMode3D(cameraObj.Camera)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("end_3d", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.EndMode3D()
			return core.NIL_VALUE
		},
	})

	// Shader mode methods
	o.RegisterMethod("begin_shader_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("begin_shader_mode expects 1 argument (shader)")
				return core.NIL_VALUE
			}

			shaderVal := vm.Stack(arg_stackptr)
			if shaderVal.Type != core.VAL_OBJ {
				vm.RunTimeError("begin_shader_mode expects shader object")
				return core.NIL_VALUE
			}

			// Type assertion to check if it's a ShaderObject
			shaderObj, ok := shaderVal.Obj.(*ShaderObject)
			if !ok {
				vm.RunTimeError("Expected shader object")
				return core.NIL_VALUE
			}

			rl.BeginShaderMode(shaderObj.Value)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("end_shader_mode", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.EndShaderMode()
			return core.NIL_VALUE
		},
	})

	// 3D Drawing primitives
	o.RegisterMethod("cube", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("cube expects 3 arguments: position(vec3), size(vec3), color(vec4)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			sizeVal := vm.Stack(arg_stackptr + 1)
			colorVal := vm.Stack(arg_stackptr + 2)

			if posVal.Type != core.VAL_VEC3 || sizeVal.Type != core.VAL_VEC3 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("cube arguments must be vec3, vec3, vec4")
				return core.NIL_VALUE
			}

			posObj := posVal.Obj.(*core.Vec3Object)
			sizeObj := sizeVal.Obj.(*core.Vec3Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			position := rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)}
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			rl.DrawCube(position, float32(sizeObj.X), float32(sizeObj.Y), float32(sizeObj.Z), color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("cube_wires", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("cube_wires expects 3 arguments: position(vec3), size(vec3), color(vec4)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			sizeVal := vm.Stack(arg_stackptr + 1)
			colorVal := vm.Stack(arg_stackptr + 2)

			if posVal.Type != core.VAL_VEC3 || sizeVal.Type != core.VAL_VEC3 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("cube_wires arguments must be vec3, vec3, vec4")
				return core.NIL_VALUE
			}

			posObj := posVal.Obj.(*core.Vec3Object)
			sizeObj := sizeVal.Obj.(*core.Vec3Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			position := rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)}
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			rl.DrawCubeWires(position, float32(sizeObj.X), float32(sizeObj.Y), float32(sizeObj.Z), color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("sphere", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("sphere expects 3 arguments: center(vec3), radius(number), color(vec4)")
				return core.NIL_VALUE
			}

			centerVal := vm.Stack(arg_stackptr)
			radiusVal := vm.Stack(arg_stackptr + 1)
			colorVal := vm.Stack(arg_stackptr + 2)

			if centerVal.Type != core.VAL_VEC3 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("sphere arguments must be vec3, number, vec4")
				return core.NIL_VALUE
			}

			centerObj := centerVal.Obj.(*core.Vec3Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			center := rl.Vector3{X: float32(centerObj.X), Y: float32(centerObj.Y), Z: float32(centerObj.Z)}
			radius := float32(radiusVal.AsFloat())
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			rl.DrawSphere(center, radius, color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("cylinder", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 5 {
				vm.RunTimeError("cylinder expects 5 arguments: position(vec3), radius_top(number), radius_bottom(number), height(number), color(vec4)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			radiusTopVal := vm.Stack(arg_stackptr + 1)
			radiusBottomVal := vm.Stack(arg_stackptr + 2)
			heightVal := vm.Stack(arg_stackptr + 3)
			colorVal := vm.Stack(arg_stackptr + 4)

			if posVal.Type != core.VAL_VEC3 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("cylinder position and color must be vec3 and vec4")
				return core.NIL_VALUE
			}

			posObj := posVal.Obj.(*core.Vec3Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			position := rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)}
			radiusTop := float32(radiusTopVal.AsFloat())
			radiusBottom := float32(radiusBottomVal.AsFloat())
			height := float32(heightVal.AsFloat())
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			rl.DrawCylinder(position, radiusTop, radiusBottom, height, 16, color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("grid", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("grid expects 2 arguments: slices(int), spacing(float)")
				return core.NIL_VALUE
			}

			slicesVal := vm.Stack(arg_stackptr)
			spacingVal := vm.Stack(arg_stackptr + 1)

			slices := int32(slicesVal.AsInt())
			spacing := float32(spacingVal.AsFloat())

			rl.DrawGrid(slices, spacing)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("plane", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("plane expects 3 arguments: center(vec3), size(vec2), color(vec4)")
				return core.NIL_VALUE
			}

			centerVal := vm.Stack(arg_stackptr)
			sizeVal := vm.Stack(arg_stackptr + 1)
			colorVal := vm.Stack(arg_stackptr + 2)

			if centerVal.Type != core.VAL_VEC3 || sizeVal.Type != core.VAL_VEC2 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("plane arguments must be vec3, vec2, vec4")
				return core.NIL_VALUE
			}

			centerObj := centerVal.Obj.(*core.Vec3Object)
			sizeObj := sizeVal.Obj.(*core.Vec2Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			center := rl.Vector3{X: float32(centerObj.X), Y: float32(centerObj.Y), Z: float32(centerObj.Z)}
			size := rl.Vector2{X: float32(sizeObj.X), Y: float32(sizeObj.Y)}
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			rl.DrawPlane(center, size, color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("ellipse3", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("ellipse expects 4 arguments: center(vec3), radiusX(number), radiusZ(number), color(vec4)")
				return core.NIL_VALUE
			}

			centerVal := vm.Stack(arg_stackptr)
			radiusXVal := vm.Stack(arg_stackptr + 1)
			radiusZVal := vm.Stack(arg_stackptr + 2)
			colorVal := vm.Stack(arg_stackptr + 3)

			if centerVal.Type != core.VAL_VEC3 || colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("ellipse arguments must be vec3, number, number, vec4")
				return core.NIL_VALUE
			}

			centerObj := centerVal.Obj.(*core.Vec3Object)
			colorObj := colorVal.Obj.(*core.Vec4Object)

			center := rl.Vector3{X: float32(centerObj.X), Y: float32(centerObj.Y), Z: float32(centerObj.Z)}
			radiusX := float32(radiusXVal.AsFloat())
			radiusZ := float32(radiusZVal.AsFloat())
			color := rl.NewColor(uint8(colorObj.X), uint8(colorObj.Y), uint8(colorObj.Z), uint8(colorObj.W))

			// Draw ellipse as a flattened cylinder
			rl.DrawCylinder(center, radiusX, radiusZ, 0.01, 16, color)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("triangle3", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 10 {
				vm.RunTimeError("triangle expects 10 arguments: x1, y1, z1, x2, y2, z2, x3, y3, z3, color")
				return core.NIL_VALUE
			}

			x1Val := vm.Stack(arg_stackptr)
			y1Val := vm.Stack(arg_stackptr + 1)
			z1Val := vm.Stack(arg_stackptr + 2)
			x2Val := vm.Stack(arg_stackptr + 3)
			y2Val := vm.Stack(arg_stackptr + 4)
			z2Val := vm.Stack(arg_stackptr + 5)
			x3Val := vm.Stack(arg_stackptr + 6)
			y3Val := vm.Stack(arg_stackptr + 7)
			z3Val := vm.Stack(arg_stackptr + 8)
			colVal := vm.Stack(arg_stackptr + 9)
			if colVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("Expected Vec4 for triangle color")
				return core.NIL_VALUE
			}
			v4obj := colVal.Obj.(*core.Vec4Object)
			r := uint8(v4obj.X)
			g := uint8(v4obj.Y)
			b := uint8(v4obj.Z)
			a := uint8(v4obj.W)
			x1 := float32(x1Val.AsFloat())
			y1 := float32(y1Val.AsFloat())
			z1 := float32(z1Val.AsFloat())
			x2 := float32(x2Val.AsFloat())
			y2 := float32(y2Val.AsFloat())
			z2 := float32(z2Val.AsFloat())
			x3 := float32(x3Val.AsFloat())
			y3 := float32(y3Val.AsFloat())
			z3 := float32(z3Val.AsFloat())
			rl.DrawTriangle3D(rl.Vector3{X: x1, Y: y1, Z: z1},
				rl.Vector3{X: x2, Y: y2, Z: z2},
				rl.Vector3{X: x3, Y: y3, Z: z3},
				rl.NewColor(r, g, b, a))
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("textured_cube", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("textured_cube expects 4 arguments: texture, position(vec3), size(vec3), base_color(vec4)")
				return core.NIL_VALUE
			}

			textureVal := vm.Stack(arg_stackptr)
			posVal := vm.Stack(arg_stackptr + 1)
			sizeVal := vm.Stack(arg_stackptr + 2)
			colorVal := vm.Stack(arg_stackptr + 3)

			// Extract texture object (can be either TextureObject or RenderTextureObject)
			if textureVal.Type != core.VAL_OBJ {
				vm.RunTimeError("textured_cube first argument must be a texture or render_texture object")
				return core.NIL_VALUE
			}

			var rayTexture rl.Texture2D
			if textureObj, ok := textureVal.Obj.(*TextureObject); ok {
				rayTexture = textureObj.Data.Texture
			} else if renderTextureObj, ok := textureVal.Obj.(*RenderTextureObject); ok {
				rayTexture = renderTextureObj.Data.RenderTexture.Texture
			} else {
				vm.RunTimeError("textured_cube first argument must be a texture or render_texture object")
				return core.NIL_VALUE
			}

			// Extract position vector
			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("textured_cube second argument must be a vec3")
				return core.NIL_VALUE
			}
			posObj := posVal.Obj.(*core.Vec3Object)

			// Extract size vector
			if sizeVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("textured_cube third argument must be a vec3")
				return core.NIL_VALUE
			}
			sizeObj := sizeVal.Obj.(*core.Vec3Object)

			// Extract base color
			if colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("textured_cube fourth argument must be a vec4")
				return core.NIL_VALUE
			}
			colorObj := colorVal.Obj.(*core.Vec4Object)

			// Create position and size vectors
			position := rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)}
			width := float32(sizeObj.X)
			height := float32(sizeObj.Y)
			length := float32(sizeObj.Z)

			// Create base color
			baseColor := rl.Color{
				R: uint8(colorObj.X),
				G: uint8(colorObj.Y),
				B: uint8(colorObj.Z),
				A: uint8(colorObj.W),
			}

			// Draw textured cube with base color tint
			rl.BeginBlendMode(rl.BlendAlpha)
			DrawTexturedCube(rayTexture, position, width, height, length, baseColor)
			rl.EndBlendMode()

			return core.NIL_VALUE
		},
	})
}

// DrawTexturedCube draws a cube with texture on all faces
func DrawTexturedCube(texture rl.Texture2D, position rl.Vector3, width, height, length float32, tint rl.Color) {
	x := position.X
	y := position.Y
	z := position.Z

	// Set texture and enable texturing
	rl.SetTexture(texture.ID)

	rl.Begin(rl.Quads)
	rl.Color4ub(tint.R, tint.G, tint.B, tint.A)

	// Front Face
	rl.Normal3f(0.0, 0.0, 1.0)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)

	// Back Face
	rl.Normal3f(0.0, 0.0, -1.0)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)

	// Top Face
	rl.Normal3f(0.0, 1.0, 0.0)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)

	// Bottom Face
	rl.Normal3f(0.0, -1.0, 0.0)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)

	// Right Face
	rl.Normal3f(1.0, 0.0, 0.0)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)

	// Left Face
	rl.Normal3f(-1.0, 0.0, 0.0)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)

	rl.End()

	// Disable texturing
	rl.SetTexture(0)
}

func RegisterAllWindowConstants(o *WindowObject) {
	// Batch type constants
	o.RegisterConstant("BATCH_CUBE", core.MakeIntValue(int(BATCH_CUBE), true))
	o.RegisterConstant("BATCH_SPHERE", core.MakeIntValue(int(BATCH_SPHERE), true))
	o.RegisterConstant("BATCH_PLANE", core.MakeIntValue(int(BATCH_PLANE), true))
	o.RegisterConstant("BATCH_TRIANGLE", core.MakeIntValue(int(BATCH_TRIANGLE), true))
	o.RegisterConstant("BATCH_TRIANGLE3", core.MakeIntValue(int(BATCH_TRIANGLE3), true))
	o.RegisterConstant("BATCH_TEXTURED_CUBE", core.MakeIntValue(int(BATCH_TEXTURED_CUBE), true))

	// Blend mode constants
	o.RegisterConstant("BLEND_ADD", core.MakeIntValue(int(rl.BlendAdditive), true))
	o.RegisterConstant("BLEND_ALPHA", core.MakeIntValue(int(rl.BlendAlpha), true))
	o.RegisterConstant("BLEND_MULTIPLY", core.MakeIntValue(int(rl.BlendMultiplied), true))
	o.RegisterConstant("BLEND_SUBCOLOR", core.MakeIntValue(int(rl.BlendSubtractColors), true))
	o.RegisterConstant("BLEND_ADDCOLOR", core.MakeIntValue(int(rl.BlendAddColors), true))

	o.RegisterConstant(("KEY_NULL"), core.MakeIntValue(int(rl.KeyNull), true))
	o.RegisterConstant(("KEY_SPACE"), core.MakeIntValue(int(rl.KeySpace), true))
	o.RegisterConstant(("KEY_ESCAPE"), core.MakeIntValue(int(rl.KeyEscape), true))
	o.RegisterConstant(("KEY_ENTER"), core.MakeIntValue(int(rl.KeyEnter), true))
	o.RegisterConstant(("KEY_TAB"), core.MakeIntValue(int(rl.KeyTab), true))
	o.RegisterConstant(("KEY_BACKSPACE"), core.MakeIntValue(int(rl.KeyBackspace), true))
	o.RegisterConstant(("KEY_INSERT"), core.MakeIntValue(int(rl.KeyInsert), true))
	o.RegisterConstant(("KEY_DELETE"), core.MakeIntValue(int(rl.KeyDelete), true))
	o.RegisterConstant(("KEY_RIGHT"), core.MakeIntValue(int(rl.KeyRight), true))
	o.RegisterConstant(("KEY_LEFT"), core.MakeIntValue(int(rl.KeyLeft), true))
	o.RegisterConstant(("KEY_DOWN"), core.MakeIntValue(int(rl.KeyDown), true))
	o.RegisterConstant(("KEY_UP"), core.MakeIntValue(int(rl.KeyUp), true))
	o.RegisterConstant(("KEY_PAGE_UP"), core.MakeIntValue(int(rl.KeyPageUp), true))
	o.RegisterConstant(("KEY_PAGE_DOWN"), core.MakeIntValue(int(rl.KeyPageDown), true))
	o.RegisterConstant(("KEY_HOME"), core.MakeIntValue(int(rl.KeyHome), true))
	o.RegisterConstant(("KEY_END"), core.MakeIntValue(int(rl.KeyEnd), true))
	o.RegisterConstant(("KEY_CAPS_LOCK"), core.MakeIntValue(int(rl.KeyCapsLock), true))
	o.RegisterConstant(("KEY_SCROLL_LOCK"), core.MakeIntValue(int(rl.KeyScrollLock), true))
	o.RegisterConstant(("KEY_NUM_LOCK"), core.MakeIntValue(int(rl.KeyNumLock), true))
	o.RegisterConstant(("KEY_PRINT_SCREEN"), core.MakeIntValue(int(rl.KeyPrintScreen), true))
	o.RegisterConstant(("KEY_PAUSE"), core.MakeIntValue(int(rl.KeyPause), true))
	o.RegisterConstant(("KEY_F1"), core.MakeIntValue(int(rl.KeyF1), true))
	o.RegisterConstant(("KEY_F2"), core.MakeIntValue(int(rl.KeyF2), true))
	o.RegisterConstant(("KEY_F3"), core.MakeIntValue(int(rl.KeyF3), true))
	o.RegisterConstant(("KEY_F4"), core.MakeIntValue(int(rl.KeyF4), true))
	o.RegisterConstant(("KEY_F5"), core.MakeIntValue(int(rl.KeyF5), true))
	o.RegisterConstant(("KEY_F6"), core.MakeIntValue(int(rl.KeyF6), true))
	o.RegisterConstant(("KEY_F7"), core.MakeIntValue(int(rl.KeyF7), true))
	o.RegisterConstant(("KEY_F8"), core.MakeIntValue(int(rl.KeyF8), true))
	o.RegisterConstant(("KEY_F9"), core.MakeIntValue(int(rl.KeyF9), true))
	o.RegisterConstant(("KEY_F10"), core.MakeIntValue(int(rl.KeyF10), true))
	o.RegisterConstant(("KEY_F11"), core.MakeIntValue(int(rl.KeyF11), true))
	o.RegisterConstant(("KEY_F12"), core.MakeIntValue(int(rl.KeyF12), true))
	o.RegisterConstant(("KEY_LEFT_SHIFT"), core.MakeIntValue(int(rl.KeyLeftShift), true))
	o.RegisterConstant(("KEY_LEFT_CONTROL"), core.MakeIntValue(int(rl.KeyLeftControl), true))
	o.RegisterConstant(("KEY_LEFT_ALT"), core.MakeIntValue(int(rl.KeyLeftAlt), true))
	o.RegisterConstant(("KEY_LEFT_SUPER"), core.MakeIntValue(int(rl.KeyLeftSuper), true))
	o.RegisterConstant(("KEY_RIGHT_SHIFT"), core.MakeIntValue(int(rl.KeyRightShift), true))
	o.RegisterConstant(("KEY_RIGHT_CONTROL"), core.MakeIntValue(int(rl.KeyRightControl), true))
	o.RegisterConstant(("KEY_RIGHT_ALT"), core.MakeIntValue(int(rl.KeyRightAlt), true))
	o.RegisterConstant(("KEY_RIGHT_SUPER"), core.MakeIntValue(int(rl.KeyRightSuper), true))
	o.RegisterConstant(("KEY_KB_MENU"), core.MakeIntValue(int(rl.KeyKbMenu), true))
	o.RegisterConstant(("KEY_LEFT_BRACKET"), core.MakeIntValue(int(rl.KeyLeftBracket), true))
	o.RegisterConstant(("KEY_BACK_SLASH"), core.MakeIntValue(int(rl.KeyBackSlash), true))
	o.RegisterConstant(("KEY_RIGHT_BRACKET"), core.MakeIntValue(int(rl.KeyRightBracket), true))
	o.RegisterConstant(("KEY_GRAVE"), core.MakeIntValue(int(rl.KeyGrave), true))

	// Keyboard Number Pad Keys
	o.RegisterConstant(("KEY_KP_0"), core.MakeIntValue(int(rl.KeyKp0), true))
	o.RegisterConstant(("KEY_KP_1"), core.MakeIntValue(int(rl.KeyKp1), true))
	o.RegisterConstant(("KEY_KP_2"), core.MakeIntValue(int(rl.KeyKp2), true))
	o.RegisterConstant(("KEY_KP_3"), core.MakeIntValue(int(rl.KeyKp3), true))
	o.RegisterConstant(("KEY_KP_4"), core.MakeIntValue(int(rl.KeyKp4), true))
	o.RegisterConstant(("KEY_KP_5"), core.MakeIntValue(int(rl.KeyKp5), true))
	o.RegisterConstant(("KEY_KP_6"), core.MakeIntValue(int(rl.KeyKp6), true))
	o.RegisterConstant(("KEY_KP_7"), core.MakeIntValue(int(rl.KeyKp7), true))
	o.RegisterConstant(("KEY_KP_8"), core.MakeIntValue(int(rl.KeyKp8), true))
	o.RegisterConstant(("KEY_KP_9"), core.MakeIntValue(int(rl.KeyKp9), true))
	o.RegisterConstant(("KEY_KP_DECIMAL"), core.MakeIntValue(int(rl.KeyKpDecimal), true))
	o.RegisterConstant(("KEY_KP_DIVIDE"), core.MakeIntValue(int(rl.KeyKpDivide), true))
	o.RegisterConstant(("KEY_KP_MULTIPLY"), core.MakeIntValue(int(rl.KeyKpMultiply), true))
	o.RegisterConstant(("KEY_KP_SUBTRACT"), core.MakeIntValue(int(rl.KeyKpSubtract), true))
	o.RegisterConstant(("KEY_KP_ADD"), core.MakeIntValue(int(rl.KeyKpAdd), true))
	o.RegisterConstant(("KEY_KP_ENTER"), core.MakeIntValue(int(rl.KeyKpEnter), true))
	o.RegisterConstant(("KEY_KP_EQUAL"), core.MakeIntValue(int(rl.KeyKpEqual), true))

	// Keyboard Alpha Numeric Keys
	o.RegisterConstant(("KEY_APOSTROPHE"), core.MakeIntValue(int(rl.KeyApostrophe), true))
	o.RegisterConstant(("KEY_COMMA"), core.MakeIntValue(int(rl.KeyComma), true))
	o.RegisterConstant(("KEY_MINUS"), core.MakeIntValue(int(rl.KeyMinus), true))
	o.RegisterConstant(("KEY_PERIOD"), core.MakeIntValue(int(rl.KeyPeriod), true))
	o.RegisterConstant(("KEY_SLASH"), core.MakeIntValue(int(rl.KeySlash), true))
	o.RegisterConstant(("KEY_ZERO"), core.MakeIntValue(int(rl.KeyZero), true))
	o.RegisterConstant(("KEY_ONE"), core.MakeIntValue(int(rl.KeyOne), true))
	o.RegisterConstant(("KEY_TWO"), core.MakeIntValue(int(rl.KeyTwo), true))
	o.RegisterConstant(("KEY_THREE"), core.MakeIntValue(int(rl.KeyThree), true))
	o.RegisterConstant(("KEY_FOUR"), core.MakeIntValue(int(rl.KeyFour), true))
	o.RegisterConstant(("KEY_FIVE"), core.MakeIntValue(int(rl.KeyFive), true))
	o.RegisterConstant(("KEY_SIX"), core.MakeIntValue(int(rl.KeySix), true))
	o.RegisterConstant(("KEY_SEVEN"), core.MakeIntValue(int(rl.KeySeven), true))
	o.RegisterConstant(("KEY_EIGHT"), core.MakeIntValue(int(rl.KeyEight), true))
	o.RegisterConstant(("KEY_NINE"), core.MakeIntValue(int(rl.KeyNine), true))
	o.RegisterConstant(("KEY_SEMICOLON"), core.MakeIntValue(int(rl.KeySemicolon), true))
	o.RegisterConstant(("KEY_EQUAL"), core.MakeIntValue(int(rl.KeyEqual), true))
	o.RegisterConstant(("KEY_A"), core.MakeIntValue(int(rl.KeyA), true))
	o.RegisterConstant(("KEY_B"), core.MakeIntValue(int(rl.KeyB), true))
	o.RegisterConstant(("KEY_C"), core.MakeIntValue(int(rl.KeyC), true))
	o.RegisterConstant(("KEY_D"), core.MakeIntValue(int(rl.KeyD), true))
	o.RegisterConstant(("KEY_E"), core.MakeIntValue(int(rl.KeyE), true))
	o.RegisterConstant(("KEY_F"), core.MakeIntValue(int(rl.KeyF), true))
	o.RegisterConstant(("KEY_G"), core.MakeIntValue(int(rl.KeyG), true))
	o.RegisterConstant(("KEY_H"), core.MakeIntValue(int(rl.KeyH), true))
	o.RegisterConstant(("KEY_I"), core.MakeIntValue(int(rl.KeyI), true))
	o.RegisterConstant(("KEY_J"), core.MakeIntValue(int(rl.KeyJ), true))
	o.RegisterConstant(("KEY_K"), core.MakeIntValue(int(rl.KeyK), true))
	o.RegisterConstant(("KEY_L"), core.MakeIntValue(int(rl.KeyL), true))
	o.RegisterConstant(("KEY_M"), core.MakeIntValue(int(rl.KeyM), true))
	o.RegisterConstant(("KEY_N"), core.MakeIntValue(int(rl.KeyN), true))
	o.RegisterConstant(("KEY_O"), core.MakeIntValue(int(rl.KeyO), true))
	o.RegisterConstant(("KEY_P"), core.MakeIntValue(int(rl.KeyP), true))
	o.RegisterConstant(("KEY_Q"), core.MakeIntValue(int(rl.KeyQ), true))
	o.RegisterConstant(("KEY_R"), core.MakeIntValue(int(rl.KeyR), true))
	o.RegisterConstant(("KEY_S"), core.MakeIntValue(int(rl.KeyS), true))
	o.RegisterConstant(("KEY_T"), core.MakeIntValue(int(rl.KeyT), true))
	o.RegisterConstant(("KEY_U"), core.MakeIntValue(int(rl.KeyU), true))
	o.RegisterConstant(("KEY_V"), core.MakeIntValue(int(rl.KeyV), true))
	o.RegisterConstant(("KEY_W"), core.MakeIntValue(int(rl.KeyW), true))
	o.RegisterConstant(("KEY_X"), core.MakeIntValue(int(rl.KeyX), true))
	o.RegisterConstant(("KEY_Y"), core.MakeIntValue(int(rl.KeyY), true))
	o.RegisterConstant(("KEY_Z"), core.MakeIntValue(int(rl.KeyZ), true))

	// Android keys
	o.RegisterConstant(("KEY_BACK"), core.MakeIntValue(int(rl.KeyBack), true))
	o.RegisterConstant(("KEY_MENU"), core.MakeIntValue(int(rl.KeyMenu), true))
	o.RegisterConstant(("KEY_VOLUME_UP"), core.MakeIntValue(int(rl.KeyVolumeUp), true))
	o.RegisterConstant(("KEY_VOLUME_DOWN"), core.MakeIntValue(int(rl.KeyVolumeDown), true))
}
