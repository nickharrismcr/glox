package core

import (
	"fmt"
)

type InstanceObject struct {
	class  *ClassObject
	fields map[string]Value
}

func MakeInstanceObject(class *ClassObject) *InstanceObject {

	return &InstanceObject{
		class:  class,
		fields: map[string]Value{},
	}
}

func (InstanceObject) IsObject() {}

func (InstanceObject) GetType() ObjectType {

	return OBJECT_INSTANCE
}

func (f *InstanceObject) String() string {

	return fmt.Sprintf("<instance %s>", f.class.name.Get())
}

// -------------------------------------------------------------------------------------------
