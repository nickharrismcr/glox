package builtin

import (
	"glox/src/core"
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
