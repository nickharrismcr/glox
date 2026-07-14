package builtin

import "glox/src/core"

func RegisterAllRegexPatternMethods(o *RegexPatternObject) {

	o.RegisterMethod("search", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("Invalid argument count to Pattern.search.")
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.search string")
			if !ok {
				return core.NIL_VALUE
			}
			return matchFromRegex(o.Re, s)
		},
	})

	o.RegisterMethod("match", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("Invalid argument count to Pattern.match.")
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.match string")
			if !ok {
				return core.NIL_VALUE
			}
			re, err := o.StartAnchored()
			if err != nil {
				vm.RunTimeError("re: invalid pattern: %v", err)
				return core.NIL_VALUE
			}
			return matchFromRegex(re, s)
		},
	})

	o.RegisterMethod("fullmatch", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("Invalid argument count to Pattern.fullmatch.")
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.fullmatch string")
			if !ok {
				return core.NIL_VALUE
			}
			re, err := o.FullAnchored()
			if err != nil {
				vm.RunTimeError("re: invalid pattern: %v", err)
				return core.NIL_VALUE
			}
			return matchFromRegex(re, s)
		},
	})

	o.RegisterMethod("sub", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 && argCount != 3 {
				vm.RunTimeError("Pattern.sub expects 2 or 3 arguments.")
				return core.NIL_VALUE
			}
			repl, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.sub repl")
			if !ok {
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "Pattern.sub string")
			if !ok {
				return core.NIL_VALUE
			}
			count := 0
			if argCount == 3 {
				cv := vm.Stack(arg_stackptr + 2)
				if !cv.IsInt() {
					vm.RunTimeError("Pattern.sub count must be an integer.")
					return core.NIL_VALUE
				}
				count = cv.AsInt()
			}
			result, _ := subWithCount(o.Re, repl, s, count)
			return core.MakeStringObjectValue(result, false)
		},
	})

	o.RegisterMethod("subn", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 2 && argCount != 3 {
				vm.RunTimeError("Pattern.subn expects 2 or 3 arguments.")
				return core.NIL_VALUE
			}
			repl, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.subn repl")
			if !ok {
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr+1), "Pattern.subn string")
			if !ok {
				return core.NIL_VALUE
			}
			count := 0
			if argCount == 3 {
				cv := vm.Stack(arg_stackptr + 2)
				if !cv.IsInt() {
					vm.RunTimeError("Pattern.subn count must be an integer.")
					return core.NIL_VALUE
				}
				count = cv.AsInt()
			}
			result, n := subWithCount(o.Re, repl, s, count)
			items := []core.Value{
				core.MakeStringObjectValue(result, false),
				core.MakeIntValue(n, false),
			}
			return core.MakeObjectValue(core.MakeListObject(items, true), false)
		},
	})

	o.RegisterMethod("split", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 && argCount != 2 {
				vm.RunTimeError("Pattern.split expects 1 or 2 arguments.")
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.split string")
			if !ok {
				return core.NIL_VALUE
			}
			maxsplit := 0
			if argCount == 2 {
				mv := vm.Stack(arg_stackptr + 1)
				if !mv.IsInt() {
					vm.RunTimeError("Pattern.split maxsplit must be an integer.")
					return core.NIL_VALUE
				}
				maxsplit = mv.AsInt()
			}
			parts := splitWithMax(o.Re, s, maxsplit)
			items := make([]core.Value, 0, len(parts))
			for _, p := range parts {
				items = append(items, core.MakeStringObjectValue(p, false))
			}
			return core.MakeObjectValue(core.MakeListObject(items, false), false)
		},
	})

	o.RegisterMethod("findall", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("Invalid argument count to Pattern.findall.")
				return core.NIL_VALUE
			}
			s, ok := argAsPattern(vm, vm.Stack(arg_stackptr), "Pattern.findall string")
			if !ok {
				return core.NIL_VALUE
			}
			return findallResults(o.Re, s)
		},
	})
}
