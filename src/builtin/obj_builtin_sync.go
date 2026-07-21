package builtin

import (
	"sync"

	"glox/src/core"
)

// MutexObject wraps a real Go sync.Mutex. CopyValueForSpawn's default
// "anything else: share by pointer" rule already does the right thing for
// this type without any special-casing -- a Mutex captured by a spawned
// closure's upvalue must be *shared*, not cloned, or every thread ends up
// locking its own private copy and the lock stops meaning anything.
type MutexObject struct {
	core.BuiltInObject
	mu      sync.Mutex
	Methods map[int]*core.BuiltInObject
}

func newMutexObject() *MutexObject {
	return &MutexObject{}
}

func (o *MutexObject) String() string {
	return "<mutex>"
}

func (o *MutexObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *MutexObject) GetNativeType() core.NativeType {
	return core.NATIVE_MUTEX
}

func (o *MutexObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *MutexObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *MutexObject) IsBuiltIn() bool {
	return true
}
