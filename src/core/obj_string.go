package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var REPLACE = InternName("replace")
var JOIN = InternName("join")

type StringObject struct {
	Chars      *string
	InternedId int
}

func MakeStringObject(s string) StringObject {

	id := InternName(s)
	return StringObject{
		Chars:      &idToName[id],
		InternedId: id,
	}
}

func (StringObject) IsObject() {}

func (StringObject) GetType() ObjectType {

	return OBJECT_STRING
}

// stringMethods is a shared, package-level table of string methods keyed by
// interned name id, built once instead of allocating a fresh closure on
// every method lookup. Each method recovers its receiver from the stack
// slot below the arguments rather than closing over a specific StringObject.
var stringMethods map[int]*BuiltInObject

func init() {
	stringMethods = map[int]*BuiltInObject{
		REPLACE: {
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 2 {
					vm.RunTimeError("replace takes two arguments.")
					return NIL_VALUE
				}
				s := vm.Stack(arg_stackptr - 1).AsString()
				fromVal := vm.Peek(1)
				toVal := vm.Peek(0)
				return s.Replace(fromVal, toVal)
			},
		},
		JOIN: {
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 || !vm.Peek(0).IsListObject() {
					vm.RunTimeError("Join takes one list argument.")
					return NIL_VALUE
				}
				s := vm.Stack(arg_stackptr - 1).AsString()
				lstVal := vm.Peek(0)
				lst := lstVal.AsList()
				v, err := lst.Join(s.Get())
				if err != nil {
					vm.RunTimeError("%v", err)
					return NIL_VALUE
				}
				return v
			},
		},
	}
}

func (s StringObject) GetMethod(stringId int) *BuiltInObject {

	return stringMethods[stringId]
}

func (s StringObject) Get() string {

	return *s.Chars
}

func (o StringObject) GetIterator() (Value, bool) {
	return MakeObjectValue(MakeStringIteratorObject(o), false), true
}

func (s StringObject) GetLength() int {
	return len(s.Get())
}

func (s StringObject) Contains(v Value) Value {

	rv := strings.Contains(*s.Chars, *v.AsString().Chars)
	return MakeBooleanValue(rv, true)

}

func (s StringObject) Replace(from Value, to Value) Value {

	old := from.AsString().Get()
	new := to.AsString().Get()
	rv := strings.Replace(*s.Chars, old, new, -1)
	return MakeStringObjectValue(rv, false)
}

func (s StringObject) String() string {

	return fmt.Sprintf("\"%s\"", *s.Chars)
}

func (s StringObject) Index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.Get()) + ix
	}

	if ix < 0 || ix > len(s.Get()) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	return MakeStringObjectValue(string(s.Get()[ix]), false), nil
}

func (s StringObject) Slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(s.Get()) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(s.Get()) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(s.Get()) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(s.Get()) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	return MakeStringObjectValue(s.Get()[from_ix:to_ix], false), nil

}

func (s StringObject) ParseFloat() (float64, bool) {

	f, err := strconv.ParseFloat(s.Get(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (s StringObject) ParseInt() (int, bool) {

	i, err := strconv.ParseInt(s.Get(), 10, 64)
	if err != nil {
		return 0, false
	}
	return int(i), true
}
func

// -------------------------------------------------------------------------------------------
(t StringObject) IsBuiltIn() bool {
	return true
}
