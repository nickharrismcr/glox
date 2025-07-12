package builtin

import (
	"glox/src/core"
)

func RegisterAllBatchInstancedMethods(o *BatchInstancedObject) {

	o.RegisterMethod("add", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("add() expects 3 arguments (position, rotation axis, angle)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			axisVal := vm.Stack(arg_stackptr + 1)
			angleVal := vm.Stack(arg_stackptr + 2)

			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() first argument must be a vec3 (position)")
				return core.NIL_VALUE
			}
			if axisVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() second argument must be a vec3 (rotation axis)")
				return core.NIL_VALUE
			}
			if angleVal.Type != core.VAL_FLOAT {
				vm.RunTimeError("add() third argument must be a float (angle)")
				return core.NIL_VALUE
			}

			pos := posVal.Obj.(*core.Vec3Object)
			axis := axisVal.Obj.(*core.Vec3Object)
			angle := angleVal.Float

			if !o.batch.AddInstance(pos.X, pos.Y, pos.Z, axis.X, axis.Y, axis.Z, angle) {
				vm.RunTimeError("add() : max instances reached, cannot add more")
			}
			return core.NIL_VALUE
		},
	})

	// call this to build the transform matrices from the instance list
	o.RegisterMethod("make_transforms", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("make_transforms() expects no arguments")
				return core.NIL_VALUE
			}

			o.batch.MakeTransforms()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("draw", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

			//core.LogFmtLn(core.INFO, "BatchInstancedObject.draw called with %d arguments", argCount)
			if argCount != 1 {
				vm.RunTimeError("draw() expects 1 argument (camera)")
				return core.NIL_VALUE
			}
			cameraVal := vm.Stack(arg_stackptr)
			if !cameraVal.IsObj() {
				vm.RunTimeError("draw() expected a camera object, got %s", cameraVal.String())
				return core.NIL_VALUE
			}
			co, ok := cameraVal.Obj.(*CameraObject)
			if !ok {
				vm.RunTimeError("draw() argument must be a camera")
				return core.NIL_VALUE
			}
			o.Draw(co)
			return core.NIL_VALUE
		},
	})

	// o.RegisterMethod("clear", &core.BuiltInObject{
	// 	Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	// 		if argCount != 0 {
	// 			vm.RunTimeError("clear() expects no arguments")
	// 			return core.NIL_VALUE
	// 		}

	// 		//o.Value.Clear()
	// 		return core.NIL_VALUE
	// 	},
	// })

	o.RegisterMethod("count", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("count() expects no arguments")
				return core.NIL_VALUE
			}

			count := len(o.batch.instances.list)
			return core.MakeIntValue(count, true)
		},
	})

}
