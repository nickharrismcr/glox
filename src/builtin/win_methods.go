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

func RegisterAllWindowConstants(o *WindowObject) {
	o.RegisterConstant("BLEND_ADD", core.MakeIntValue(int(rl.BlendAdditive), true))
	o.RegisterConstant("BLEND_ALPHA", core.MakeIntValue(int(rl.BlendAlpha), true))
	o.RegisterConstant("BLEND_MULTIPLY", core.MakeIntValue(int(rl.BlendMultiplied), true))
	o.RegisterConstant("BLEND_SUBTRACT", core.MakeIntValue(int(rl.BlendSubtractColors), true))
	o.RegisterConstant("BLEND_DEFAULT", core.MakeIntValue(int(rl.BlendAlpha), true)) // default blend mode

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
