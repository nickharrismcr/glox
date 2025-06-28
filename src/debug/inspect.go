package debug

import (
	"fmt"
	"glox/src/core"
	"sort"
	"strings"
)

func FrameDictValue(vm core.VMContext) core.Value {
	frameCount := vm.FrameCount()

	return FrameDictValueFromFrame(frameCount, vm)
}

func FrameDictValueFromFrame(frameCount int, vm core.VMContext) core.Value {
	dict := core.MakeEmptyDictObject()
	frame := vm.FrameAt(frameCount - 1)
	if frame == nil {
		return core.NIL_VALUE
	}
	dict.Set("function", core.MakeObjectValue(frame.Closure.Function.Name, true))
	dict.Set("line", core.MakeIntValue(frame.Closure.Function.Chunk.Lines[frame.Ip], true))
	dict.Set("file", core.MakeStringObjectValue(vm.FileName(), true))
	dict.Set("args", ListOfArgs(frame, vm))
	locals := DictOfLocals(frame, vm)
	dict.Set("locals", locals)
	globals := DictOfGlobals(vm)
	dict.Set("globals", globals)
	if frameCount > 0 {
		dict.Set("prev_frame", FrameDictValueFromFrame(frameCount-1, vm))
	}
	return core.MakeObjectValue(dict, false)
}

func ListOfArgs(frame *core.CallFrame, vm core.VMContext) core.Value {
	list := []core.Value{}
	for i := 0; i <= frame.Closure.Function.Arity; i++ {
		arg := vm.Stack(frame.Slots + i)
		list = append(list, arg)
	}
	return core.MakeObjectValue(core.MakeListObject(list, false), false)
}

func DictOfLocals(frame *core.CallFrame, vm core.VMContext) core.Value {
	dict := core.MakeEmptyDictObject()
	st := vm.StackTop()
	localSlots := frame.Slots + frame.Closure.Function.Arity
	localVars := frame.Closure.Function.Chunk.LocalVars

	for slot := localSlots; slot < st; slot++ {
		if slot >= len(localVars) {
			break // Avoid out of bounds access if there are more slots than local variables
		}
		value := vm.Stack(slot)
		i := slot - localSlots
		localName := localVars[i].Name
		if localName != "" {
			dict.Set(localName, value)
		}
	}
	return core.MakeObjectValue(dict, false)
}

func DictOfGlobals(vm core.VMContext) core.Value {
	dict := core.MakeEmptyDictObject()
	globals := vm.GetGlobals()
	if globals == nil {
		return core.MakeObjectValue(dict, false)
	}

	for name, value := range globals.Vars {
		dict.Set(core.NameFromID(name), value)
	}
	return core.MakeObjectValue(dict, false)
}

func ShowGlobals(env *core.Environment) string {
	if env == nil {
		return "No globals (nil environment)"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n", env.Name))
	// Collect and sort the keys
	keys := make([]int, 0, len(env.Vars))
	for k := range env.Vars {
		keys = append(keys, k)
	}
	// Sort keys by name for readability
	sort.Slice(keys, func(i, j int) bool {
		return core.NameFromID(keys[i]) < core.NameFromID(keys[j])
	})
	for _, k := range keys {
		v := env.Vars[k]
		sb.WriteString(fmt.Sprintf("%s -> %s\n", core.NameFromID(k), v))
	}
	return sb.String()
}
