package builtin

import (
	"io"
	"os/exec"

	"glox/src/core"
)

// recvResult is what the background reader goroutine pushes onto a
// ProcessObject's recvCh: either a decoded value or the error that ended
// the read loop (EOF when the peer closed its end, or an I/O error).
type recvResult struct {
	val core.Value
	err error
}

// ProcessObject represents one end of a pipe-backed channel carrying
// length-prefixed pickled values (see core.WriteFramedValue/ReadFramedValue).
// It doubles as both "the process I spawned" (Cmd != nil, exposes
// wait/kill/pid) and "the channel back to whoever spawned me" (Cmd == nil,
// constructed by ParentBuiltIn wrapping this process's own stdin/stdout).
type ProcessObject struct {
	core.BuiltInObject
	Cmd     *exec.Cmd // nil for the parent-channel variant
	Stdin   io.WriteCloser
	recvCh  chan recvResult
	Methods map[int]*core.BuiltInObject
}

// newProcessObject starts the background reader goroutine (one
// ReadFramedValue call after another, pushed onto recvCh) and returns the
// constructed object. The goroutine exits after pushing the first error
// (EOF or otherwise) it hits.
func newProcessObject(stdin io.WriteCloser, stdout io.ReadCloser, cmd *exec.Cmd) *ProcessObject {
	o := &ProcessObject{
		Cmd:    cmd,
		Stdin:  stdin,
		recvCh: make(chan recvResult, 16),
	}
	go func() {
		for {
			val, err := core.ReadFramedValue(stdout)
			o.recvCh <- recvResult{val: val, err: err}
			if err != nil {
				return
			}
		}
	}()
	return o
}

func (o *ProcessObject) String() string {
	return "<process>"
}

func (o *ProcessObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *ProcessObject) GetNativeType() core.NativeType {
	return core.NATIVE_PROCESS
}

func (o *ProcessObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *ProcessObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *ProcessObject) IsBuiltIn() bool {
	return true
}
