package core

import (
	"errors"
	"fmt"
	"strings"
)

type DictObject struct {
	Items map[string]Value
}

func MakeDictObject(items map[string]Value) *DictObject {

	return &DictObject{
		Items: items,
	}
}

func (DictObject) IsObject() {}

func (DictObject) GetType() ObjectType {
	return OBJECT_DICT
}

func (o *DictObject) String() string {
	s := "Dict({ "
	for k, v := range o.Items {
		s = s + fmt.Sprintf("\"%s\":%s,", k, v.String())
	}
	return s[:len(s)-1] + " })"
}

func (d *DictObject) GetMethod(name string) *BuiltInObject {

	switch name {

	case "get":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 2 {
					vm.RunTimeError("Invalid argument count to get.")
					return MakeNilValue()
				}
				key := vm.Stack(arg_stackptr)
				def := vm.Stack(arg_stackptr + 1)

				if key.IsStringObject() {
					if def.IsStringObject() {
						rv, error := d.Get(key.AsString().Get())
						if error != nil {
							return def
						}
						return rv
					}
				}

				vm.RunTimeError("Argument to get must be key, default")
				return MakeNilValue()
			},
		}
	case "keys":

		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {

				if argCount != 0 {
					vm.RunTimeError("Invalid argument count to keys.")
					return MakeNilValue()
				}
				return d.Keys()
			},
		}

	default:
		return nil
	}
}

func (o *DictObject) Set(key string, value Value) {

	o.Items[key] = value
}

func (o *DictObject) Get(key string) (Value, error) {

	rv, ok := o.Items[key]
	if !ok {
		return MakeNilValue(), errors.New("key not found")
	}
	return rv, nil
}

func (o *DictObject) Keys() Value {

	Keys := []Value{}
	for k := range o.Items {
		key := strings.Replace(k, "\"", "", -1)
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
