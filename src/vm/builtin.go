package vm

import (
	"glox/src/builtin"
	"glox/src/core"
)

func defineBuiltIn(vm *VM, module string, name string, fn core.BuiltInFn) {
	// Add the built-in to the specified module namespace (environment)

	if module != "" {
		addBuiltInModuleFunction(vm, module, name, fn)
	} else {
		vm.BuiltIns[core.InternName(name)] = core.MakeObjectValue(core.MakeBuiltInObject(fn), false)
	}

}

func DefineBuiltIns(vm *VM) {

	makeBuiltInModule(vm, "sys")
	makeBuiltInModule(vm, "inspect")
	makeBuiltInModule(vm, "colour_utils")
	makeBuiltInModule(vm, "os")

	core.Log(core.INFO, "Defining built-in functions")

	defineBuiltIn(vm, "sys", "args", builtin.ArgsBuiltIn)
	defineBuiltIn(vm, "sys", "clock", builtin.ClockBuiltIn)
	defineBuiltIn(vm, "", "type", builtin.TypeBuiltIn)
	defineBuiltIn(vm, "", "len", builtin.LenBuiltIn)
	defineBuiltIn(vm, "", "_sin", builtin.SinBuiltIn)
	defineBuiltIn(vm, "", "_cos", builtin.CosBuiltIn)
	defineBuiltIn(vm, "", "_tan", builtin.TanBuiltIn)
	defineBuiltIn(vm, "", "_sqrt", builtin.SqrtBuiltIn)
	defineBuiltIn(vm, "", "_pow", builtin.PowBuiltIn)
	defineBuiltIn(vm, "", "append", builtin.AppendBuiltIn)
	defineBuiltIn(vm, "", "float", builtin.FloatBuiltIn)
	defineBuiltIn(vm, "", "int", builtin.IntBuiltIn)
	defineBuiltIn(vm, "", "lox_mandel_array", builtin.MandelArrayBuiltIn)
	defineBuiltIn(vm, "", "lox_julia_array", builtin.JuliaArrayBuiltIn)
	defineBuiltIn(vm, "", "draw_png", builtin.DrawPNGBuiltIn)
	defineBuiltIn(vm, "", "replace", builtin.ReplaceBuiltIn)
	defineBuiltIn(vm, "", "format", builtin.FormatBuiltIn)
	defineBuiltIn(vm, "", "sleep", builtin.SleepBuiltIn)
	defineBuiltIn(vm, "", "range", builtin.RangeBuiltIn)
	defineBuiltIn(vm, "", "rand", builtin.RandBuiltIn)
	defineBuiltIn(vm, "", "_atan2", builtin.Atan2BuiltIn)
	defineBuiltIn(vm, "", "encode_rgba", builtin.EncodeRGBABuiltIn)
	defineBuiltIn(vm, "", "decode_rgba", builtin.DecodeRGBABuiltIn)
	defineBuiltIn(vm, "", "vec2", builtin.Vec2BuiltIn)
	defineBuiltIn(vm, "", "vec3", builtin.Vec3BuiltIn)
	defineBuiltIn(vm, "", "vec4", builtin.Vec4BuiltIn)
	defineBuiltIn(vm, "", "window", builtin.WindowBuiltIn)
	defineBuiltIn(vm, "", "image", builtin.ImageBuiltIn)
	defineBuiltIn(vm, "", "texture", builtin.TextureBuiltIn)
	defineBuiltIn(vm, "", "render_texture", builtin.RenderTextureBuiltIn)
	defineBuiltIn(vm, "", "shader", builtin.ShaderBuiltIn)
	defineBuiltIn(vm, "", "camera", builtin.CameraBuiltIn)
	defineBuiltIn(vm, "", "batch", builtin.BatchBuiltIn)
	defineBuiltIn(vm, "", "batch_instanced", builtin.BatchInstancedBuiltIn)
	defineBuiltIn(vm, "", "float_array", builtin.FloatArrayBuiltin)
	defineBuiltIn(vm, "inspect", "dump_frame", builtin.DumpFrameBuiltIn)
	defineBuiltIn(vm, "inspect", "get_frame", builtin.GetFrameBuiltIn)

	// os module functions
	defineBuiltIn(vm, "os", "open", builtin.OpenBuiltIn)
	defineBuiltIn(vm, "os", "close", builtin.CloseBuiltIn)
	defineBuiltIn(vm, "os", "readln", builtin.ReadlnBuiltIn)
	defineBuiltIn(vm, "os", "write", builtin.WriteBuiltIn)
	defineBuiltIn(vm, "os", "listdir", builtin.ListdirBuiltIn)
	defineBuiltIn(vm, "os", "isdir", builtin.IsdirBuiltIn)
	defineBuiltIn(vm, "os", "isfile", builtin.IsfileBuiltIn)
	defineBuiltIn(vm, "os", "exists", builtin.ExistsBuiltIn)
	defineBuiltIn(vm, "os", "mkdir", builtin.MkdirBuiltIn)
	defineBuiltIn(vm, "os", "rmdir", builtin.RmdirBuiltIn)
	defineBuiltIn(vm, "os", "remove", builtin.RemoveBuiltIn)
	defineBuiltIn(vm, "os", "getcwd", builtin.GetcwdBuiltIn)
	defineBuiltIn(vm, "os", "chdir", builtin.ChdirBuiltIn)
	defineBuiltIn(vm, "os", "join", builtin.JoinBuiltIn)
	defineBuiltIn(vm, "os", "dirname", builtin.DirnameBuiltIn)
	defineBuiltIn(vm, "os", "basename", builtin.BasenameBuiltIn)
	defineBuiltIn(vm, "os", "splitext", builtin.SpliTextBuiltIn)

	// Color utility functions
	defineBuiltIn(vm, "colour_utils", "fade", builtin.ColourUtilsFadeBuiltIn)
	defineBuiltIn(vm, "colour_utils", "tint", builtin.ColourUtilsTintBuiltIn)
	defineBuiltIn(vm, "colour_utils", "brightness", builtin.ColourUtilsBrightnessBuiltIn)
	defineBuiltIn(vm, "colour_utils", "lerp", builtin.ColourUtilsLerpBuiltIn)
	defineBuiltIn(vm, "colour_utils", "hsv_to_rgb", builtin.ColourUtilsHSVToRGBBuiltIn)
	defineBuiltIn(vm, "colour_utils", "random", builtin.ColourUtilsRandomBuiltIn)

	// lox built ins e.g Exception classes
	loadBuiltInFromSource(vm, exceptionSource, "exception")

	// Do NOT inject sys into the global environment here.
	// It must be imported by client code to be available.
}

// Helper functions for module management
func makeBuiltInModule(vm *VM, moduleName string) {

	env := core.NewEnvironment(moduleName)
	module := core.MakeModuleObject(moduleName, *env)
	vm.BuiltInModules[core.InternName(moduleName)] = module
	core.LogFmtLn(core.INFO, "Created built-in module %s", moduleName)

}

func addBuiltInModuleFunction(vm *VM, moduleName string, name string, fn core.BuiltInFn) {
	// Add a function to a built-in module
	module := vm.BuiltInModules[core.InternName(moduleName)]
	fo := core.MakeBuiltInObject(fn)
	module.Environment.Vars[core.InternName(name)] = core.MakeObjectValue(fo, false)
}

// load built-in functions from source code
func loadBuiltInFromSource(vm *VM, source string, moduleName string) {
	core.Log(core.INFO, "Loading built-in module ")
	subvm := NewVM("", false)
	//	DebugSuppress = true
	_, _ = subvm.Interpret(source, moduleName)
	for k, v := range subvm.Frames[0].Closure.Function.Environment.Vars {
		vm.BuiltIns[k] = v
	}

	core.DebugSuppress = false
}

// predefine an Exception class using Lox source
const exceptionSource = `class Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "Exception";
	}
	toString() {
	    return this.msg;
	}
}
class EOFError < Exception {
     init(msg) {
	    this.msg = msg;
		this.name = "EOFError";
	}
	toString() {
	    return this.msg;
	}
}
class RunTimeError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "RunTimeError";
	}
	toString() {
		return this.msg;
	}
}
`
