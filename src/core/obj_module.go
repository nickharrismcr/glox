package core

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

func (ModuleObject) IsObject() {}

func (ModuleObject) GetType() ObjectType {

	return OBJECT_MODULE
}

func (f *ModuleObject) String() string {

	return fmt.Sprintf("<module %s>", f.name)
}

// -------------------------------------------------------------------------------------------
