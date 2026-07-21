package builtin

import (
	"reflect"

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

// ThreadWaitAnyBuiltIn blocks until any one of the given Thread objects
// has a message ready, mirroring process.wait_any's fan-in shape exactly
// (same reflect.Select-over-a-dynamic-count-of-channels approach, same
// "drop a finished one instead of erroring, return nil once every thread
// is done" semantics) but selecting over each Thread's
// Handle.FromWorker instead of a Process's recvCh.
//
// A message with Err == nil is a value the thread sent via
// channel().send(); returned as the tuple (index, value). A message with
// Err != nil is the thread's terminal "ended abnormally" notice (a
// recovered panic or an unhandled Lox exception) and raises ThreadError
// immediately, matching recv()'s behaviour. The channel closing with no
// message (ok == false) is a clean, normal return with nothing pending on
// the channel -- not an error -- so that thread is simply dropped from
// consideration, the same as a Process finishing cleanly. Once every
// thread in the list has finished, wait_any returns nil.
func ThreadWaitAnyBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("wait_any() expects 1 argument (a list of threads).")
		return core.NIL_VALUE
	}

	listVal := vm.Stack(arg_stackptr)
	if listVal.Type != core.VAL_OBJ || listVal.Obj.GetType() != core.OBJECT_LIST {
		vm.RunTimeError("wait_any() argument must be a list of threads.")
		return core.NIL_VALUE
	}
	list := listVal.AsList()
	if len(list.Items) == 0 {
		vm.RunTimeError("wait_any() list must not be empty.")
		return core.NIL_VALUE
	}

	threads := make([]*ThreadObject, len(list.Items))
	for i, item := range list.Items {
		threadObj, ok := item.Obj.(*ThreadObject)
		if !ok {
			vm.RunTimeError("wait_any() list must contain only thread objects.")
			return core.NIL_VALUE
		}
		threads[i] = threadObj
	}

	live := make([]int, 0, len(threads))
	for i, t := range threads {
		if !t.recvDone {
			live = append(live, i)
		}
	}

	for len(live) > 0 {
		cases := make([]reflect.SelectCase, len(live))
		for i, origIdx := range live {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(threads[origIdx].Handle.FromWorker)}
		}

		chosen, recv, ok := reflect.Select(cases)
		origIdx := live[chosen]

		if !ok {
			threads[origIdx].recvDone = true
			live = append(live[:chosen], live[chosen+1:]...)
			continue
		}

		msg := recv.Interface().(core.ThreadMessage)
		if msg.Err != nil {
			vm.RunTimeErrorNamed("ThreadError", "thread %d: %v", origIdx, msg.Err)
			return core.NIL_VALUE
		}

		tuple := core.MakeListObject([]core.Value{core.MakeIntValue(origIdx, false), msg.Val}, true)
		return core.MakeObjectValue(tuple, false)
	}

	return core.NIL_VALUE
}
