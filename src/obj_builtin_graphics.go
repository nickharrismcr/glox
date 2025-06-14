package lox

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Graphics struct {
	width, height int32
	blend_mode    rl.BlendMode
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

type GraphicsObject struct {
	BuiltInObject
	value   *Graphics
	methods map[string]*BuiltInObject
}

func makeGraphicsObject(w int, h int) *GraphicsObject {

	rv := &GraphicsObject{
		BuiltInObject: BuiltInObject{},
		value:         &Graphics{width: int32(w), height: int32(h), blend_mode: rl.BlendAlpha},
	}
	rv.RegisterAllMethods()
	return rv
}

func (o *GraphicsObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.value.width, o.value.height)
}

func (o *GraphicsObject) getType() ObjectType {
	return OBJECT_GRAPHICS
}

func (o *GraphicsObject) GetMethod(name string) *BuiltInObject {
	return o.methods[name]
}
func (o *GraphicsObject) RegisterMethod(name string, method *BuiltInObject) {
	if o.methods == nil {
		o.methods = make(map[string]*BuiltInObject)
	}
	o.methods[name] = method
}
func (o *GraphicsObject) RegisterAllMethods() {

	o.RegisterMethod("init", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.SetTraceLogLevel(rl.LogNone)
			rl.InitWindow(o.value.width, o.value.height, "GLOX")
			return makeNilValue()
		},
	})
	o.RegisterMethod("begin", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.BeginDrawing()
			rl.BeginBlendMode(rl.BlendAdditive)
			return makeNilValue()
		},
	})
	o.RegisterMethod("begin_blend_mode", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			modeVal := vm.Stack(arg_stackptr)
			o.value.setBlendMode(modeVal.asString().get())
			rl.BeginBlendMode(o.value.blend_mode)
			return makeNilValue()
		},
	})
	o.RegisterMethod("end_blend_mode", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.EndBlendMode()
			return makeNilValue()
		},
	})
	o.RegisterMethod("end", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.DrawFPS(10, 10)
			rl.EndDrawing()
			return makeNilValue()
		},
	})
	o.RegisterMethod("clear", &BuiltInObject{
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
	})
	o.RegisterMethod("line", &BuiltInObject{
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
	})
	o.RegisterMethod("circle_fill", &BuiltInObject{
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
	})
	o.RegisterMethod("circle", &BuiltInObject{
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
	})
	o.RegisterMethod("text", &BuiltInObject{
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
	})
	o.RegisterMethod("draw_array", &BuiltInObject{
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
	})

	o.RegisterMethod("should_close", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			return makeBooleanValue(rl.WindowShouldClose(), true)
		},
	})
	o.RegisterMethod("close", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.CloseWindow()
			return makeNilValue()
		},
	})

}

//-------------------------------------------------------------------------------------------
