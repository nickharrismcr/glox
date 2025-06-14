package lox

import (
	"glox/src/core"
	"reflect"
	"strings"
)

func DumpObject(obj core.Object) {
	seen := make(map[uintptr]bool)
	dumpObject(obj, seen, 0)
}

func DumpValue(where string, val core.Value) {
	Debugf("Value at %s:\n", where)
	seen := make(map[uintptr]bool)
	dumpValue(val, seen, 0)
}

func dumpObject(obj core.Object, seen map[uintptr]bool, indent int) {
	if obj == nil {
		Debugf("%s<nil object>\n", indentPad(indent))
		return
	}

	ptr := objectPtr(obj)
	if seen[ptr] {
		Debugf("%s<%T @%p> (already seen)\n", indentPad(indent), obj, obj)
		return
	}
	seen[ptr] = true

	switch o := obj.(type) {

	case *core.FunctionObject:
		fo := obj.(*core.FunctionObject)
		Debugf("%s<Function %s @%p>\n", indentPad(indent), fo.Name, fo)
		if fo.Environment != nil {
			Debugf("%s  Env:\n", indentPad(indent))
			dumpEnvironment(fo.Environment, seen, indent+2)
		}
		for i, c := range o.Chunk.Constants {
			Debugf("%s  Const[%d]: ", indentPad(indent), i)
			dumpValue(c, seen, indent+2)
		}

	case *core.ClosureObject:
		Debugf("%s<Closure @%p>\n", indentPad(indent), o)
		dumpObject(o.Function, seen, indent+1)

	default:
		Debugf("%s<%T: %v>\n", indentPad(indent), o, o)
	}
}

func dumpEnvironment(env *core.Environment, seen map[uintptr]bool, indent int) {
	if env == nil {
		Debugf("%s<nil environment>\n", indentPad(indent))
		return
	}

	ptr := environmentPtr(env)
	if seen[ptr] {
		Debugf("%s<Environment @%p> (already seen)\n", indentPad(indent), env)
		return
	}
	seen[ptr] = true
	Debugf("%s<Environment @%p '%s'>\n", indentPad(indent), env, env.Name)
	Debugf("%sVars:\n", indentPad(indent))
	for k, v := range env.Vars {
		Debugf("%s%s: ", indentPad(indent+1), k)
		dumpValue(v, seen, indent+2)
	}

}

func dumpValue(val core.Value, seen map[uintptr]bool, indent int) {
	if val.IsObj() {
		dumpObject(val.Obj, seen, indent)
	} else {
		Debugf("%s%v\n", indentPad(indent), val)
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
