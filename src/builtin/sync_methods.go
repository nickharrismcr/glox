package builtin

import "glox/src/core"

// RegisterAllMutexMethods wires up a MutexObject's Lox-visible methods.
func RegisterAllMutexMethods(o *MutexObject) {

	o.RegisterMethod("acquire", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("acquire() expects no arguments")
				return core.NIL_VALUE
			}
			o.mu.Lock()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("release", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) (result core.Value) {
			if argCount != 0 {
				vm.RunTimeError("release() expects no arguments")
				return core.NIL_VALUE
			}
			// Go's sync.Mutex.Unlock() panics if called without a matching
			// Lock() -- recover and surface it as a catchable SyncError
			// instead of crashing the calling thread's goroutine.
			defer func() {
				if r := recover(); r != nil {
					vm.RunTimeErrorNamed("SyncError", "release() without a matching acquire(): %v", r)
					result = core.NIL_VALUE
				}
			}()
			o.mu.Unlock()
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("locked", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("locked() expects 1 argument (a function)")
				return core.NIL_VALUE
			}
			closureVal := vm.Stack(arg_stackptr)
			o.mu.Lock()
			defer o.mu.Unlock()
			result, err := vm.CallClosure(closureVal, nil)
			if err != nil {
				vm.RunTimeErrorNamed("SyncError", "%v", err)
				return core.NIL_VALUE
			}
			return result
		},
	})
}
