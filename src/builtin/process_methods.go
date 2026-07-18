package builtin

import (
	"os/exec"

	"glox/src/core"
)

// RegisterAllProcessMethods wires up a ProcessObject's Lox-visible methods.
// isChild selects the "parent channel" variant (constructed by
// ParentBuiltIn): it has no underlying *exec.Cmd to manage, so
// wait/kill/pid are simply never registered -- calling them from Lox then
// raises the ordinary "undefined method" error, no special-casing needed.
func RegisterAllProcessMethods(o *ProcessObject, isChild bool) {

	o.RegisterMethod("send", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 1 {
				vm.RunTimeError("send() expects 1 argument")
				return core.NIL_VALUE
			}
			val := vm.Stack(arg_stackptr)
			if err := core.WriteFramedValue(o.Stdin, val); err != nil {
				vm.RunTimeErrorNamed("ProcessError", "send failed: %v", err)
				return core.NIL_VALUE
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
			result := <-o.recvCh
			if result.err != nil {
				vm.RunTimeErrorNamed("ProcessError", "recv failed: %v", result.err)
				return core.NIL_VALUE
			}
			return result.val
		},
	})

	o.RegisterMethod("try_recv", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("try_recv() expects no arguments")
				return core.NIL_VALUE
			}
			select {
			case result := <-o.recvCh:
				if result.err != nil {
					vm.RunTimeErrorNamed("ProcessError", "recv failed: %v", result.err)
					return core.NIL_VALUE
				}
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(true, false), result.val}, true)
				return core.MakeObjectValue(tuple, false)
			default:
				tuple := core.MakeListObject([]core.Value{core.MakeBooleanValue(false, false), core.NIL_VALUE}, true)
				return core.MakeObjectValue(tuple, false)
			}
		},
	})

	if isChild {
		return
	}

	o.RegisterMethod("wait", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("wait() expects no arguments")
				return core.NIL_VALUE
			}
			err := o.Cmd.Wait()
			if err != nil {
				if _, isExit := err.(*exec.ExitError); !isExit {
					vm.RunTimeErrorNamed("ProcessError", "wait failed: %v", err)
					return core.NIL_VALUE
				}
			}
			code := 0
			if o.Cmd.ProcessState != nil {
				code = o.Cmd.ProcessState.ExitCode()
			}
			return core.MakeIntValue(code, false)
		},
	})

	o.RegisterMethod("kill", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("kill() expects no arguments")
				return core.NIL_VALUE
			}
			if err := o.Cmd.Process.Kill(); err != nil {
				vm.RunTimeErrorNamed("ProcessError", "kill failed: %v", err)
				return core.NIL_VALUE
			}
			return core.NIL_VALUE
		},
	})

	o.RegisterMethod("pid", &core.BuiltInObject{
		Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
			if argCount != 0 {
				vm.RunTimeError("pid() expects no arguments")
				return core.NIL_VALUE
			}
			return core.MakeIntValue(o.Cmd.Process.Pid, false)
		},
	})
}
