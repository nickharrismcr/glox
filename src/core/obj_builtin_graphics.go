package core

import (
	"fmt"
	"glox/src/util"

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

func MakeGraphicsObject(w int, h int) *GraphicsObject {

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

func (o *GraphicsObject) GetType() ObjectType {
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
			return MakeNilValue()
		},
	})
	o.RegisterMethod("begin", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.BeginDrawing()
			rl.BeginBlendMode(rl.BlendAdditive)
			return MakeNilValue()
		},
	})
	o.RegisterMethod("begin_blend_mode", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			modeVal := vm.Stack(arg_stackptr)
			o.value.setBlendMode(modeVal.AsString().Get())
			rl.BeginBlendMode(o.value.blend_mode)
			return MakeNilValue()
		},
	})
	o.RegisterMethod("end_blend_mode", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.EndBlendMode()
			return MakeNilValue()
		},
	})
	o.RegisterMethod("end", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.DrawFPS(10, 10)
			rl.EndDrawing()
			return MakeNilValue()
		},
	})
	o.RegisterMethod("clear", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rval := vm.Stack(arg_stackptr)
			gval := vm.Stack(arg_stackptr + 1)
			bval := vm.Stack(arg_stackptr + 2)
			aval := vm.Stack(arg_stackptr + 3)
			r := rval.AsInt()
			g := gval.AsInt()
			b := bval.AsInt()
			a := aval.AsInt()
			rl.ClearBackground(rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			return MakeNilValue()
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

			x1 := int32(x1val.AsInt())
			y1 := int32(y1val.AsInt())
			x2 := int32(x2val.AsInt())
			y2 := int32(y2val.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.DrawLine(x1, y1, x2, y2, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			return MakeNilValue()
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

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			rad := float32(radVal.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.DrawCircle(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			return MakeNilValue()
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

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			rad := float32(radVal.AsInt())
			r := int32(rval.AsInt())
			g := int32(gval.AsInt())
			b := int32(bval.AsInt())
			a := int32(aval.AsInt())

			rl.DrawCircleLines(x, y, rad, rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
			return MakeNilValue()
		},
	})
	o.RegisterMethod("text", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			xval := vm.Stack(arg_stackptr)
			yval := vm.Stack(arg_stackptr + 1)
			sval := vm.Stack(arg_stackptr + 2)

			x := int32(xval.AsInt())
			y := int32(yval.AsInt())
			s := sval.AsString().Get()

			rl.DrawText(s, x, y, 10, rl.White)
			return MakeNilValue()
		},
	})
	o.RegisterMethod("draw_array", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			arrVal := vm.Stack(arg_stackptr)
			arrobj := arrVal.AsFloatArray()
			arr := arrobj.Value

			for x := range arr.Width {
				for y := range arr.Height {
					f := arr.Get(x, y)
					r, g, b := util.DecodeRGB(f)
					col := rl.NewColor(r, g, b, 255)
					rl.DrawPixel(int32(x), int32(y), col)
				}
			}

			return MakeNilValue()
		},
	})

	o.RegisterMethod("should_close", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			return MakeBooleanValue(rl.WindowShouldClose(), true)
		},
	})
	o.RegisterMethod("close", &BuiltInObject{
		function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			rl.CloseWindow()
			return MakeNilValue()
		},
	})

}

//-------------------------------------------------------------------------------------------
