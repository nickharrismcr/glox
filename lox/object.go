package lox

import "fmt"

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
	OBJECT_FUNCTION
	OBJECT_NATIVE
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
	if f.name.String() == "" {
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
	return *s.chars
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
