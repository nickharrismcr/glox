package lox

import (
	"reflect"
	"strings"
)

func DumpObject(obj Object) {
	seen := make(map[uintptr]bool)
	dumpObject(obj, seen, 0)
}

func DumpValue(where string, val Value) {
	Debugf("Value at %s:\n", where)
	seen := make(map[uintptr]bool)
	dumpValue(val, seen, 0)
}

func dumpObject(obj Object, seen map[uintptr]bool, indent int) {
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

	case *FunctionObject:
		fo := obj.(*FunctionObject)
		Debugf("%s<Function %s @%p>\n", indentPad(indent), fo.name, fo)
		if fo.environment != nil {
			Debugf("%s  Env:\n", indentPad(indent))
			dumpEnvironment(fo.environment, seen, indent+2)
		}
		for i, c := range o.chunk.constants {
			Debugf("%s  Const[%d]: ", indentPad(indent), i)
			dumpValue(c, seen, indent+2)
		}

	case *ClosureObject:
		Debugf("%s<Closure @%p>\n", indentPad(indent), o)
		dumpObject(o.function, seen, indent+1)

	default:
		Debugf("%s<%T: %v>\n", indentPad(indent), o, o)
	}
}

func dumpEnvironment(env *Environment, seen map[uintptr]bool, indent int) {
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
	Debugf("%s<Environment @%p '%s'>\n", indentPad(indent), env, env.name)
	Debugf("%sVars:\n", indentPad(indent))
	for k, v := range env.vars {
		Debugf("%s%s: ", indentPad(indent+1), k)
		dumpValue(v, seen, indent+2)
	}

}

func dumpValue(val Value, seen map[uintptr]bool, indent int) {
	if val.IsObj() {
		dumpObject(val.Obj, seen, indent)
	} else {
		Debugf("%s%v\n", indentPad(indent), val)
	}
}

func indentPad(n int) string {
	return strings.Repeat("  ", n)
}

func objectPtr(obj Object) uintptr {
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Ptr:
		return val.Pointer()
	default:
		return reflect.ValueOf(&obj).Pointer()
	}
}

func environmentPtr(env *Environment) uintptr {
	return reflect.ValueOf(env).Pointer()
}
