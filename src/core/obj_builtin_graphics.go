package core

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Graphics struct {
	Width, Height int32
	Blend_mode    rl.BlendMode
}

func (g *Graphics) SetBlendMode(modename string) {
	switch modename {
	case "add":
		g.Blend_mode = rl.BlendAdditive
	case "alpha":
		g.Blend_mode = rl.BlendAlpha
	case "multiply":
		g.Blend_mode = rl.BlendMultiplied
	case "subtract":
		g.Blend_mode = rl.BlendSubtractColors
	default:
		g.Blend_mode = rl.BlendAlpha // default to alpha blending
	}
}

type GraphicsObject struct {
	BuiltInObject
	Value   *Graphics
	Methods map[string]*BuiltInObject
}

func MakeGraphicsObject(w int, h int) *GraphicsObject {

	rv := &GraphicsObject{
		BuiltInObject: BuiltInObject{},
		Value:         &Graphics{Width: int32(w), Height: int32(h), Blend_mode: rl.BlendAlpha},
	}

	return rv
}

func (o *GraphicsObject) String() string {
	return fmt.Sprintf("<Graphics %dx%d>", o.Value.Width, o.Value.Height)
}

func (o *GraphicsObject) GetType() ObjectType {
	return OBJECT_GRAPHICS
}

func (o *GraphicsObject) GetMethod(name string) *BuiltInObject {
	return o.Methods[name]
}
func (o *GraphicsObject) RegisterMethod(name string, method *BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*BuiltInObject)
	}
	o.Methods[name] = method
}

// -------------------------------------------------------------------------------------------
func (t *GraphicsObject) IsBuiltIn() bool {
	return true
}
