package builtin

import (
	"glox/src/core"
)

// ThreadSpawnBuiltIn runs a closure on a new goroutine-backed VM (see
// vm.SpawnThread), glox's answer to Python's threading.Thread -- unlike
// process.spawn(), which needs a .lox script path because there's no way
// to serialise a closure across a process boundary, thread.spawn() can
// take an in-memory function directly since no process boundary is
// crossed. Extra arguments are passed to the closure as its own call
// arguments, both deep-copied first so the new thread shares no mutable
// captured state with the caller.
func ThreadSpawnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount < 1 {
		vm.RunTimeError("spawn() requires at least 1 argument (a function).")
		return core.NIL_VALUE
	}

	closureVal := vm.Stack(arg_stackptr)
	args := make([]core.Value, 0, argCount-1)
	for i := 1; i < argCount; i++ {
		args = append(args, vm.Stack(arg_stackptr+i))
	}

	handle, err := vm.SpawnThread(closureVal, args)
	if err != nil {
		vm.RunTimeErrorNamed("ThreadError", "%v", err)
		return core.NIL_VALUE
	}

	threadObj := newThreadObject(handle)
	RegisterAllThreadMethods(threadObj)
	return core.MakeObjectValue(threadObj, true)
}

// ThreadChannelBuiltIn returns a ThreadChannelObject wired to the
// *current* thread's own communication channels -- called from inside a
// thread.spawn()-ed function to talk back to whoever spawned it, the
// thread-module analogue of process.parent().
func ThreadChannelBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("channel() expects no arguments.")
		return core.NIL_VALUE
	}

	chans, ok := vm.ThreadChannels()
	if !ok {
		vm.RunTimeErrorNamed("ThreadError", "channel() may only be called from inside a thread.spawn()-ed function")
		return core.NIL_VALUE
	}

	chanObj := newThreadChannelObject(chans)
	RegisterAllThreadChannelMethods(chanObj)
	return core.MakeObjectValue(chanObj, true)
}
