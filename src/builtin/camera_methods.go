package builtin

import (
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllCameraMethods(c *CameraObject) {
	c.RegisterMethod("set_position", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("set_position expects 1 argument (vec3)")
				return core.NIL_VALUE
			}
			posVal := vm.Stack(arg_stackptr)
			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("Expected Vec3 for camera position")
				return core.NIL_VALUE
			}
			posObj := posVal.Obj.(*core.Vec3Object)
			c.Camera.Position = rl.Vector3{X: float32(posObj.X), Y: float32(posObj.Y), Z: float32(posObj.Z)}
			return core.NIL_VALUE
		},
	})

	c.RegisterMethod("set_target", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("set_target expects 1 argument (vec3)")
				return core.NIL_VALUE
			}
			targetVal := vm.Stack(arg_stackptr)
			if targetVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("Expected Vec3 for camera target")
				return core.NIL_VALUE
			}
			targetObj := targetVal.Obj.(*core.Vec3Object)
			c.Camera.Target = rl.Vector3{X: float32(targetObj.X), Y: float32(targetObj.Y), Z: float32(targetObj.Z)}
			return core.NIL_VALUE
		},
	})

	c.RegisterMethod("set_fovy", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("set_fovy expects 1 argument (number)")
				return core.NIL_VALUE
			}
			fovVal := vm.Stack(arg_stackptr)
			c.Camera.Fovy = float32(fovVal.AsFloat())
			return core.NIL_VALUE
		},
	})

	c.RegisterMethod("update", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			rl.UpdateCamera(&c.Camera, rl.CameraFree)
			return core.NIL_VALUE
		},
	})
}
