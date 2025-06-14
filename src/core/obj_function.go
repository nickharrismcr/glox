package core

import (
	"fmt"
)

type FunctionObject struct {
	arity        int
	chunk        *Chunk
	name         StringObject
	upvalueCount int
	environment  *Environment
}

func makeFunctionObject(filename string, environment *Environment) *FunctionObject {

	return &FunctionObject{
		arity:       0,
		name:        makeStringObject(""),
		chunk:       NewChunk(filename),
		environment: environment,
	}
}

func (FunctionObject) IsObject() {}

func (FunctionObject) GetType() ObjectType {

	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {

	if f.name.get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.name)
}

// -------------------------------------------------------------------------------------------
