package lox

import (
	"fmt"
)

type ModuleObject struct {
	name    string
	globals map[string]Value
}

func makeModuleObject(name string, globals map[string]Value) *ModuleObject {

	return &ModuleObject{
		name:    name,
		globals: globals,
	}
}

func (ModuleObject) isObject() {}

func (ModuleObject) getType() ObjectType {

	return OBJECT_MODULE
}

func (f *ModuleObject) String() string {

	return fmt.Sprintf("<module %s>", f.name)
}

// -------------------------------------------------------------------------------------------
