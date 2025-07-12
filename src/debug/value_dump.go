package debug

import (
	"glox/src/core"
	"reflect"
	"strings"
)

func TraceDumpValue(where string, val core.Value) {
	core.LogFmtLn(core.TRACE, "Value at %s:\n", where)
	seen := make(map[uintptr]bool)
	dumpValue(val, seen, 0)
	core.LogFmtLn(core.TRACE, "End %s\n", where)
}

func dumpObject(obj core.Object, seen map[uintptr]bool, indent int) {
	if obj == nil {
		core.LogFmtLn(core.TRACE, "%s<nil object>\n", indentPad(indent))
		return
	}

	ptr := objectPtr(obj)
	if seen[ptr] {
		core.LogFmtLn(core.TRACE, "%s<%T @%p> (already seen)\n", indentPad(indent), obj, obj)
		return
	}
	seen[ptr] = true

	switch o := obj.(type) {

	case *core.FunctionObject:
		fo := obj.(*core.FunctionObject)
		core.LogFmtLn(core.TRACE, "%s<Function %s @%p>\n", indentPad(indent), fo.Name, fo)
		if fo.Environment != nil {
			core.LogFmtLn(core.TRACE, "%s  Env:\n", indentPad(indent))
			DumpEnvironment(fo.Environment, seen, indent+2)
		}
		for i, c := range o.Chunk.Constants {
			core.LogFmtLn(core.TRACE, "%s  Const[%d]: ", indentPad(indent), i)
			dumpValue(c, seen, indent+2)
		}

	case *core.ClosureObject:
		core.LogFmtLn(core.TRACE, "%s<Closure @%p>\n", indentPad(indent), o)
		dumpObject(o.Function, seen, indent+1)

	case *core.ModuleObject:
		core.LogFmtLn(core.TRACE, "%s<Module %s @%p>\n", indentPad(indent), o.Name, o)
		mo := obj.(*core.ModuleObject)
		core.LogFmtLn(core.TRACE, "%s  Env:\n", indentPad(indent))
		DumpEnvironment(&mo.Environment, seen, indent+2)

	default:
		core.LogFmtLn(core.TRACE, "%s<%T: %v>\n", indentPad(indent), o, o)
	}
}

func DumpEnvironment(env *core.Environment, seen map[uintptr]bool, indent int) {
	if env == nil {
		core.LogFmtLn(core.TRACE, "%s<nil environment>\n", indentPad(indent))
		return
	}

	ptr := environmentPtr(env)
	if seen[ptr] {
		core.LogFmtLn(core.TRACE, "%s<Environment @%p> (already seen)\n", indentPad(indent), env)
		return
	}
	seen[ptr] = true
	core.LogFmtLn(core.TRACE, "%s<Environment @%p '%s'>\n", indentPad(indent), env, env.Name)
	core.LogFmtLn(core.TRACE, "%sVars:\n", indentPad(indent))
	for k, v := range env.Vars {
		core.LogFmtLn(core.TRACE, "%s%s: ", indentPad(indent+1), core.NameFromID(k))
		dumpValue(v, seen, indent+2)
	}

}

func dumpValue(val core.Value, seen map[uintptr]bool, indent int) {
	if val.IsObj() {
		dumpObject(val.Obj, seen, indent)
	} else {
		core.LogFmtLn(core.TRACE, "%s%v\n", indentPad(indent), val)
	}
}

func indentPad(n int) string {
	return strings.Repeat("  ", n)
}

func objectPtr(obj core.Object) uintptr {
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Ptr:
		return val.Pointer()
	default:
		return reflect.ValueOf(&obj).Pointer()
	}
}

func environmentPtr(env *core.Environment) uintptr {
	return reflect.ValueOf(env).Pointer()
}
