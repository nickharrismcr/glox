package core

import "time"

type ObjectType uint8
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
	NATIVE_BATCH
	NATIVE_BATCH_INSTANCED
	NATIVE_PHYSICS_WORLD
	NATIVE_PROCESS
	NATIVE_THREAD
	NATIVE_THREAD_CHANNEL
	NATIVE_MUTEX
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
	RunTimeErrorNamed(string, string, ...interface{})
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
	ResolveClass(name string) (*ClassObject, bool)

	// SpawnThread runs closure (with args) on a new goroutine-backed VM
	// instance, deep-copying closure/args first (see CopyValueForSpawn) so
	// the new thread shares no mutable captured state with the caller.
	SpawnThread(closure Value, args []Value) (*ThreadHandle, error)
	// ThreadChannels returns this VM's own communication channels, and
	// false unless this VM was itself created by SpawnThread -- called by
	// thread.channel() from inside a spawned worker.
	ThreadChannels() (*ThreadChannels, bool)
	// CallClosure synchronously invokes closure on the *current* VM (no
	// new VM, no copy, no goroutine) and returns its result -- used by
	// thread.spawn's worker body and by sync.Mutex.locked().
	CallClosure(closure Value, args []Value) (Value, error)
}

type BuiltInFn func(argCount int, args_stackptr int, vm VMContext) Value
