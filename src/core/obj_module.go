package core

import (
	"fmt"
)

type ModuleObject struct {
	Name        string
	Environment Environment
}

func MakeModuleObject(name string, environment Environment) *ModuleObject {

	return &ModuleObject{
		Name:        name,
		Environment: environment,
	}
}

func (ModuleObject) IsObject() {}

func (ModuleObject) GetType() ObjectType {

	return OBJECT_MODULE
}

func (f *ModuleObject) String() string {

	return fmt.Sprintf("<module %s>", f.Name)
}

// -------------------------------------------------------------------------------------------
