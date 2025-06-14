package core

import (
	"fmt"
)

type ClassObject struct {
	name    StringObject
	methods map[string]Value
	super   *ClassObject
}

func makeClassObject(name string) *ClassObject {

	return &ClassObject{
		name:    makeStringObject(name),
		methods: map[string]Value{},
	}
}

func (ClassObject) IsObject() {}

func (ClassObject) GetType() ObjectType {

	return OBJECT_CLASS
}

func (f *ClassObject) String() string {

	return fmt.Sprintf("<class %s>", f.name.get())
}

func (f *ClassObject) IsSubclassOf(other *ClassObject) bool {
	for c := f; c != nil; c = c.super {
		if c == other {
			return true
		}
	}
	return false
}

//-------------------------------------------------------------------------------------------
