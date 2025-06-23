package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RenderTextureBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 2 {
		vm.RunTimeError("texture expects 2 arguments (width,height))")
		return core.NIL_VALUE
	}
	widthVal := vm.Stack(arg_stackptr)
	heightVal := vm.Stack(arg_stackptr + 1)
	width := widthVal.Int
	height := heightVal.Int

	o := MakeRenderTextureObject(width, height)
	RegisterAllRenderTextureMethods(o)
	return core.MakeObjectValue(o, true)
}

type RenderTexture struct {
	Width, Height int32
	RenderTexture rl.RenderTexture2D
}

type RenderTextureObject struct {
	core.BuiltInObject
	Data    RenderTexture
	Methods map[int]*core.BuiltInObject
}

func MakeRenderTextureObject(width int, height int) *RenderTextureObject {

	texture := rl.LoadRenderTexture(int32(width), int32(height))

	data := RenderTexture{
		Width:         int32(width),
		Height:        int32(height),
		RenderTexture: texture,
	}

	rv := &RenderTextureObject{
		BuiltInObject: core.BuiltInObject{},
		Data:          data,
	}

	return rv
}

func (o *RenderTextureObject) String() string {
	return fmt.Sprintf("<RenderTexture %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *RenderTextureObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (tex *RenderTextureObject) GetNativeType() core.NativeType {
	return core.NATIVE_RENDER_TEXTURE
}

func (o *RenderTextureObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}
func (o *RenderTextureObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (t *RenderTextureObject) IsBuiltIn() bool {
	return true
}
