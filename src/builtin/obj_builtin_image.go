package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func ImageBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 1 {
		vm.RunTimeError("image expects 1 argument")
		return core.NIL_VALUE
	}
	filenameVal := vm.Stack(arg_stackptr)
	if !filenameVal.IsStringObject() {
		vm.RunTimeError("image argument must be a string")
		return core.NIL_VALUE
	}
	o := MakeImageObject(filenameVal.AsString().Get())
	RegisterAllImageMethods(o)
	return core.MakeObjectValue(o, true)
}

type Image struct {
	Width, Height int32
	Image         *rl.Image
}

type ImageObject struct {
	core.BuiltInObject
	Data    *Image
	Methods map[int]*core.BuiltInObject
}

func MakeImageObject(filename string) *ImageObject {

	img := rl.LoadImage(filename)
	if img.Data == nil {
		panic(fmt.Sprintf("Failed to load image from %s", filename))
	}

	w := img.Width
	h := img.Height
	rv := &ImageObject{
		BuiltInObject: core.BuiltInObject{},
		Data: &Image{Width: int32(w),
			Height: int32(h),
			Image:  img,
		},
	}

	return rv
}
func (img *ImageObject) GetNativeType() core.NativeType {
	return core.NATIVE_IMAGE
}

func (o *ImageObject) String() string {
	return fmt.Sprintf("<Image %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *ImageObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *ImageObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}
func (o *ImageObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (t *ImageObject) IsBuiltIn() bool {
	return true
}
