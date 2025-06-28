package builtin

import (
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func CameraBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 {
		vm.RunTimeError("camera expects 3 arguments: position(vec3), target(vec3), up(vec3)")
		return core.NIL_VALUE
	}

	posVal := vm.Stack(arg_stackptr)
	targetVal := vm.Stack(arg_stackptr + 1)
	upVal := vm.Stack(arg_stackptr + 2)

	if posVal.Type != core.VAL_VEC3 || targetVal.Type != core.VAL_VEC3 || upVal.Type != core.VAL_VEC3 {
		vm.RunTimeError("camera arguments must be vec3")
		return core.NIL_VALUE
	}

	posObj := posVal.Obj.(*core.Vec3Object)
	targetObj := targetVal.Obj.(*core.Vec3Object)
	upObj := upVal.Obj.(*core.Vec3Object)

	camera := rl.Camera3D{
		Position:   rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)},
		Target:     rl.Vector3{X: float32(targetObj.X), Y: float32(targetObj.Y), Z: float32(targetObj.Z)},
		Up:         rl.Vector3{X: float32(upObj.X), Y: float32(upObj.Y), Z: float32(upObj.Z)},
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	o := MakeCameraObject(camera)
	RegisterAllCameraMethods(o)
	return core.MakeObjectValue(o, true)
}

type CameraObject struct {
	core.BuiltInObject
	Camera    rl.Camera3D
	Methods   map[int]*core.BuiltInObject
	Constants map[int]core.Value
}

func MakeCameraObject(camera rl.Camera3D) *CameraObject {
	return &CameraObject{
		BuiltInObject: core.BuiltInObject{},
		Camera:        camera,
		Methods:       make(map[int]*core.BuiltInObject),
		Constants:     make(map[int]core.Value),
	}
}

func (c *CameraObject) IsObject()                      {}
func (c *CameraObject) GetType() core.ObjectType       { return core.OBJECT_NATIVE }
func (c *CameraObject) GetNativeType() core.NativeType { return core.NATIVE_CAMERA }
func (c *CameraObject) String() string                 { return "<Camera3D>" }
func (c *CameraObject) IsBuiltIn() bool                { return true }

func (c *CameraObject) GetMethod(stringId int) *core.BuiltInObject {
	return c.Methods[stringId]
}

func (c *CameraObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if c.Methods == nil {
		c.Methods = make(map[int]*core.BuiltInObject)
	}
	c.Methods[core.InternName(name)] = method
}

func (c *CameraObject) GetConstant(stringId int) core.Value {
	rv, ok := c.Constants[stringId]
	if !ok {
		return core.NIL_VALUE
	}
	return rv
}

func (c *CameraObject) RegisterConstant(name string, value core.Value) {
	if c.Constants == nil {
		c.Constants = make(map[int]core.Value)
	}
	c.Constants[core.InternName(name)] = value
}
