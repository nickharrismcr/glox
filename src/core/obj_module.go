package core

import (
	"fmt"
)

type ModuleObject struct {
	Name		string
	Environment	*Environment
}

// MakeModuleObject stores environment by reference, never a copy -- a copy
// would duplicate Environment's mutex (see varsMu) while continuing to
// alias the same underlying Vars map, so two "different" environments
// would guard the same shared map with two different, uncontended locks.
func MakeModuleObject(name string, environment *Environment) *ModuleObject {

	return &ModuleObject{
		Name:		name,
		Environment:	environment,
	}
}

func (ModuleObject) IsObject()	{}

func (ModuleObject) GetType() ObjectType {

	return OBJECT_MODULE
}

func (f *ModuleObject) String() string {

	return fmt.Sprintf("<module %s>", f.Name)
}
func

// -------------------------------------------------------------------------------------------
(t *ModuleObject) IsBuiltIn() bool {
	return false
}
