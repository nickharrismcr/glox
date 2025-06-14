package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type StringObject struct {
	chars *string
}

func makeStringObject(s string) StringObject {

	return StringObject{
		chars: &s,
	}
}

func (StringObject) IsObject() {}

func (StringObject) GetType() ObjectType {

	return OBJECT_STRING
}

func (s StringObject) GetMethod(name string) *BuiltInObject {

	switch name {

	case "replace":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 2 {
					vm.RunTimeError("replace takes two arguments.")
					return makeNilValue()
				}
				fromVal := vm.Peek(1)
				toVal := vm.Peek(0)
				return s.replace(fromVal, toVal)
			},
		}
	case "join":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 || !vm.Peek(0).isListObject() {
					vm.RunTimeError("Join takes one list argument.")
					return makeNilValue()
				}
				lstVal := vm.Peek(0)
				lst := lstVal.asList()
				v, err := lst.join(s.get())
				if err != nil {
					vm.RunTimeError("%v", err)
					return makeNilValue()
				}
				return v
			},
		}
	default:
		return nil
	}
}

func (s StringObject) get() string {

	return *s.chars
}

func (s StringObject) contains(v Value) Value {

	rv := strings.Contains(*s.chars, *v.asString().chars)
	return makeBooleanValue(rv, true)

}

func (s StringObject) replace(from Value, to Value) Value {

	old := from.asString().get()
	new := to.asString().get()
	rv := strings.Replace(*s.chars, old, new, -1)
	return makeObjectValue(makeStringObject(rv), false)
}

func (s StringObject) String() string {

	return fmt.Sprintf("\"%s\"", *s.chars)
}

func (s StringObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.get()) + ix
	}

	if ix < 0 || ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	so := makeStringObject(string(s.get()[ix]))
	return makeObjectValue(so, false), nil
}

func (s StringObject) slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(s.get()) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(s.get()) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	so := makeStringObject(s.get()[from_ix:to_ix])
	return makeObjectValue(so, false), nil

}

func (s StringObject) parseFloat() (float64, bool) {

	f, err := strconv.ParseFloat(s.get(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (s StringObject) parseInt() (int, bool) {

	i, err := strconv.ParseInt(s.get(), 10, 64)
	if err != nil {
		return 0, false
	}
	return int(i), true
}

//-------------------------------------------------------------------------------------------
