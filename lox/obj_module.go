package lox

import (
	"fmt"
)

type ModuleObject struct {
	name        string
	environment Environment
}

func makeModuleObject(name string, environment Environment) *ModuleObject {

	return &ModuleObject{
		name:        name,
		environment: environment,
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
