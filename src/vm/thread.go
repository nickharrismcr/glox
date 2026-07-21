package vm

import (
	"context"
	"fmt"

	"glox/src/core"
)

// thread.go implements the VMContext primitives backing the thread module
// (see docs/thread-module-plan.md): SpawnThread runs a closure on a new
// goroutine-backed VM, ThreadChannels lets that worker reach its own
// communication channels, and CallClosure synchronously invokes a closure
// on the *current* VM (no goroutine) for use by thread.spawn's worker body
// and by sync.Mutex.locked().

// channelBufferSize matches process.ProcessObject's recvCh buffer
// (src/builtin/obj_builtin_process.go) for consistency between the two
// worker models.
const channelBufferSize = 16

// SpawnThread deep-copies closure and args (so the new thread shares no
// mutable captured state with the caller -- see core.CopyClosureForSpawn),
// then runs the copy on a fresh *VM in its own goroutine.
func (vm *VM) SpawnThread(closureVal core.Value, args []core.Value) (*core.ThreadHandle, error) {
	if vm.Repl {
		// The REPL's Environment.GrowGlobals reallocates Globals/Defined on
		// every new line; run() caches a slice header from them per frame
		// push. A thread still running when the user enters another REPL
		// line would be working against a stale backing array -- simplest
		// correct fix is disallowing spawn() from the REPL entirely.
		return nil, fmt.Errorf("thread.spawn() is not supported from the REPL")
	}
	closure, ok := closureVal.Obj.(*core.ClosureObject)
	if closureVal.Type != core.VAL_OBJ || !ok {
		return nil, fmt.Errorf("thread.spawn() argument must be a function")
	}

	memo := map[core.Object]core.Object{}
	copiedClosure := core.CopyClosureForSpawn(closure, memo)
	copiedArgs := make([]core.Value, len(args))
	for i, a := range args {
		copiedArgs[i] = core.CopyValueForSpawn(a, memo)
	}

	toWorker := make(chan core.Value, channelBufferSize)
	fromWorker := make(chan core.ThreadMessage, channelBufferSize)
	doneCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	worker := NewVM(vm.script, false)
	worker.BuiltIns = vm.BuiltIns
	worker.BuiltInModules = vm.BuiltInModules
	worker.SetArgs(vm.Args())
	worker.threadChans = &core.ThreadChannels{
		In:        toWorker,
		Out:       fromWorker,
		Cancelled: ctx.Done(),
	}

	handle := &core.ThreadHandle{
		ToWorker:   toWorker,
		FromWorker: fromWorker,
		Done:       doneCh,
		Cancel:     cancel,
	}

	go runThreadWorker(worker, copiedClosure, copiedArgs, handle, fromWorker, doneCh)

	return handle, nil
}

// runThreadWorker is the spawned goroutine's body. Mirrors
// callLoadedChunk's push+call+run shape (a fresh VM, no outer frames), but
// wrapped in its own recover() -- mandatory, since nothing else catches a
// panic on a non-main goroutine (main.go's runFile recover only guards the
// one calling goroutine).
func runThreadWorker(worker *VM, closure *core.ClosureObject, args []core.Value,
	handle *core.ThreadHandle, fromWorker chan core.ThreadMessage, doneCh chan struct{}) {

	var workErr error
	var result core.Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				workErr = fmt.Errorf("thread panicked: %v", r)
			}
		}()
		worker.push(core.MakeObjectValue(closure, false))
		for _, a := range args {
			worker.push(a)
		}
		if !worker.call(closure, len(args)) {
			workErr = fmt.Errorf("%s", worker.ErrorMsg)
			return
		}
		res, retVal := worker.run(RUN_TO_COMPLETION)
		if res != INTERPRET_OK {
			workErr = fmt.Errorf("%s", worker.ErrorMsg)
			return
		}
		result = retVal
	}()

	// A cancelled thread ending -- whether via the uncaught ThreadError its
	// own channel().send()/recv() raises on seeing Cancelled fire, or it
	// just happening to finish independently around the same time
	// cancel() was called -- is an expected, self-inflicted shutdown, not
	// a fault. Mirrors process.kill() producing a clean EOF for
	// process.wait_any rather than a ProcessError: cancel() is glox's
	// equivalent of kill(), so it must look the same way to wait()/
	// recv()/wait_any() once it's taken effect.
	cancelled := false
	select {
	case <-worker.threadChans.Cancelled:
		cancelled = true
	default:
	}

	// Both fields are written here, before fromWorker/doneCh are closed --
	// Go's channel-close happens-before guarantee is what makes it safe
	// for wait()/recv() to read them after observing either channel
	// closed, with no mutex needed.
	if workErr != nil && !cancelled {
		handle.Err = workErr
		select {
		case fromWorker <- core.ThreadMessage{Err: workErr}:
		default: // recv() isn't listening right now: wait() is still authoritative
		}
	} else if workErr == nil {
		handle.Result = result
	}
	close(fromWorker)
	close(doneCh)
}

// ThreadChannels returns this VM's own communication channels -- ok is
// false unless this VM was itself created by SpawnThread.
func (vm *VM) ThreadChannels() (*core.ThreadChannels, bool) {
	return vm.threadChans, vm.threadChans != nil
}

// CallClosure synchronously invokes closure on this VM: no new VM, no
// copy, no goroutine. Safe to call from inside any native builtin's Go
// function body -- it uses RUN_CURRENT_FUNCTION, which runs only until the
// newly-pushed frame returns (see run()'s startFrame check), leaving the
// caller's own frames/stack untouched, and OP_CALL's handler already
// unconditionally calls refreshFrame() after any call returns (native or
// closure), so a builtin invoking this mid-dispatch is safe by
// construction.
func (vm *VM) CallClosure(closureVal core.Value, args []core.Value) (core.Value, error) {
	closure, ok := closureVal.Obj.(*core.ClosureObject)
	if closureVal.Type != core.VAL_OBJ || !ok {
		return core.NIL_VALUE, fmt.Errorf("expected a function, got %s", closureVal.String())
	}
	vm.push(closureVal)
	for _, a := range args {
		vm.push(a)
	}
	if !vm.call(closure, len(args)) {
		return core.NIL_VALUE, fmt.Errorf("%s", vm.ErrorMsg)
	}
	res, retVal := vm.run(RUN_CURRENT_FUNCTION)
	if res != INTERPRET_OK {
		return core.NIL_VALUE, fmt.Errorf("%s", vm.ErrorMsg)
	}
	return retVal, nil
}
