package core

import "time"

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
	OBJECT_FUNCTION
	OBJECT_CLOSURE
	OBJECT_UPVALUE
	OBJECT_NATIVE
	OBJECT_LIST
	OBJECT_DICT
	OBJECT_CLASS
	OBJECT_INSTANCE
	OBJECT_BOUNDMETHOD
	OBJECT_MODULE
	OBJECT_FILE
	OBJECT_FLOAT_ARRAY
	OBJECT_GRAPHICS
	OBJECT_IMAGE
	OBJECT_TEXTURE
	OBJECT_ITERATOR
)

type Object interface {
	IsObject()
	GetType() ObjectType
	String() string
	IsBuiltIn() bool
}

// lists and strings are iterable objects
type Iterable interface {
	GetIterator() Value
	Index(int) (Value, error)
	GetLength() int
}

type HasMethods interface {
	GetMethod(string) *BuiltInObject
}

type VMContext interface {
	Stack(int) Value
	RunTimeError(string, ...interface{})
	Args() []string
	StartTime() time.Time
	RaiseExceptionByName(string, string) bool
	Peek(int) Value
}

type BuiltInFn func(argCount int, args_stackptr int, vm VMContext) Value
