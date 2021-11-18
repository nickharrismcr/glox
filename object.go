package main

import "fmt"

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
	OBJECT_FUNCTION
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

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
