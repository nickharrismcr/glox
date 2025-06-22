package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func WindowBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("window expects 2 arguments")
		return core.NIL_VALUE
	}
	wVal := vm.Stack(arg_stackptr)
	hVal := vm.Stack(arg_stackptr + 1)
	if !wVal.IsInt() || !hVal.IsInt() {
		vm.RunTimeError("window arguments must be integers")
		return core.NIL_VALUE
	}
	o := MakeWindowObject(wVal.Int, hVal.Int)
	RegisterAllWindowMethods(o)
	RegisterAllWindowConstants(o)
	return core.MakeObjectValue(o, true)
}

type Graphics struct {
	Width, Height int32
	Blend_mode    rl.BlendMode
}

func (g *Graphics) SetBlendMode(mode int) {
	g.Blend_mode = (rl.BlendMode)(mode)

}

type WindowObject struct {
	core.BuiltInObject
	Value     *Graphics
	Methods   map[int]*core.BuiltInObject
	Constants map[int]core.Value
}

func MakeWindowObject(w int, h int) *WindowObject {

	rv := &WindowObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         &Graphics{Width: int32(w), Height: int32(h), Blend_mode: rl.BlendAlpha},
	}

	return rv
}

func (o *WindowObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.Value.Width, o.Value.Height)
}

func (o *WindowObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *WindowObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}
func (o *WindowObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *WindowObject) GetConstant(stringId int) core.Value {
	rv, ok := o.Constants[stringId]
	if !ok {
		return core.NIL_VALUE
	}
	return rv
}

func (o *WindowObject) RegisterConstant(name string, value core.Value) {
	if o.Constants == nil {
		o.Constants = make(map[int]core.Value)
	}
	o.Constants[core.InternName(name)] = value
}

// -------------------------------------------------------------------------------------------
func (t *WindowObject) IsBuiltIn() bool {
	return true
}
