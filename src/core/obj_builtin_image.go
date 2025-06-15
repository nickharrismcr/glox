package core

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Image struct {
	Width, Height int32
	Image         *rl.Image
}

type ImageObject struct {
	BuiltInObject
	Data    *Image
	Methods map[string]*BuiltInObject
}

func MakeImageObject(filename string) *ImageObject {

	img := rl.LoadImage(filename)
	if img.Data == nil {
		panic(fmt.Sprintf("Failed to load image from %s", filename))
	}

	w := img.Width
	h := img.Height
	rv := &ImageObject{
		BuiltInObject: BuiltInObject{},
		Data: &Image{Width: int32(w),
			Height: int32(h),
			Image:  img,
		},
	}
	rv.RegisterAllMethods()
	return rv
}

func (o *ImageObject) String() string {
	return fmt.Sprintf("<Image %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *ImageObject) GetType() ObjectType {
	return OBJECT_IMAGE
}

func (o *ImageObject) GetMethod(name string) *BuiltInObject {
	return o.Methods[name]
}
func (o *ImageObject) RegisterMethod(name string, method *BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*BuiltInObject)
	}
	o.Methods[name] = method
}
func (o *ImageObject) RegisterAllMethods() {

	o.RegisterMethod("width", &BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			return MakeIntValue(int(o.Data.Width), true)
		},
	})
	o.RegisterMethod("height", &BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			return MakeIntValue(int(o.Data.Height), true)
		},
	})

}

func (t *ImageObject) IsBuiltIn() bool {
	return true
}
