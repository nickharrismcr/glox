package lox

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

	if environment == nil {
		Debug("makeFunctionObject: environment is nil")
		Debugf("makeFunctionObject: filename: %s", filename)
	}
	return &FunctionObject{
		arity:       0,
		name:        makeStringObject(""),
		chunk:       newChunk(filename),
		environment: environment,
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
