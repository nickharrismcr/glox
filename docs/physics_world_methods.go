package builtin

import (
	"glox/src/core"
)

func RegisterAllPhysicsWorldMethods(o *PhysicsWorldObject) {

	o.RegisterMethod("add_material", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 3 {
				vm.RunTimeError("add_material() expects 3 arguments (restitution, friction, damping)")
				return core.NIL_VALUE
			}

			restVal := vm.Stack(arg_stackptr)
			fricVal := vm.Stack(arg_stackptr + 1)
			dampVal := vm.Stack(arg_stackptr + 2)

			if !restVal.IsNumber() || !fricVal.IsNumber() || !dampVal.IsNumber() {
				vm.RunTimeError("add_material() arguments must be numbers")
				return core.NIL_VALUE
			}

			id := o.Value.AddMaterial(restVal.AsFloat(), fricVal.AsFloat(), dampVal.AsFloat())
			return core.MakeIntValue(id, true)
		},
	})

	o.RegisterMethod("add", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 4 {
				vm.RunTimeError("add() expects 4 arguments (position, velocity, radius, material_id)")
				return core.NIL_VALUE
			}

			posVal := vm.Stack(arg_stackptr)
			velVal := vm.Stack(arg_stackptr + 1)
			radiusVal := vm.Stack(arg_stackptr + 2)
			matVal := vm.Stack(arg_stackptr + 3)

			if posVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() first argument must be a vec3 (position)")
				return core.NIL_VALUE
			}
			if velVal.Type != core.VAL_VEC3 {
				vm.RunTimeError("add() second argument must be a vec3 (velocity)")
				return core.NIL_VALUE
			}
			if !radiusVal.IsNumber() {
				vm.RunTimeError("add() third argument must be a number (radius)")
				return core.NIL_VALUE
			}
			if !matVal.IsInt() {
				vm.RunTimeError("add() fourth argument must be an integer (material_id)")
				return core.NIL_VALUE
			}

			pos := posVal.Obj.(*core.Vec3Object)
			vel := velVal.Obj.(*core.Vec3Object)

			id, err := o.Value.Add(
				PVec3{pos.X, pos.Y, pos.Z},
				PVec3{vel.X, vel.Y, vel.Z},
				radiusVal.AsFloat(),
				matVal.AsInt(),
			)
			if err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.MakeIntValue(id, true)
		},
	})

	o.RegisterMethod("remove", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("remove() expects 1 argument (id)")
				return core.NIL_VALUE
			}

			idVal := vm.Stack(arg_stackptr)
			if !idVal.IsInt() {
				vm.RunTimeError("remove() argument must be an integer (id)")
				return core.NIL_VALUE
			}

			if err := o.Value.Remove(idVal.AsInt()); err != nil {
				vm.RunTimeError(err.Error())
			}
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("get_position", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("get_position() expects 1 argument (id)")
				return core.NIL_VALUE
			}

			idVal := vm.Stack(arg_stackptr)
			if !idVal.IsInt() {
				vm.RunTimeError("get_position() argument must be an integer (id)")
				return core.NIL_VALUE
			}

			pos, err := o.Value.GetPosition(idVal.AsInt())
			if err != nil {
				vm.RunTimeError(err.Error())
				return core.NIL_VALUE
			}

			return core.MakeObjectValue(core.MakeVec3Object(pos.X, pos.Y, pos.Z), false)
		},
	})

	o.RegisterMethod("step", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("step() expects 1 argument (dt)")
				return core.NIL_VALUE
			}

			dtVal := vm.Stack(arg_stackptr)
			if !dtVal.IsNumber() {
				vm.RunTimeError("step() argument must be a number (dt)")
				return core.NIL_VALUE
			}

			o.Value.Step(dtVal.AsFloat())
			return core.NIL_VALUE
		},
	})

	// Returns a list of small immutable tuples: (a, b, normal, impulse),
	// one per pair that newly started touching during the last step().
	// Resting/still-touching pairs from prior frames are not repeated.
	o.RegisterMethod("collisions", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("collisions() expects no arguments")
				return core.NIL_VALUE
			}

			pairs := o.Value.Collisions()
			items := make([]core.Value, len(pairs))
			for i, p := range pairs {
				tupleItems := []core.Value{
					core.MakeIntValue(p.A, true),
					core.MakeIntValue(p.B, true),
					core.MakeObjectValue(core.MakeVec3Object(p.Normal.X, p.Normal.Y, p.Normal.Z), true),
					core.MakeFloatValue(p.Impulse, true),
				}
				items[i] = core.MakeObjectValue(core.MakeListObject(tupleItems, true), true)
			}

			return core.MakeObjectValue(core.MakeListObject(items, true), true)
		},
	})

	o.RegisterMethod("count", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("count() expects no arguments")
				return core.NIL_VALUE
			}
			return core.MakeIntValue(o.Value.Count(), true)
		},
	})
}
