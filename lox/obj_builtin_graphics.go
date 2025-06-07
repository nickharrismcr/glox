package lox

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Graphics struct {
	width, height int32
}

type GraphicsObject struct {
	BuiltInObject
	value    *Graphics
	drawFunc *ClosureObject
}

func makeGraphicsObject(w int, h int) *GraphicsObject {

	return &GraphicsObject{
		BuiltInObject: BuiltInObject{},
		value:         &Graphics{width: int32(w), height: int32(h)},
	}
}

func (o *GraphicsObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.value.width, o.value.height)
}

func (o *GraphicsObject) getType() ObjectType {
	return OBJECT_GRAPHICS
}

func (o *GraphicsObject) GetMethod(name string) *BuiltInObject {
	switch name {
	case "init":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				rl.SetTraceLogLevel(rl.LogNone)
				rl.InitWindow(o.value.width, o.value.height, "GLOX")
				return makeNilValue()
			},
		}
	case "begin":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				rl.BeginDrawing()
				return makeNilValue()
			},
		}
	case "end":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {

				rl.DrawFPS(10, 10)
				rl.EndDrawing()
				return makeNilValue()
			},
		}
	case "clear":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				rval := vm.stack[arg_stackptr]
				gval := vm.stack[arg_stackptr+1]
				bval := vm.stack[arg_stackptr+2]
				aval := vm.stack[arg_stackptr+3]
				r := rval.asInt()
				g := gval.asInt()
				b := bval.asInt()
				a := aval.asInt()
				rl.ClearBackground(rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)))
				return makeNilValue()
			},
		}
	case "circle_fill":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				xval := vm.stack[arg_stackptr]
				yval := vm.stack[arg_stackptr+1]
				radVal := vm.stack[arg_stackptr+2]
				rval := vm.stack[arg_stackptr+3]
				gval := vm.stack[arg_stackptr+4]
				bval := vm.stack[arg_stackptr+5]
				aval := vm.stack[arg_stackptr+6]

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
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				xval := vm.stack[arg_stackptr]
				yval := vm.stack[arg_stackptr+1]
				radVal := vm.stack[arg_stackptr+2]
				rval := vm.stack[arg_stackptr+3]
				gval := vm.stack[arg_stackptr+4]
				bval := vm.stack[arg_stackptr+5]
				aval := vm.stack[arg_stackptr+6]

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
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				xval := vm.stack[arg_stackptr]
				yval := vm.stack[arg_stackptr+1]
				sval := vm.stack[arg_stackptr+2]

				x := int32(xval.asInt())
				y := int32(yval.asInt())
				s := sval.asString().get()

				rl.DrawText(s, x, y, 120, rl.Black)
				return makeNilValue()
			},
		}
	case "draw_array":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				arrVal := vm.stack[arg_stackptr]
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
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				return makeBooleanValue(rl.WindowShouldClose(), true)
			},
		}
	case "close":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				rl.CloseWindow()
				return makeNilValue()
			},
		}

	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
