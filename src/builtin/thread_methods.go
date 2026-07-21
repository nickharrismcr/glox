package builtin

import "glox/src/core"

// RegisterAllThreadMethods wires up the parent-side Thread object's
// Lox-visible methods (send/recv/try_recv/wait/cancel).
func RegisterAllThreadMethods(o *ThreadObject) {

	o.RegisterMethod("send", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("send() expects 1 argument")
				return core.NIL_VALUE
			}
			val := vm.Stack(arg_stackptr)
			select {
			case o.Handle.ToWorker <- val:
			case <-o.Handle.Done:
				vm.RunTimeErrorNamed("ThreadError", "thread has already finished")
			}
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("recv", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("recv() expects no arguments")
				return core.NIL_VALUE
			}
			msg, ok := <-o.Handle.FromWorker
			if !ok {
				vm.RunTimeErrorNamed("ThreadError", "thread finished, no more messages")
				return core.NIL_VALUE
			}
			if msg.Err != nil {
				vm.RunTimeErrorNamed("ThreadError", "%v", msg.Err)
				return core.NIL_VALUE
			}
			return msg.Val
		},
	})

	o.RegisterMethod("try_recv", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("try_recv() expects no arguments")
				return core.NIL_VALUE
			}
			select {
			case msg, ok := <-o.Handle.FromWorker:
				if !ok {
					vm.RunTimeErrorNamed("ThreadError", "thread finished, no more messages")
					return core.NIL_VALUE
				}
				if msg.Err != nil {
					vm.RunTimeErrorNamed("ThreadError", "%v", msg.Err)
					return core.NIL_VALUE
				}
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(true, false), msg.Val}, true)
				return core.MakeObjectValue(tuple, false)
			default:
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(false, false), core.NIL_VALUE}, true)
				return core.MakeObjectValue(tuple, false)
			}
		},
	})

	o.RegisterMethod("wait", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("wait() expects no arguments")
				return core.NIL_VALUE
			}
			<-o.Handle.Done
			if o.Handle.Err != nil {
				vm.RunTimeErrorNamed("ThreadError", "%v", o.Handle.Err)
				return core.NIL_VALUE
			}
			return o.Handle.Result
		},
	})

	o.RegisterMethod("cancel", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("cancel() expects no arguments")
				return core.NIL_VALUE
			}
			// Cooperative only: unblocks a worker parked in channel()'s
			// send/recv, but cannot interrupt one stuck in a tight
			// non-channel loop -- there's no instrumented preemption point
			// in run(). See docs/thread-module-plan.md.
			o.Handle.Cancel()
			return core.NIL_VALUE
		},
	})
}

// RegisterAllThreadChannelMethods wires up the worker-side ThreadChannel
// object's Lox-visible methods (send/recv/try_recv), each selecting
// against Cancelled instead of Done -- the worker's own signal that the
// parent called cancel().
func RegisterAllThreadChannelMethods(o *ThreadChannelObject) {

	o.RegisterMethod("send", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("send() expects 1 argument")
				return core.NIL_VALUE
			}
			val := vm.Stack(arg_stackptr)
			select {
			case o.Chans.Out <- core.ThreadMessage{Val: val}:
			case <-o.Chans.Cancelled:
				vm.RunTimeErrorNamed("ThreadError", "thread was cancelled")
			}
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("recv", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("recv() expects no arguments")
				return core.NIL_VALUE
			}
			select {
			case val, ok := <-o.Chans.In:
				if !ok {
					vm.RunTimeErrorNamed("ThreadError", "parent closed the channel")
					return core.NIL_VALUE
				}
				return val
			case <-o.Chans.Cancelled:
				vm.RunTimeErrorNamed("ThreadError", "thread was cancelled")
				return core.NIL_VALUE
			}
		},
	})

	o.RegisterMethod("try_recv", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("try_recv() expects no arguments")
				return core.NIL_VALUE
			}
			select {
			case val, ok := <-o.Chans.In:
				if !ok {
					vm.RunTimeErrorNamed("ThreadError", "parent closed the channel")
					return core.NIL_VALUE
				}
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(true, false), val}, true)
				return core.MakeObjectValue(tuple, false)
			case <-o.Chans.Cancelled:
				vm.RunTimeErrorNamed("ThreadError", "thread was cancelled")
				return core.NIL_VALUE
			default:
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(false, false), core.NIL_VALUE}, true)
				return core.MakeObjectValue(tuple, false)
			}
		},
	})
}
