package builtin

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"

	"glox/src/core"
)

// resolveSpawnExecutable returns the path to relaunch as a child process.
// On Windows, os.Executable() can return an extension-less path (e.g. the
// "bin/glox" this project's own build command produces per CLAUDE.md) --
// but Windows' process creation silently prefers a same-directory file with
// an extension appended (".exe" etc.) over an extension-less path, even
// when that path is passed explicitly (confirmed: setting exec.Cmd.Path
// directly does not avoid it). If some older/unrelated "glox.exe" happens
// to sit next to the real binary, spawn() would then silently relaunch
// that stale build instead of itself. Rather than risk that, copy this
// process's own binary to a freshly created, uniquely-named ".exe" once
// per process and relaunch from there -- no ambiguity, no stale sibling to
// collide with. The copy is created in the *same directory* as the real
// binary (not the OS temp dir), because raylib's DLL is loaded from the
// executable's own directory at startup -- a copy elsewhere would fail to
// find it. Cost (a one-time binary copy) is paid at most once per parent
// process, not per spawn() call.
var (
	windowsExeOnce sync.Once
	windowsExePath string
	windowsExeErr  error
)

func resolveSpawnExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" || filepath.Ext(exe) != "" {
		return exe, nil
	}
	windowsExeOnce.Do(func() {
		windowsExePath, windowsExeErr = copySelfToTempExe(exe)
	})
	return windowsExePath, windowsExeErr
}

func copySelfToTempExe(exe string) (string, error) {
	src, err := os.Open(exe)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.CreateTemp(filepath.Dir(exe), "glox-relaunch-*.exe")
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return dst.Name(), nil
}

// SpawnBuiltIn launches another glox process running scriptPath, wired to
// the caller by a pipe pair on the child's stdin/stdout (see
// newProcessObject). Extra string arguments become the child's own
// sys.args, exactly as if it had been invoked directly from the CLI --
// see docs/md/PROCESS_MODULE.md for the "extra args must not start with -"
// caveat (main.go's flag parser has no `--` escape hatch).
func SpawnBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount < 1 {
		vm.RunTimeError("spawn() requires at least 1 argument (script path).")
		return core.NIL_VALUE
	}

	scriptVal := vm.Stack(arg_stackptr)
	if scriptVal.Type != core.VAL_OBJ || scriptVal.Obj.GetType() != core.OBJECT_STRING {
		vm.RunTimeError("spawn() first argument must be a string (script path).")
		return core.NIL_VALUE
	}
	scriptPath := scriptVal.AsString().Get()

	var extraArgs []string
	for i := 1; i < argCount; i++ {
		argVal := vm.Stack(arg_stackptr + i)
		if argVal.Type != core.VAL_OBJ || argVal.Obj.GetType() != core.OBJECT_STRING {
			vm.RunTimeError("spawn() extra arguments must be strings.")
			return core.NIL_VALUE
		}
		extraArgs = append(extraArgs, argVal.AsString().Get())
	}

	exe, err := resolveSpawnExecutable()
	if err != nil {
		vm.RunTimeErrorNamed("ProcessError", "failed to resolve glox executable: %v", err)
		return core.NIL_VALUE
	}

	cmdArgs := append([]string{scriptPath}, extraArgs...)
	cmd := exec.Command(exe, cmdArgs...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		vm.RunTimeErrorNamed("ProcessError", "failed to open stdin pipe: %v", err)
		return core.NIL_VALUE
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		vm.RunTimeErrorNamed("ProcessError", "failed to open stdout pipe: %v", err)
		return core.NIL_VALUE
	}

	if err := cmd.Start(); err != nil {
		vm.RunTimeErrorNamed("ProcessError", "failed to start process %q: %v", scriptPath, err)
		return core.NIL_VALUE
	}

	procObj := newProcessObject(stdin, stdout, cmd)
	RegisterAllProcessMethods(procObj, false)
	return core.MakeObjectValue(procObj, true)
}

// ParentBuiltIn returns a Process object wired to this process's own
// stdin/stdout, the far ends of the pipes SpawnBuiltIn set up in whichever
// process spawned this one -- swapped relative to the spawn() side (this
// process writes to its own stdout to reach its parent, reads its own
// stdin to receive from it).
func ParentBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 0 {
		vm.RunTimeError("parent() expects no arguments.")
		return core.NIL_VALUE
	}
	procObj := newProcessObject(os.Stdout, os.Stdin, nil)
	RegisterAllProcessMethods(procObj, true)
	return core.MakeObjectValue(procObj, true)
}

// WaitAnyBuiltIn blocks until any one of the given Process objects has a
// message ready, using reflect.Select for a dynamic-count select over their
// recvCh channels (Go's select statement needs a fixed case count at
// compile time). Returns the tuple (index, value).
//
// A process whose peer closed its pipe cleanly (io.EOF -- the child script
// simply ran to completion) is not a fatal error for the wait as a whole:
// it's dropped from consideration and the select retries among whatever
// processes are still live, so one finished worker in a pool can't abort
// the fan-in wait for the others. Only once every process in the list has
// closed does wait_any raise, and a genuine I/O error (a broken pipe, a
// truncated frame -- anything other than a clean io.EOF) still raises
// immediately, since that's not an expected "the worker is done" event.
func WaitAnyBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("wait_any() expects 1 argument (a list of processes).")
		return core.NIL_VALUE
	}

	listVal := vm.Stack(arg_stackptr)
	if listVal.Type != core.VAL_OBJ || listVal.Obj.GetType() != core.OBJECT_LIST {
		vm.RunTimeError("wait_any() argument must be a list of processes.")
		return core.NIL_VALUE
	}
	list := listVal.AsList()
	if len(list.Items) == 0 {
		vm.RunTimeError("wait_any() list must not be empty.")
		return core.NIL_VALUE
	}

	procs := make([]*ProcessObject, len(list.Items))
	for i, item := range list.Items {
		procObj, ok := item.Obj.(*ProcessObject)
		if !ok {
			vm.RunTimeError("wait_any() list must contain only process objects.")
			return core.NIL_VALUE
		}
		procs[i] = procObj
	}

	// live holds original-list indices still worth selecting on. Anything
	// already latched as recvDone from an earlier wait_any call is
	// excluded up front -- its channel will never be ready again, so
	// including it here would risk a select whose every case is
	// permanently dead (an unrecoverable Go runtime deadlock, not a
	// catchable Lox exception) once every process has finished across
	// separate calls.
	live := make([]int, 0, len(procs))
	for i, p := range procs {
		if !p.recvDone {
			live = append(live, i)
		}
	}

	for len(live) > 0 {
		cases := make([]reflect.SelectCase, len(live))
		for i, origIdx := range live {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(procs[origIdx].recvCh)}
		}

		chosen, recv, _ := reflect.Select(cases)
		result := recv.Interface().(recvResult)
		origIdx := live[chosen]

		if result.err != nil {
			if result.err == io.EOF {
				procs[origIdx].recvDone = true
				live = append(live[:chosen], live[chosen+1:]...)
				continue
			}
			vm.RunTimeErrorNamed("ProcessError", "process %d: %v", origIdx, result.err)
			return core.NIL_VALUE
		}

		tuple := core.MakeListObject([]core.Value{core.MakeIntValue(origIdx, false), result.val}, true)
		return core.MakeObjectValue(tuple, false)
	}

	vm.RunTimeErrorNamed("ProcessError", "wait_any(): all processes have finished")
	return core.NIL_VALUE
}
