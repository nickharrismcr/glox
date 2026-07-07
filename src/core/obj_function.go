package core

import (
	"fmt"
)

type FunctionObject struct {
	Arity        int  // total named parameter slots, including *rest
	MinArity     int  // minimum args a caller must supply (fixed params without defaults)
	IsVariadic   bool // last named parameter is *rest
	Chunk        *Chunk
	Name         StringObject
	UpvalueCount int
	Environment  *Environment
}

func MakeFunctionObject(filename string, environment *Environment) *FunctionObject {

	return &FunctionObject{
		Arity:       0,
		Name:        MakeStringObject(""),
		Chunk:       NewChunk(filename),
		Environment: environment,
	}
}

func (FunctionObject) IsObject() {}

func (FunctionObject) GetType() ObjectType {

	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {

	if f.Name.Get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.Name)
}
func

// -------------------------------------------------------------------------------------------
(t *FunctionObject) IsBuiltIn() bool {
	return false
}
