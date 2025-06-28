package core

import "time"

type ObjectType int
type NativeType int

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
	OBJECT_ITERATOR
	OBJECT_VEC2
	OBJECT_VEC3
	OBJECT_VEC4
)

const (
	NATIVE_FLOAT_ARRAY NativeType = iota
	NATIVE_VEC2
	NATIVE_VEC3
	NATIVE_VEC4
	NATIVE_WINDOW
	NATIVE_IMAGE
	NATIVE_TEXTURE
	NATIVE_RENDER_TEXTURE
	NATIVE_CAMERA
	NATIVE_SHADER
)

type Object interface {
	IsObject()
	GetType() ObjectType
	String() string
	IsBuiltIn() bool
}

// lists and strings are iterable objects
type Iterable interface {
	GetIterator() (Value, bool)
}

type Iterator interface {
	Next() Value
}

type HasMethods interface {
	GetMethod(int) *BuiltInObject
}
type HasConstants interface {
	GetConstant(int) Value
}
type IsNative interface {
	GetNativeType() NativeType
}

type VMContext interface {
	Stack(int) Value
	RunTimeError(string, ...interface{})
	Args() []string
	StartTime() time.Time
	RaiseExceptionByName(string, string) bool
	Peek(int) Value
	Frame() *CallFrame
	FrameAt(depth int) *CallFrame
	FrameCount() int
	StackTop() int
	ShowStack() string
	GetGlobals() *Environment
	FileName() string
}

type BuiltInFn func(argCount int, args_stackptr int, vm VMContext) Value
