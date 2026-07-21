package core

import (
	"errors"
	"fmt"
	"strings"
)

type DictObject struct {
	Items map[int]Value
}

func MakeDictObject(items map[int]Value) *DictObject {

	return &DictObject{
		Items: items,
	}
}

func MakeEmptyDictObject() *DictObject {
	return MakeDictObject(make(map[int]Value))
}

func (DictObject) IsObject() {}

func (DictObject) GetType() ObjectType {
	return OBJECT_DICT
}

func (o *DictObject) String() string {
	depth := stringDepth.Add(1)
	defer stringDepth.Add(-1)
	if depth > maxStringDepth {
		return "..."
	}

	s := "Dict({ "
	for k, v := range o.Items {
		s = s + fmt.Sprintf("\"%s\":%s,", NameFromID(k), v.String())
	}
	return s[:len(s)-1] + " })"
}

// dictMethods is a shared, package-level table of dict methods keyed by
// interned name id, built once instead of being rebuilt (map + closures)
// for every dict literal. Each method recovers its receiver from the stack
// slot below the arguments rather than closing over a specific *DictObject.
var dictMethods map[int]*BuiltInObject

func init() {
	dictMethods = map[int]*BuiltInObject{
		InternName("get"): {
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 2 {
					vm.RunTimeError("Invalid argument count to get.")
					return NIL_VALUE
				}
				d := vm.Stack(arg_stackptr - 1).AsDict()
				key := vm.Stack(arg_stackptr)
				def := vm.Stack(arg_stackptr + 1)

				if key.IsStringObject() {
					rv, error := d.Get(key.AsString().Get())
					if error != nil {
						return def
					}
					return rv
				}

				vm.RunTimeError("Key argument to get must be a string")
				return NIL_VALUE
			},
		},
		InternName("keys"): {
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {

				if argCount != 0 {
					vm.RunTimeError("Invalid argument count to keys.")
					return NIL_VALUE
				}
				d := vm.Stack(arg_stackptr - 1).AsDict()
				return d.Keys()
			},
		},
		InternName("remove"): {
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 {
					vm.RunTimeError("Invalid argument count to remove.")
					return NIL_VALUE
				}
				d := vm.Stack(arg_stackptr - 1).AsDict()
				key := vm.Stack(arg_stackptr)

				if key.IsStringObject() {
					rv, error := d.Get(key.AsString().Get())
					if error != nil {
						return NIL_VALUE
					}
					delete(d.Items, InternName(key.AsString().Get()))
					return rv
				}

				vm.RunTimeError("Argument to remove must be key.")
				return NIL_VALUE
			},
		},
	}
}

func (d *DictObject) GetMethod(stringId int) *BuiltInObject {

	return dictMethods[stringId]
}

func (o *DictObject) Set(key string, value Value) {

	o.Items[InternName(key)] = value
}

func (o *DictObject) Get(key string) (Value, error) {

	rv, ok := o.Items[InternName(key)]
	if !ok {
		return NIL_VALUE, errors.New("key not found")
	}
	return rv, nil
}

func (o *DictObject) Keys() Value {

	Keys := []Value{}
	for k := range o.Items {
		key := strings.Replace(NameFromID(k), "\"", "", -1)
		so := MakeStringObject(key)
		v := MakeObjectValue(so, false)
		Keys = append(Keys, v)
	}
	return MakeObjectValue(MakeListObject(Keys, false), false)
}

// -------------------------------------------------------------------------------------------
func (t *DictObject) IsBuiltIn() bool {
	return true
}
