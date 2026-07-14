package builtin

import "glox/src/core"

// groupArgToIndex resolves the (optional) single argument to group()/start()/end()/span()
// to a group index. Returns (index, ok).
func groupArgToIndex(o *RegexMatchObject, argCount int, arg core.Value) (int, bool) {
	if argCount == 0 {
		return 0, true
	}
	if arg.IsInt() {
		return arg.AsInt(), true
	}
	if arg.IsStringObject() {
		idx := o.GroupIndexByName(arg.AsString().Get())
		return idx, idx >= 0
	}
	return 0, false
}

func RegisterAllRegexMatchMethods(o *RegexMatchObject) {

	o.RegisterMethod("group", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount > 1 {
				vm.RunTimeError("group takes at most one argument.")
				return core.NIL_VALUE
			}
			var arg core.Value
			if argCount == 1 {
				arg = vm.Stack(arg_stackptr)
			}
			idx, ok := groupArgToIndex(o, argCount, arg)
			if !ok || idx < 0 || idx > o.GroupCount() {
				vm.RunTimeError("no such group in match.")
				return core.NIL_VALUE
			}
			text, participated := o.GroupText(idx)
			if !participated {
				return core.NIL_VALUE
			}
			return core.MakeStringObjectValue(text, false)
		},
	})

	o.RegisterMethod("groups", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("groups takes no arguments.")
				return core.NIL_VALUE
			}
			items := make([]core.Value, 0, o.GroupCount())
			for i := 1; i <= o.GroupCount(); i++ {
				text, participated := o.GroupText(i)
				if !participated {
					items = append(items, core.NIL_VALUE)
				} else {
					items = append(items, core.MakeStringObjectValue(text, false))
				}
			}
			return core.MakeObjectValue(core.MakeListObject(items, true), false)
		},
	})

	o.RegisterMethod("start", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount > 1 {
				vm.RunTimeError("start takes at most one argument.")
				return core.NIL_VALUE
			}
			var arg core.Value
			if argCount == 1 {
				arg = vm.Stack(arg_stackptr)
			}
			idx, ok := groupArgToIndex(o, argCount, arg)
			if !ok || idx < 0 || idx > o.GroupCount() {
				vm.RunTimeError("no such group in match.")
				return core.NIL_VALUE
			}
			return core.MakeIntValue(o.Indices[2*idx], false)
		},
	})

	o.RegisterMethod("end", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount > 1 {
				vm.RunTimeError("end takes at most one argument.")
				return core.NIL_VALUE
			}
			var arg core.Value
			if argCount == 1 {
				arg = vm.Stack(arg_stackptr)
			}
			idx, ok := groupArgToIndex(o, argCount, arg)
			if !ok || idx < 0 || idx > o.GroupCount() {
				vm.RunTimeError("no such group in match.")
				return core.NIL_VALUE
			}
			return core.MakeIntValue(o.Indices[2*idx+1], false)
		},
	})

	o.RegisterMethod("span", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount > 1 {
				vm.RunTimeError("span takes at most one argument.")
				return core.NIL_VALUE
			}
			var arg core.Value
			if argCount == 1 {
				arg = vm.Stack(arg_stackptr)
			}
			idx, ok := groupArgToIndex(o, argCount, arg)
			if !ok || idx < 0 || idx > o.GroupCount() {
				vm.RunTimeError("no such group in match.")
				return core.NIL_VALUE
			}
			items := []core.Value{
				core.MakeIntValue(o.Indices[2*idx], false),
				core.MakeIntValue(o.Indices[2*idx+1], false),
			}
			return core.MakeObjectValue(core.MakeListObject(items, true), false)
		},
	})

	o.RegisterMethod("groupdict", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("groupdict takes no arguments.")
				return core.NIL_VALUE
			}
			items := make(map[int]core.Value)
			for i := 1; i <= o.GroupCount(); i++ {
				name := o.Names[i]
				if name == "" {
					continue
				}
				text, participated := o.GroupText(i)
				if !participated {
					items[core.InternName(name)] = core.NIL_VALUE
				} else {
					items[core.InternName(name)] = core.MakeStringObjectValue(text, false)
				}
			}
			return core.MakeObjectValue(core.MakeDictObject(items), false)
		},
	})
}
