package core

import (
	"fmt"
)

type ClassObject struct {
	Name          StringObject
	Methods       map[string]Value
	StaticMethods map[string]Value
	Super         *ClassObject
}

func MakeClassObject(name string) *ClassObject {

	return &ClassObject{
		Name:          MakeStringObject(name),
		Methods:       map[string]Value{},
		StaticMethods: map[string]Value{},
	}
}

func (ClassObject) IsObject() {}

func (ClassObject) GetType() ObjectType {

	return OBJECT_CLASS
}

func (f *ClassObject) String() string {

	return fmt.Sprintf("<class %s>", f.Name.Get())
}

func (f *ClassObject) IsSubclassOf(other *ClassObject) bool {
	for c := f; c != nil; c = c.Super {
		if c == other {
			return true
		}
	}
	return false
}
func

// -------------------------------------------------------------------------------------------
(t *ClassObject) IsBuiltIn() bool {
	return false
}
