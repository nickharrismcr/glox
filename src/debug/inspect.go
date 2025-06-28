package debug

import (
	"glox/src/core"
)

func FrameDictValue(vm core.VMContext) core.Value {
	frame := vm.Frame()
	if frame == nil {
		return core.NIL_VALUE
	}

	dict := core.MakeEmptyDictObject()
	dict.Set("function", core.MakeObjectValue(frame.Closure.Function.Name, true))
	dict.Set("line", core.MakeIntValue(frame.Closure.Function.Chunk.Lines[frame.Ip], true))
	dict.Set("file", core.MakeStringObjectValue(vm.FileName(), true))
	dict.Set("args", ListOfArgs(frame, vm))
	locals := DictOfLocals(frame, vm)
	dict.Set("locals", locals)
	globals := DictOfGlobals(vm)
	dict.Set("globals", globals)
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
	_ = st // to avoid unused variable warning
	localSlots := frame.Slots + frame.Closure.Function.Arity + 1
	for slot := localSlots; slot < st; slot++ {
		value := vm.Stack(slot)
		localName := frame.Closure.Function.Chunk.LocalVars[slot-frame.Slots].Name
		dict.Set(localName, value)
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
