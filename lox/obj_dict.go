package lox

import (
	"errors"
	"fmt"
	"strings"
)

type DictObject struct {
	items map[string]Value
}

func makeDictObject(items map[string]Value) *DictObject {

	return &DictObject{
		items: items,
	}
}

func (DictObject) isObject() {}

func (DictObject) getType() ObjectType {
	return OBJECT_DICT
}

func (o *DictObject) String() string {
	s := "Dict({ "
	for k, v := range o.items {
		s = s + fmt.Sprintf("\"%s\":%s,", k, v.String())
	}
	return s[:len(s)-1] + " })"
}

func (d *DictObject) GetMethod(name string) *BuiltInObject {

	switch name {

	case "get":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 2 {
					vm.RunTimeError("Invalid argument count to get.")
					return makeNilValue()
				}
				key := vm.Stack(arg_stackptr)
				def := vm.Stack(arg_stackptr + 1)

				if key.isStringObject() {
					if def.isStringObject() {
						rv, error := d.get(key.asString().get())
						if error != nil {
							return def
						}
						return rv
					}
				}

				vm.RunTimeError("Argument to get must be key, default")
				return makeNilValue()
			},
		}
	case "keys":

		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {

				if argCount != 0 {
					vm.RunTimeError("Invalid argument count to keys.")
					return makeNilValue()
				}
				return d.keys()
			},
		}

	default:
		return nil
	}
}

func (o *DictObject) set(key string, value Value) {

	o.items[key] = value
}

func (o *DictObject) get(key string) (Value, error) {

	rv, ok := o.items[key]
	if !ok {
		return makeNilValue(), errors.New("key not found")
	}
	return rv, nil
}

func (o *DictObject) keys() Value {

	keys := []Value{}
	for k := range o.items {
		key := strings.Replace(k, "\"", "", -1)
		so := makeStringObject(key)
		v := makeObjectValue(so, false)
		keys = append(keys, v)
	}
	return makeObjectValue(makeListObject(keys, false), false)
}

//-------------------------------------------------------------------------------------------
