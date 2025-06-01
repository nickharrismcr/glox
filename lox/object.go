package lox

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
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

type BuiltInFn func(argCount int, args_stackptr int, vm *VM) Value
