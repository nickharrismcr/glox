package core

import (
	"fmt"
)

type InstanceObject struct {
	Class  *ClassObject
	Fields map[int]Value
}

func MakeInstanceObject(class *ClassObject) *InstanceObject {

	return &InstanceObject{
		Class:  class,
		Fields: map[int]Value{},
	}
}

func (InstanceObject) IsObject() {}

func (InstanceObject) GetType() ObjectType {

	return OBJECT_INSTANCE
}

func (f *InstanceObject) String() string {

	return fmt.Sprintf("<instance %s>", f.Class.Name.Get())
}
func

// -------------------------------------------------------------------------------------------
(t *InstanceObject) IsBuiltIn() bool {
	return false
}
