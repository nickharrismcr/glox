package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func GraphicsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("graphics expects 2 arguments")
		return core.MakeNilValue()
	}
	wVal := vm.Stack(arg_stackptr)
	hVal := vm.Stack(arg_stackptr + 1)
	if !wVal.IsInt() || !hVal.IsInt() {
		vm.RunTimeError("graphics arguments must be integers")
		return core.MakeNilValue()
	}
	o := MakeGraphicsObject(wVal.Int, hVal.Int)
	RegisterAllGraphicsMethods(o)
	RegisterAllGraphicsConstants(o)
	return core.MakeObjectValue(o, true)
}

type Graphics struct {
	Width, Height int32
	Blend_mode    rl.BlendMode
}

func (g *Graphics) SetBlendMode(modename string) {
	switch modename {
	case "BLEND_ADD":
		g.Blend_mode = rl.BlendAdditive
	case "BLEND_ALPHA":
		g.Blend_mode = rl.BlendAlpha
	case "BLEND_MULTIPLY":
		g.Blend_mode = rl.BlendMultiplied
	case "BLEND_SUBTRACT":
		g.Blend_mode = rl.BlendSubtractColors
	default:
		g.Blend_mode = rl.BlendAlpha // default to alpha blending
	}
}

type GraphicsObject struct {
	core.BuiltInObject
	Value     *Graphics
	Methods   map[string]*core.BuiltInObject
	Constants map[string]core.Value
}

func MakeGraphicsObject(w int, h int) *GraphicsObject {

	rv := &GraphicsObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         &Graphics{Width: int32(w), Height: int32(h), Blend_mode: rl.BlendAlpha},
	}

	return rv
}

func (o *GraphicsObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.Value.Width, o.Value.Height)
}

func (o *GraphicsObject) GetType() core.ObjectType {
	return core.OBJECT_GRAPHICS
}

func (o *GraphicsObject) GetMethod(name string) *core.BuiltInObject {
	return o.Methods[name]
}
func (o *GraphicsObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*core.BuiltInObject)
	}
	o.Methods[name] = method
}

func (o *GraphicsObject) GetConstant(name string) core.Value {
	rv, ok := o.Constants[name]
	if !ok {
		return core.MakeNilValue()
	}
	return rv
}

func (o *GraphicsObject) RegisterConstant(name string, value core.Value) {
	if o.Constants == nil {
		o.Constants = make(map[string]core.Value)
	}
	o.Constants[name] = value
}

// -------------------------------------------------------------------------------------------
func (t *GraphicsObject) IsBuiltIn() bool {
	return true
}
