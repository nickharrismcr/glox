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
	makeBuiltInModule(vm, "gfx")
	makeBuiltInModule(vm, "physics")
	makeBuiltInModule(vm, "re")
	makeBuiltInModule(vm, "pickle")
	makeBuiltInModule(vm, "process")
	makeBuiltInModule(vm, "thread")
	makeBuiltInModule(vm, "sync")

	core.Log(core.INFO, "Defining built-in functions")

	defineBuiltIn(vm, "sys", "args", builtin.ArgsBuiltIn)
	defineBuiltIn(vm, "sys", "clock", builtin.ClockBuiltIn)
	defineBuiltIn(vm, "sys", "sleep", builtin.SleepBuiltIn)
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
	defineBuiltIn(vm, "gfx", "lox_mandel_array", builtin.MandelArrayBuiltIn)
	defineBuiltIn(vm, "gfx", "lox_julia_array", builtin.JuliaArrayBuiltIn)
	defineBuiltIn(vm, "gfx", "draw_png", builtin.DrawPNGBuiltIn)
	defineBuiltIn(vm, "", "replace", builtin.ReplaceBuiltIn)
	defineBuiltIn(vm, "", "format", builtin.FormatBuiltIn)
	defineBuiltIn(vm, "", "range", builtin.RangeBuiltIn)
	defineBuiltIn(vm, "", "rand", builtin.RandBuiltIn)
	defineBuiltIn(vm, "", "_atan2", builtin.Atan2BuiltIn)
	defineBuiltIn(vm, "gfx", "encode_rgba", builtin.EncodeRGBABuiltIn)
	defineBuiltIn(vm, "gfx", "decode_rgba", builtin.DecodeRGBABuiltIn)
	defineBuiltIn(vm, "", "vec2", builtin.Vec2BuiltIn)
	defineBuiltIn(vm, "", "vec3", builtin.Vec3BuiltIn)
	defineBuiltIn(vm, "", "vec4", builtin.Vec4BuiltIn)
	defineBuiltIn(vm, "gfx", "window", builtin.WindowBuiltIn)
	defineBuiltIn(vm, "gfx", "image", builtin.ImageBuiltIn)
	defineBuiltIn(vm, "gfx", "texture", builtin.TextureBuiltIn)
	defineBuiltIn(vm, "gfx", "render_texture", builtin.RenderTextureBuiltIn)
	defineBuiltIn(vm, "gfx", "shader", builtin.ShaderBuiltIn)
	defineBuiltIn(vm, "gfx", "camera", builtin.CameraBuiltIn)
	defineBuiltIn(vm, "gfx", "batch", builtin.BatchBuiltIn)
	defineBuiltIn(vm, "gfx", "batch_instanced", builtin.BatchInstancedBuiltIn)
	defineBuiltIn(vm, "physics", "physics_world", builtin.PhysicsWorldBuiltIn)
	defineBuiltIn(vm, "gfx", "float_array", builtin.FloatArrayBuiltin)
	defineBuiltIn(vm, "inspect", "dump_frame", builtin.DumpFrameBuiltIn)
	defineBuiltIn(vm, "inspect", "get_frame", builtin.GetFrameBuiltIn)

	// os module functions
	defineBuiltIn(vm, "os", "open", builtin.OpenBuiltIn)
	defineBuiltIn(vm, "os", "close", builtin.CloseBuiltIn)
	defineBuiltIn(vm, "os", "readln", builtin.ReadlnBuiltIn)
	defineBuiltIn(vm, "os", "write", builtin.WriteBuiltIn)
	defineBuiltIn(vm, "os", "read_all", builtin.ReadAllBuiltIn)
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

	// re module functions
	defineBuiltIn(vm, "re", "search", builtin.RegexSearchBuiltIn)
	defineBuiltIn(vm, "re", "match", builtin.RegexMatchBuiltIn)
	defineBuiltIn(vm, "re", "fullmatch", builtin.RegexFullmatchBuiltIn)
	defineBuiltIn(vm, "re", "sub", builtin.RegexSubBuiltIn)
	defineBuiltIn(vm, "re", "subn", builtin.RegexSubnBuiltIn)
	defineBuiltIn(vm, "re", "split", builtin.RegexSplitBuiltIn)
	defineBuiltIn(vm, "re", "findall", builtin.RegexFindallBuiltIn)
	defineBuiltIn(vm, "re", "compile", builtin.RegexCompileBuiltIn)

	// pickle module functions
	defineBuiltIn(vm, "pickle", "dumps", builtin.DumpsBuiltIn)
	defineBuiltIn(vm, "pickle", "loads", builtin.LoadsBuiltIn)

	// process module functions
	defineBuiltIn(vm, "process", "spawn", builtin.SpawnBuiltIn)
	defineBuiltIn(vm, "process", "parent", builtin.ParentBuiltIn)
	defineBuiltIn(vm, "process", "wait_any", builtin.WaitAnyBuiltIn)

	// thread module functions
	defineBuiltIn(vm, "thread", "spawn", builtin.ThreadSpawnBuiltIn)
	defineBuiltIn(vm, "thread", "channel", builtin.ThreadChannelBuiltIn)

	// sync module functions
	defineBuiltIn(vm, "sync", "Mutex", builtin.MutexBuiltIn)

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
	module := core.MakeModuleObject(moduleName, env)
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
	_, _ = subvm.Interpret(source, moduleName)
	fn := subvm.Frames[0].Closure.Function
	// Read from the globals slice (indexed by slot) rather than Vars map,
	// since OP_DEFINE_GLOBAL now only writes to the fast Globals slice.
	for slot, name := range fn.Chunk.GlobalNames {
		if fn.Environment.Defined[slot] {
			vm.BuiltIns[core.InternName(name)] = fn.Environment.Globals[slot]
		}
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
class PickleError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "PickleError";
	}
	toString() {
		return this.msg;
	}
}
class ProcessError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "ProcessError";
	}
	toString() {
		return this.msg;
	}
}
class ThreadError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "ThreadError";
	}
	toString() {
		return this.msg;
	}
}
class SyncError < Exception {
    init(msg) {
	    this.msg = msg;
		this.name = "SyncError";
	}
	toString() {
		return this.msg;
	}
}
`
