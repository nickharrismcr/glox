package core

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Width, Height int32
	Image         *rl.Image
	Texture       rl.Texture2D
	Frames        int
	FrameWidth    int
	StartFrame    int            // Starting frame for animation
	EndFrame      int            // Ending frame for animation
	FrameRects    []rl.Rectangle // Rectangles for each frame in animation
	TicksPerFrame int            // Ticks per frame for animation 0 means no animation
	Ticks         int            // Ticks since last frame change
	CurrentFrame  int            // Current frame index for animation
}

func (t *Texture) GetFrameRect() rl.Rectangle {

	return t.FrameRects[t.CurrentFrame]
}

func (t *Texture) Animate() {
	if t.TicksPerFrame == 0 {
		return
	}
	t.Ticks++
	if t.Ticks == t.TicksPerFrame {
		t.Ticks = 0
		t.CurrentFrame = (t.CurrentFrame + 1) % t.Frames
	}
}

type TextureObject struct {
	BuiltInObject
	Data    Texture
	Methods map[string]*BuiltInObject
}

func MakeTextureObject(image *rl.Image, frames int, startFrame int, endFrame int) *TextureObject {

	texture := rl.LoadTextureFromImage(image)
	LogFmt(INFO, "Loaded texture from image")

	w := texture.Width
	h := texture.Height
	data := Texture{
		Width:         int32(w),
		Height:        int32(h),
		Image:         image,
		Texture:       texture,
		Frames:        frames,
		FrameWidth:    int(w) / frames,
		StartFrame:    startFrame,
		EndFrame:      endFrame,
		FrameRects:    make([]rl.Rectangle, 0, frames),
		TicksPerFrame: 0, // Default to no animation
		Ticks:         0, // Ticks since last frame change
		CurrentFrame:  0, // Start at the first frame
	}
	for f := startFrame; f <= endFrame; f++ {
		x1 := float32((f - 1) * data.FrameWidth)
		rect := rl.NewRectangle(x1, 0, float32(data.FrameWidth), float32(h))
		data.FrameRects = append(data.FrameRects, rect)
	}

	rv := &TextureObject{
		BuiltInObject: BuiltInObject{},
		Data:          data,
	}
	rv.RegisterAllMethods()
	return rv
}

func (o *TextureObject) String() string {
	return fmt.Sprintf("<Texture %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *TextureObject) GetType() ObjectType {
	return OBJECT_TEXTURE
}

func (o *TextureObject) GetMethod(name string) *BuiltInObject {
	return o.Methods[name]
}
func (o *TextureObject) RegisterMethod(name string, method *BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*BuiltInObject)
	}
	o.Methods[name] = method
}
func (o *TextureObject) RegisterAllMethods() {

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
	o.RegisterMethod("animate", &BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
			if argCount < 1 {
				vm.RunTimeError("animate requires at least one argument")

			}
			ticksVal := vm.Stack(arg_stackptr)
			if !ticksVal.IsNumber() {
				vm.RunTimeError("animate requires a number argument for ticks per frame")
				return MakeNilValue()
			}
			o.Data.TicksPerFrame = ticksVal.Int
			return MakeNilValue()
		},
	})
}

func (t *TextureObject) IsBuiltIn() bool {
	return true
}
