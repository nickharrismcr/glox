package lox

import (
	"fmt"
)

type FunctionObject struct {
	arity        int
	chunk        *Chunk
	name         StringObject
	upvalueCount int
}

func makeFunctionObject() *FunctionObject {

	return &FunctionObject{
		arity: 0,
		name:  makeStringObject(""),
		chunk: newChunk(),
	}
}

func (FunctionObject) isObject() {}

func (FunctionObject) getType() ObjectType {

	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {

	if f.name.get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.name)
}

// -------------------------------------------------------------------------------------------
