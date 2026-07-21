package builtin

import "glox/src/core"

// MutexBuiltIn implements sync.Mutex(), constructing a new, unlocked mutex.
func MutexBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("Mutex() expects no arguments.")
		return core.NIL_VALUE
	}
	mutexObj := newMutexObject()
	RegisterAllMutexMethods(mutexObj)
	return core.MakeObjectValue(mutexObj, true)
}
