package lox

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Graphics struct {
	width, height int32
	blend_mode    rl.BlendMode
}

type GraphicsObject struct {
	BuiltInObject
	value *Graphics
}

func makeGraphicsObject(w int, h int) *GraphicsObject {

	return &GraphicsObject{
		BuiltInObject: BuiltInObject{},
		value:         &Graphics{width: int32(w), height: int32(h), blend_mode: rl.BlendAlpha},
	}
}

func (o *GraphicsObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.value.width, o.value.height)
}

func (g *Graphics) setBlendMode(modename string) {
	switch modename {
	case "add":
		g.blend_mode = rl.BlendAdditive
	case "alpha":
		g.blend_mode = rl.BlendAlpha
	case "multiply":
		g.blend_mode = rl.BlendMultiplied
	case "subtract":
		g.blend_mode = rl.BlendSubtractColors
	default:
		g.blend_mode = rl.BlendAlpha // default to alpha blending
	}
}

func (o *GraphicsObject) getType() ObjectType {
	return OBJECT_GRAPHICS
}

func (o *GraphicsObject) GetMethod(name string) *BuiltInObject {
	switch name {
	case "init":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				rl.SetTraceLogLevel(rl.LogNone)
				rl.InitWindow(o.value.width, o.value.height, "GLOX")
				return makeNilValue()
			},
		}
	case "begin":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				rl.BeginDrawing()
				rl.BeginBlendMode(rl.BlendAdditive)
				return makeNilValue()
			},
		}
	case "begin_blend_mode":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				modeVal := vm.Stack(arg_stackptr)

				o.value.setBlendMode(modeVal.asString().get())
				rl.BeginBlendMode(o.value.blend_mode)
				return makeNilValue()
			},
		}
	case "end_blend_mode":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {

				rl.EndBlendMode()
				return makeNilValue()
			},
		}
	case "end":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {

				rl.DrawFPS(10, 10)
				rl.EndDrawing()
				return makeNilValue()
			},
		}
	case "clear":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				rval := vm.Stack(arg_stackptr)
				gval := vm.Stack(arg_stackptr + 1)
				bval := vm.Stack(arg_stackptr + 2)
				aval := vm.Stack(arg_stackptr + 3)
				r := rval.asInt()
				g := gval.asInt()
				b := bval.asInt()
				a := aval.asInt()
				rl.ClearBackground(rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
				return makeNilValue()
			},
		}
	case "line":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				x1val := vm.Stack(arg_stackptr)
				y1val := vm.Stack(arg_stackptr + 1)
				x2val := vm.Stack(arg_stackptr + 2)
				y2val := vm.Stack(arg_stackptr + 3)
				rval := vm.Stack(arg_stackptr + 4)
				gval := vm.Stack(arg_stackptr + 5)
				bval := vm.Stack(arg_stackptr + 6)
				aval := vm.Stack(arg_stackptr + 7)

				x1 := int32(x1val.asInt())
				y1 := int32(y1val.asInt())
				x2 := int32(x2val.asInt())
				y2 := int32(y2val.asInt())
				r := int32(rval.asInt())
				g := int32(gval.asInt())
				b := int32(bval.asInt())
				a := int32(aval.asInt())

				rl.DrawLine(x1, y1, x2, y2, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
				return makeNilValue()
			},
		}
	case "circle_fill":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				radVal := vm.Stack(arg_stackptr + 2)
				rval := vm.Stack(arg_stackptr + 3)
				gval := vm.Stack(arg_stackptr + 4)
				bval := vm.Stack(arg_stackptr + 5)
				aval := vm.Stack(arg_stackptr + 6)

				x := int32(xval.asInt())
				y := int32(yval.asInt())
				rad := float32(radVal.asInt())
				r := int32(rval.asInt())
				g := int32(gval.asInt())
				b := int32(bval.asInt())
				a := int32(aval.asInt())

				rl.DrawCircle(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
				return makeNilValue()
			},
		}
	case "circle":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				radVal := vm.Stack(arg_stackptr + 2)
				rval := vm.Stack(arg_stackptr + 3)
				gval := vm.Stack(arg_stackptr + 4)
				bval := vm.Stack(arg_stackptr + 5)
				aval := vm.Stack(arg_stackptr + 6)

				x := int32(xval.asInt())
				y := int32(yval.asInt())
				rad := float32(radVal.asInt())
				r := int32(rval.asInt())
				g := int32(gval.asInt())
				b := int32(bval.asInt())
				a := int32(aval.asInt())

				rl.DrawCircleLines(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
				return makeNilValue()
			},
		}
	case "text":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				sval := vm.Stack(arg_stackptr + 2)

				x := int32(xval.asInt())
				y := int32(yval.asInt())
				s := sval.asString().get()

				rl.DrawText(s, x, y, 120, rl.Black)
				return makeNilValue()
			},
		}
	case "draw_array":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				arrVal := vm.Stack(arg_stackptr)
				arrobj := arrVal.asFloatArray()
				arr := arrobj.value
				for x := range arr.width {
					for y := range arr.height {
						f := arr.get(x, y)
						r, g, b := DecodeRGB(f)
						col := rl.NewColor(r, g, b, 255)
						rl.DrawPixel(int32(x), int32(y), col)
					}
				}

				return makeNilValue()
			},
		}

	case "should_close":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				return makeBooleanValue(rl.WindowShouldClose(), true)
			},
		}
	case "close":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				rl.CloseWindow()
				return makeNilValue()
			},
		}

	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
