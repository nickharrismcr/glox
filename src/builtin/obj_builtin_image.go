package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Image struct {
	Width, Height int32
	Image         *rl.Image
}

type ImageObject struct {
	core.BuiltInObject
	Data    *Image
	Methods map[string]*core.BuiltInObject
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

func (o *ImageObject) String() string {
	return fmt.Sprintf("<Image %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *ImageObject) GetType() core.ObjectType {
	return core.OBJECT_IMAGE
}

func (o *ImageObject) GetMethod(name string) *core.BuiltInObject {
	return o.Methods[name]
}
func (o *ImageObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*core.BuiltInObject)
	}
	o.Methods[name] = method
}

func (t *ImageObject) IsBuiltIn() bool {
	return true
}
