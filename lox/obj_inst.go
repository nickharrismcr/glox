package lox

import (
	"fmt"
)

type InstanceObject struct {
	class  *ClassObject
	fields map[string]Value
}

func makeInstanceObject(class *ClassObject) *InstanceObject {

	return &InstanceObject{
		class:  class,
		fields: map[string]Value{},
	}
}

func (InstanceObject) isObject() {}

func (InstanceObject) getType() ObjectType {

	return OBJECT_INSTANCE
}

func (f *InstanceObject) String() string {

	return fmt.Sprintf("<instance %s>", f.class.name.get())
}

// -------------------------------------------------------------------------------------------
