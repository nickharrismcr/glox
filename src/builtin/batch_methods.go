package builtin

import (
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RegisterAllBatchMethods(o *BatchObject) {

	o.RegisterMethod("add", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("add() expects 3 arguments (position, size, color)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			sizeVal := vm.Stack(arg_stackptr + 1)
			colorVal := vm.Stack(arg_stackptr + 2)

			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() first argument must be a vec3 (position)")
				return core.NIL_VALUE
			}
			if sizeVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() second argument must be a vec3 (size)")
				return core.NIL_VALUE
			}
			if colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("add() third argument must be a vec4 (color)")
				return core.NIL_VALUE
			}

			pos := posVal.Obj.(*core.Vec3Object)
			size := sizeVal.Obj.(*core.Vec3Object)
			color := colorVal.Obj.(*core.Vec4Object)

			index := o.Value.Add(pos, size, color)
			return core.MakeIntValue(index, true)
		},
	})

	o.RegisterMethod("add_triangle3", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("add_triangle3() expects 4 arguments (point1, point2, point3, color)")
				return core.NIL_VALUE
			}

			// Only allow this method for BATCH_TRIANGLE3 type
			if o.Value.BatchType != BATCH_TRIANGLE3 {
				vm.RunTimeError("add_triangle3() can only be used with BATCH_TRIANGLE3 batch type")
				return core.NIL_VALUE
			}

			p1Val := vm.Stack(arg_stackptr)
			p2Val := vm.Stack(arg_stackptr + 1)
			p3Val := vm.Stack(arg_stackptr + 2)
			colorVal := vm.Stack(arg_stackptr + 3)

			if p1Val.Type != core.VAL_VEC3 {
				vm.RunTimeError("add_triangle3() first argument must be a vec3 (point1)")
				return core.NIL_VALUE
			}
			if p2Val.Type != core.VAL_VEC3 {
				vm.RunTimeError("add_triangle3() second argument must be a vec3 (point2)")
				return core.NIL_VALUE
			}
			if p3Val.Type != core.VAL_VEC3 {
				vm.RunTimeError("add_triangle3() third argument must be a vec3 (point3)")
				return core.NIL_VALUE
			}
			if colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("add_triangle3() fourth argument must be a vec4 (color)")
				return core.NIL_VALUE
			}

			p1 := p1Val.Obj.(*core.Vec3Object)
			p2 := p2Val.Obj.(*core.Vec3Object)
			p3 := p3Val.Obj.(*core.Vec3Object)
			color := colorVal.Obj.(*core.Vec4Object)

			index := o.Value.AddTriangle3(p1, p2, p3, color)
			return core.MakeIntValue(index, true)
		},
	})

	o.RegisterMethod("add_textured_cube", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("add_textured_cube() expects 4 arguments (texture, position, size, base_color)")
				return core.NIL_VALUE
			}

			// Check if this is a textured cube batch
			if o.Value.BatchType != BATCH_TEXTURED_CUBE {
				vm.RunTimeError("add_textured_cube() can only be used with BATCH_TEXTURED_CUBE batch type")
				return core.NIL_VALUE
			}

			textureVal := vm.Stack(arg_stackptr)
			posVal := vm.Stack(arg_stackptr + 1)
			sizeVal := vm.Stack(arg_stackptr + 2)
			colorVal := vm.Stack(arg_stackptr + 3)

			// Extract texture object (can be either TextureObject or RenderTextureObject)
			if textureVal.Type != core.VAL_OBJ {
				vm.RunTimeError("add_textured_cube() first argument must be a texture or render_texture object")
				return core.NIL_VALUE
			}

			var rayTexture rl.Texture2D
			if textureObj, ok := textureVal.Obj.(*TextureObject); ok {
				rayTexture = textureObj.Data.Texture
			} else if renderTextureObj, ok := textureVal.Obj.(*RenderTextureObject); ok {
				rayTexture = renderTextureObj.Data.RenderTexture.Texture
			} else {
				vm.RunTimeError("add_textured_cube() first argument must be a texture or render_texture object")
				return core.NIL_VALUE
			}

			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add_textured_cube() second argument must be a vec3 (position)")
				return core.NIL_VALUE
			}
			if sizeVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add_textured_cube() third argument must be a vec3 (size)")
				return core.NIL_VALUE
			}
			if colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("add_textured_cube() fourth argument must be a vec4 (base_color)")
				return core.NIL_VALUE
			}

			pos := posVal.Obj.(*core.Vec3Object)
			size := sizeVal.Obj.(*core.Vec3Object)
			color := colorVal.Obj.(*core.Vec4Object)

			index := o.Value.AddTexturedCube(rayTexture, pos, size, color)
			return core.MakeIntValue(index, true)
		},
	})

	o.RegisterMethod("set_position", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_position() expects 2 arguments (index, position)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			posVal := vm.Stack(arg_stackptr + 1)

			if !indexVal.IsInt() {
				vm.RunTimeError("set_position() first argument must be an integer (index)")
				return core.NIL_VALUE
			}
			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("set_position() second argument must be a vec3 (position)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			pos := posVal.Obj.(*core.Vec3Object)

			if err := o.Value.SetPosition(index, pos); err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("set_color", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_color() expects 2 arguments (index, color)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			colorVal := vm.Stack(arg_stackptr + 1)

			if !indexVal.IsInt() {
				vm.RunTimeError("set_color() first argument must be an integer (index)")
				return core.NIL_VALUE
			}
			if colorVal.Type != core.VAL_VEC4 {
				vm.RunTimeError("set_color() second argument must be a vec4 (color)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			color := colorVal.Obj.(*core.Vec4Object)

			if err := o.Value.SetColor(index, color); err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("set_size", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_size() expects 2 arguments (index, size)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			sizeVal := vm.Stack(arg_stackptr + 1)

			if !indexVal.IsInt() {
				vm.RunTimeError("set_size() first argument must be an integer (index)")
				return core.NIL_VALUE
			}
			if sizeVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("set_size() second argument must be a vec3 (size)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			size := sizeVal.Obj.(*core.Vec3Object)

			if err := o.Value.SetSize(index, size); err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("get_position", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("get_position() expects 1 argument (index)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			if !indexVal.IsInt() {
				vm.RunTimeError("get_position() argument must be an integer (index)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			pos, err := o.Value.GetPosition(index)
			if err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.MakeObjectValue(pos, false)
		},
	})

	o.RegisterMethod("get_color", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("get_color() expects 1 argument (index)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			if !indexVal.IsInt() {
				vm.RunTimeError("get_color() argument must be an integer (index)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			color, err := o.Value.GetColor(index)
			if err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.MakeObjectValue(color, false)
		},
	})

	o.RegisterMethod("get_size", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("get_size() expects 1 argument (index)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			if !indexVal.IsInt() {
				vm.RunTimeError("get_size() argument must be an integer (index)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			size, err := o.Value.GetSize(index)
			if err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.MakeObjectValue(size, false)
		},
	})

	o.RegisterMethod("draw", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("draw() expects no arguments")
				return core.NIL_VALUE
			}
			o.Value.Draw()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("draw_culled", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("draw_culled() expects 2 arguments (camera_position, max_distance)")
				return core.NIL_VALUE
			}

			camPosVal := vm.Stack(arg_stackptr)
			maxDistVal := vm.Stack(arg_stackptr + 1)

			if camPosVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("draw_culled() first argument must be a vec3 (camera position)")
				return core.NIL_VALUE
			}
			if !maxDistVal.IsFloat() && !maxDistVal.IsInt() {
				vm.RunTimeError("draw_culled() second argument must be a number (max distance)")
				return core.NIL_VALUE
			}

			camPos := camPosVal.Obj.(*core.Vec3Object)
			maxDistance := float32(maxDistVal.AsFloat())

			// Convert to raylib Vector3
			rlCamPos := rl.Vector3{
				X: float32(camPos.X),
				Y: float32(camPos.Y),
				Z: float32(camPos.Z),
			}

			o.Value.DrawWithCulling(rlCamPos, maxDistance)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("draw_frustum_culled", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("draw_frustum_culled() expects 4 arguments (camera_position, camera_forward, max_distance, fov_degrees)")
				return core.NIL_VALUE
			}

			camPosVal := vm.Stack(arg_stackptr)
			camForwardVal := vm.Stack(arg_stackptr + 1)
			maxDistVal := vm.Stack(arg_stackptr + 2)
			fovVal := vm.Stack(arg_stackptr + 3)

			if camPosVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("draw_frustum_culled() first argument must be a vec3 (camera position)")
				return core.NIL_VALUE
			}
			if camForwardVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("draw_frustum_culled() second argument must be a vec3 (camera forward direction)")
				return core.NIL_VALUE
			}
			if !maxDistVal.IsFloat() && !maxDistVal.IsInt() {
				vm.RunTimeError("draw_frustum_culled() third argument must be a number (max distance)")
				return core.NIL_VALUE
			}
			if !fovVal.IsFloat() && !fovVal.IsInt() {
				vm.RunTimeError("draw_frustum_culled() fourth argument must be a number (FOV in degrees)")
				return core.NIL_VALUE
			}

			camPos := camPosVal.Obj.(*core.Vec3Object)
			camForward := camForwardVal.Obj.(*core.Vec3Object)
			maxDistance := float32(maxDistVal.AsFloat())
			fovDegrees := float32(fovVal.AsFloat())

			// Convert to raylib Vector3
			rlCamPos := rl.Vector3{
				X: float32(camPos.X),
				Y: float32(camPos.Y),
				Z: float32(camPos.Z),
			}
			rlCamForward := rl.Vector3{
				X: float32(camForward.X),
				Y: float32(camForward.Y),
				Z: float32(camForward.Z),
			}

			o.Value.DrawWithDirectionalCulling(rlCamPos, rlCamForward, maxDistance, fovDegrees)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("clear", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("clear() expects no arguments")
				return core.NIL_VALUE
			}

			o.Value.Clear()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("count", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("count() expects no arguments")
				return core.NIL_VALUE
			}

			count := o.Value.Count()
			return core.MakeIntValue(count, true)
		},
	})

	o.RegisterMethod("capacity", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("capacity() expects no arguments")
				return core.NIL_VALUE
			}

			capacity := o.Value.Capacity
			return core.MakeIntValue(capacity, true)
		},
	})

	o.RegisterMethod("reserve", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("reserve() expects 1 argument (capacity)")
				return core.NIL_VALUE
			}

			capacityVal := vm.Stack(arg_stackptr)
			if !capacityVal.IsInt() {
				vm.RunTimeError("reserve() argument must be an integer (capacity)")
				return core.NIL_VALUE
			}

			capacity := capacityVal.Int
			o.Value.Reserve(capacity)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("is_valid_index", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("is_valid_index() expects 1 argument (index)")
				return core.NIL_VALUE
			}

			indexVal := vm.Stack(arg_stackptr)
			if !indexVal.IsInt() {
				vm.RunTimeError("is_valid_index() argument must be an integer (index)")
				return core.NIL_VALUE
			}

			index := indexVal.Int
			valid := o.Value.IsValidIndex(index)
			return core.MakeBooleanValue(valid, true)
		},
	})
}
