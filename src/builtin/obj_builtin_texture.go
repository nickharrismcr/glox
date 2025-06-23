package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TextureBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 4 {
		vm.RunTimeError("texture expects 4 arguments (image, frames, start_frame, end_frame)")
		return core.NIL_VALUE
	}
	imgVal := vm.Stack(arg_stackptr)
	framesVal := vm.Stack(arg_stackptr + 1)
	startFrameVal := vm.Stack(arg_stackptr + 2)
	endFrameVal := vm.Stack(arg_stackptr + 3)

	var to *ImageObject
	to, ok := imgVal.Obj.(*ImageObject)
	if !ok {
		vm.RunTimeError("texture argument must be an image object")
		return core.NIL_VALUE
	}
	frames := framesVal.Int
	if frames < 1 {
		vm.RunTimeError("texture frames must be at least 1")
		return core.NIL_VALUE
	}
	startFrame := startFrameVal.Int
	if startFrame < 1 || startFrame > frames {
		vm.RunTimeError("texture start_frame must be between 1 and frames")
		return core.NIL_VALUE
	}
	endFrame := endFrameVal.Int
	if endFrame < 1 || endFrame > frames {
		vm.RunTimeError("texture end_frame must be between 1 and frames")
		return core.NIL_VALUE
	}
	o := MakeTextureObject(to.Data.Image, frames, startFrame, endFrame)
	RegisterAllTextureMethods(o)
	return core.MakeObjectValue(o, true)
}

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
	core.BuiltInObject
	Data    Texture
	Methods map[int]*core.BuiltInObject
}

func MakeTextureObject(image *rl.Image, frames int, startFrame int, endFrame int) *TextureObject {

	texture := rl.LoadTextureFromImage(image)
	core.LogFmt(core.INFO, "Loaded texture from image")

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
		BuiltInObject: core.BuiltInObject{},
		Data:          data,
	}

	return rv
}

func (tex *TextureObject) GetNativeType() core.NativeType {
	return core.NATIVE_TEXTURE
}

func (o *TextureObject) String() string {
	return fmt.Sprintf("<Texture %dx%d>", o.Data.Width, o.Data.Height)
}

func (o *TextureObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *TextureObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}
func (o *TextureObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (t *TextureObject) IsBuiltIn() bool {
	return true
}
