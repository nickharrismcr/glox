package lox

import (
	"errors"
	"fmt"
	"strings"
)

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
	OBJECT_FUNCTION
	OBJECT_NATIVE
	OBJECT_LIST
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

type NativeFn func(argCount int, args_stackptr int, vm *VM) Value

//-------------------------------------------------------------------------------------------
type FunctionObject struct {
	arity int
	chunk *Chunk
	name  StringObject
}

func makeFunctionObject() *FunctionObject {
	return &FunctionObject{
		arity: 0,
		name:  MakeStringObject(""),
		chunk: newChunk(),
	}
}

func (_ FunctionObject) isObject() {}

func (_ FunctionObject) getType() ObjectType {
	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {
	if f.name.get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.name)
}

//-------------------------------------------------------------------------------------------

type StringObject struct {
	chars *string
}

func MakeStringObject(s string) StringObject {
	return StringObject{
		chars: &s,
	}
}

func (_ StringObject) isObject() {}

func (_ StringObject) getType() ObjectType {
	return OBJECT_STRING
}

func (s StringObject) get() string {
	return *s.chars
}

func (s StringObject) String() string {
	return fmt.Sprintf("\"%s\"", *s.chars)
}

func (s StringObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.get()) + ix
	}

	if ix < 0 || ix > len(s.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	so := MakeStringObject(string(s.get()[ix]))
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
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix < 0 || from_ix > len(s.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	so := MakeStringObject(s.get()[from_ix:to_ix])
	return makeObjectValue(so, false), nil

}

//-------------------------------------------------------------------------------------------

type NativeObject struct {
	function NativeFn
}

func makeNativeObject(function NativeFn) *NativeObject {
	return &NativeObject{
		function: function,
	}
}

func (_ NativeObject) isObject() {}

func (_ NativeObject) getType() ObjectType {
	return OBJECT_NATIVE
}

func (f *NativeObject) String() string {
	return "<built-in>"
}

//-------------------------------------------------------------------------------------------

type ListObject struct {
	items []Value
}

func makeListObject(items []Value) *ListObject {
	return &ListObject{
		items: items,
	}
}

func (_ ListObject) isObject() {}

func (_ ListObject) getType() ObjectType {
	return OBJECT_LIST
}

func (s *ListObject) get() []Value {
	return s.items
}

func (s *ListObject) String() string {

	list := []string{}

	for _, v := range s.items {
		list = append(list, v.String())
	}
	return fmt.Sprintf("[ %s ]", strings.Join(list, " , "))
}

func (s *ListObject) add(other *ListObject) *ListObject {
	l := []Value{}
	l = append(l, s.items...)
	l = append(l, other.items...)
	return makeListObject(l)
}

func (s *ListObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.get()) + ix
	}

	if ix < 0 || ix > len(s.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	return s.get()[ix], nil
}

func (s *ListObject) slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(s.items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(s.items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(s.items) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix < 0 || from_ix > len(s.items) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix > to_ix {
		return NilValue{}, errors.New("Invalid slice indices.")
	}

	lo := makeListObject(s.items[from_ix:to_ix])
	return makeObjectValue(lo, false), nil
}
