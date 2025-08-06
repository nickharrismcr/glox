package builtin

import (
	"fmt"
	"glox/src/core"
	"glox/src/util"
	"os"
	"strings"
)

// File I/O functions

func OpenBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to open.")
		return core.NIL_VALUE
	}
	path := vm.Stack(arg_stackptr)
	mode := vm.Stack(arg_stackptr + 1)

	if path.Type != core.VAL_OBJ || path.Obj.GetType() != core.OBJECT_STRING ||
		mode.Type != core.VAL_OBJ || mode.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to open.")
		return core.NIL_VALUE
	}

	s_path := path.AsString().Get()
	s_mode := mode.AsString().Get()
	fp, err := openFile(s_path, s_mode)
	if err != nil {
		vm.RunTimeError("%v", err)
		return core.NIL_VALUE
	}
	file := core.MakeObjectValue(core.MakeFileObject(fp), true)
	return file
}

func CloseBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to close.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to close.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	fo.Close()
	return core.MakeBooleanValue(true, false)
}

func ReadlnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to readln.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to readln.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("readln attempted on closed file.")
		return core.NIL_VALUE
	}

	rv := fo.ReadLine()
	if rv.Type == core.VAL_NIL {
		vm.RaiseExceptionByName("EOFError", "End of file reached")
		return core.MakeBooleanValue(true, false)
	}
	return rv
}

func WriteBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to writeln.")
		return core.NIL_VALUE
	}
	fov := vm.Stack(arg_stackptr)
	str := vm.Stack(arg_stackptr + 1)

	if fov.Type != core.VAL_OBJ || fov.Obj.GetType() != core.OBJECT_FILE {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.NIL_VALUE
	}
	if str.Type != core.VAL_OBJ || str.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to writeln.")
		return core.NIL_VALUE
	}

	fo := fov.Obj.(*core.FileObject)
	if fo.Closed {
		vm.RunTimeError("writeln attempted on closed file.")
		return core.NIL_VALUE
	}

	fo.Write(str)
	return core.MakeBooleanValue(true, false)
}

// Encoding functions
func EncodeRGBABuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 3 {
		vm.RunTimeError("encode_rgb expects 3 arguments")
		return core.NIL_VALUE
	}
	rVal := vm.Stack(arg_stackptr)
	gVal := vm.Stack(arg_stackptr + 1)
	bVal := vm.Stack(arg_stackptr + 2)
	if !rVal.IsInt() || !gVal.IsInt() || !bVal.IsInt() {
		vm.RunTimeError("encode_rgb arguments must be integers")
		return core.NIL_VALUE
	}
	r := rVal.Int
	g := gVal.Int
	b := bVal.Int
	color := util.EncodeRGB(r, g, b)
	return core.MakeFloatValue(color, false)
}

func DecodeRGBABuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("decode_rgb expects 1 float argument")
		return core.NIL_VALUE
	}
	fVal := vm.Stack(arg_stackptr)

	if !fVal.IsFloat() {
		vm.RunTimeError("decode_rgb argument must be a float")
		return core.NIL_VALUE
	}
	f := fVal.Float
	r, g, b := util.DecodeRGB(f)
	rVal := core.MakeIntValue(int(r), false)
	gVal := core.MakeIntValue(int(g), false)
	bVal := core.MakeIntValue(int(b), false)
	ro := core.MakeListObject([]core.Value{rVal, gVal, bVal}, true)
	return core.MakeObjectValue(ro, false)
}

// OS Module Functions - Directory and File Operations

func ListdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to listdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to listdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	entries, err := os.ReadDir(path)
	if err != nil {
		vm.RunTimeError("Failed to read directory '%s': %v", path, err)
		return core.NIL_VALUE
	}
	// Create a list of filenames
	var items []core.Value
	for _, entry := range entries {
		filename := core.MakeStringObjectValue(entry.Name(), false)
		items = append(items, filename)
	}
	list := core.MakeListObject(items, false)

	return core.MakeObjectValue(list, false)
}

func IsdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to isdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to isdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	info, err := os.Stat(path)
	if err != nil {
		return core.MakeBooleanValue(false, false)
	}

	return core.MakeBooleanValue(info.IsDir(), false)
}

func IsfileBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to isfile.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to isfile, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	info, err := os.Stat(path)
	if err != nil {
		return core.MakeBooleanValue(false, false)
	}

	return core.MakeBooleanValue(!info.IsDir(), false)
}

func ExistsBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to exists.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to exists, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	_, err := os.Stat(path)
	return core.MakeBooleanValue(err == nil, false)
}

func MkdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to mkdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to mkdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		vm.RunTimeError("Failed to create directory '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

func RmdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to rmdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to rmdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Remove(path)
	if err != nil {
		vm.RunTimeError("Failed to remove directory '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

func RemoveBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to remove.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to remove, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Remove(path)
	if err != nil {
		vm.RunTimeError("Failed to remove file '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

func GetcwdBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("Invalid argument count to getcwd.")
		return core.NIL_VALUE
	}

	cwd, err := os.Getwd()
	if err != nil {
		vm.RunTimeError("Failed to get current directory: %v", err)
		return core.NIL_VALUE
	}

	return core.MakeStringObjectValue(cwd, false)
}

func ChdirBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to chdir.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to chdir, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	err := os.Chdir(path)
	if err != nil {
		vm.RunTimeError("Failed to change directory to '%s': %v", path, err)
		return core.NIL_VALUE
	}

	return core.MakeBooleanValue(true, false)
}

// Path Manipulation Functions

func JoinBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount < 1 {
		vm.RunTimeError("join requires at least one argument.")
		return core.NIL_VALUE
	}

	var paths []string
	for i := 0; i < argCount; i++ {
		pathVal := vm.Stack(arg_stackptr + i)
		if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
			vm.RunTimeError("Invalid argument type to join, expected string.")
			return core.NIL_VALUE
		}
		paths = append(paths, pathVal.AsString().Get())
	}

	result := paths[0]
	for i := 1; i < len(paths); i++ {
		if strings.HasSuffix(result, "/") || strings.HasSuffix(result, "\\") {
			result = result + paths[i]
		} else {
			result = result + "/" + paths[i]
		}
	}

	return core.MakeStringObjectValue(result, false)
}

func DirnameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to dirname.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to dirname, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Simple dirname implementation
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return core.MakeStringObjectValue(".", false)
	}
	if lastSlash == 0 {
		return core.MakeStringObjectValue("/", false)
	}

	return core.MakeStringObjectValue(path[:lastSlash], false)
}

func BasenameBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to basename.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to basename, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Simple basename implementation
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return core.MakeStringObjectValue(path, false)
	}

	return core.MakeStringObjectValue(path[lastSlash+1:], false)
}

func SpliTextBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("Invalid argument count to splitext.")
		return core.NIL_VALUE
	}

	pathVal := vm.Stack(arg_stackptr)
	if pathVal.Type != core.VAL_OBJ || pathVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("Invalid argument type to splitext, expected string.")
		return core.NIL_VALUE
	}

	path := pathVal.AsString().Get()
	// Find the last dot
	lastDot := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			lastDot = i
			break
		}
		if path[i] == '/' || path[i] == '\\' {
			break // Stop at directory separator
		}
	}

	var name, ext string
	if lastDot == -1 {
		name = path
		ext = ""
	} else {
		name = path[:lastDot]
		ext = path[lastDot:]
	}

	nameVal := core.MakeStringObjectValue(name, false)
	extVal := core.MakeStringObjectValue(ext, false)
	items := []core.Value{nameVal, extVal}
	list := core.MakeListObject(items, false)

	return core.MakeObjectValue(list, false)
}

// Helper function for file operations
func openFile(path string, mode string) (*os.File, error) {
	switch mode {
	case "r":
		return os.Open(path) // Read-only
	case "w":
		return os.Create(path) // Write (truncate if exists)
	case "a":
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Append
	default:
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}
}
