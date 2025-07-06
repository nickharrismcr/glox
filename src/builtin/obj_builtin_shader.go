package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func ShaderBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	switch argCount {
	case 0:
		// Create empty shader object
		o := MakeShaderObject(rl.Shader{})
		RegisterAllShaderMethods(o)
		return core.MakeObjectValue(o, true)
	case 2:
		// Load shader from files
		vsFileVal := vm.Stack(arg_stackptr)
		fsFileVal := vm.Stack(arg_stackptr + 1)

		if !core.IsString(vsFileVal) || !core.IsString(fsFileVal) {
			vm.RunTimeError("shader expects string arguments for vertex and fragment shader files")
			return core.NIL_VALUE
		}

		vsFile := core.GetStringValue(vsFileVal)
		fsFile := core.GetStringValue(fsFileVal)

		shader := rl.LoadShader(vsFile, fsFile)
		o := MakeShaderObject(shader)
		RegisterAllShaderMethods(o)
		return core.MakeObjectValue(o, true)
	default:
		vm.RunTimeError("shader expects 0 or 2 arguments")
		return core.NIL_VALUE
	}
}

type ShaderObject struct {
	core.BuiltInObject
	Value   rl.Shader
	Methods map[int]*core.BuiltInObject
}

func MakeShaderObject(shader rl.Shader) *ShaderObject {
	return &ShaderObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         shader,
	}
}

func (o *ShaderObject) String() string {
	return fmt.Sprintf("<Shader ID:%d>", o.Value.ID)
}

func (o *ShaderObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *ShaderObject) GetNativeType() core.NativeType {
	return core.NATIVE_SHADER
}

func (o *ShaderObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *ShaderObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *ShaderObject) IsBuiltIn() bool {
	return true
}

func RegisterAllShaderMethods(o *ShaderObject) {
	o.RegisterMethod("load_from_memory", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("load_from_memory expects 2 arguments: vertex shader code, fragment shader code")
				return core.NIL_VALUE
			}

			vsCodeVal := vm.Stack(arg_stackptr)
			fsCodeVal := vm.Stack(arg_stackptr + 1)

			if !core.IsString(vsCodeVal) || !core.IsString(fsCodeVal) {
				vm.RunTimeError("load_from_memory expects string arguments")
				return core.NIL_VALUE
			}

			vsCode := core.GetStringValue(vsCodeVal)
			fsCode := core.GetStringValue(fsCodeVal)

			o.Value = rl.LoadShaderFromMemory(vsCode, fsCode)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("get_location", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("get_location expects 1 argument: uniform name")
				return core.NIL_VALUE
			}

			nameVal := vm.Stack(arg_stackptr)
			if !core.IsString(nameVal) {
				vm.RunTimeError("get_location expects string argument")
				return core.NIL_VALUE
			}

			name := core.GetStringValue(nameVal)
			location := rl.GetShaderLocation(o.Value, name)
			return core.MakeIntValue(int(location), true)
		},
	})

	o.RegisterMethod("set_value_float", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_value_float expects 2 arguments: location, value")
				return core.NIL_VALUE
			}

			locVal := vm.Stack(arg_stackptr)
			valueVal := vm.Stack(arg_stackptr + 1)

			if !locVal.IsInt() || !valueVal.IsFloat() {
				vm.RunTimeError("set_value_float expects int location and float value")
				return core.NIL_VALUE
			}

			location := int32(locVal.AsInt())
			value := float32(valueVal.AsFloat())

			rl.SetShaderValue(o.Value, location, []float32{value}, rl.ShaderUniformFloat)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("set_value_vec2", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_value_vec2 expects 2 arguments: location, vec2 value")
				return core.NIL_VALUE
			}

			locVal := vm.Stack(arg_stackptr)
			vec2Val := vm.Stack(arg_stackptr + 1)

			if !locVal.IsInt() || vec2Val.Type != core.VAL_VEC2 {
				vm.RunTimeError("set_value_vec2 expects int location and vec2 value")
				return core.NIL_VALUE
			}

			location := int32(locVal.AsInt())
			vec2Obj := vec2Val.Obj.(*core.Vec2Object)
			values := []float32{float32(vec2Obj.X), float32(vec2Obj.Y)}

			rl.SetShaderValue(o.Value, location, values, rl.ShaderUniformVec2)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("set_value_vec3", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_value_vec3 expects 2 arguments: location, vec3 value")
				return core.NIL_VALUE
			}

			locVal := vm.Stack(arg_stackptr)
			vec3Val := vm.Stack(arg_stackptr + 1)

			if !locVal.IsInt() || vec3Val.Type != core.VAL_VEC3 {
				vm.RunTimeError("set_value_vec3 expects int location and vec3 value")
				return core.NIL_VALUE
			}

			location := int32(locVal.AsInt())
			vec3Obj := vec3Val.Obj.(*core.Vec3Object)
			values := []float32{float32(vec3Obj.X), float32(vec3Obj.Y), float32(vec3Obj.Z)}

			rl.SetShaderValue(o.Value, location, values, rl.ShaderUniformVec3)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("set_value_vec4", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 {
				vm.RunTimeError("set_value_vec4 expects 2 arguments: location, vec4 value")
				return core.NIL_VALUE
			}

			locVal := vm.Stack(arg_stackptr)
			vec4Val := vm.Stack(arg_stackptr + 1)

			if !locVal.IsInt() || vec4Val.Type != core.VAL_VEC4 {
				vm.RunTimeError("set_value_vec4 expects int location and vec4 value")
				return core.NIL_VALUE
			}

			location := int32(locVal.AsInt())
			vec4Obj := vec4Val.Obj.(*core.Vec4Object)
			values := []float32{float32(vec4Obj.X), float32(vec4Obj.Y), float32(vec4Obj.Z), float32(vec4Obj.W)}

			rl.SetShaderValue(o.Value, location, values, rl.ShaderUniformVec4)
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("is_valid", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("is_valid expects no arguments")
				return core.NIL_VALUE
			}

			isValid := rl.IsShaderValid(o.Value)
			return core.MakeBooleanValue(isValid, true)
		},
	})

	o.RegisterMethod("unload", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("unload expects no arguments")
				return core.NIL_VALUE
			}

			rl.UnloadShader(o.Value)
			return core.NIL_VALUE
		},
	})
}
