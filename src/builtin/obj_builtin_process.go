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

	// recvDone latches once wait_any has observed a clean io.EOF on this
	// process's recvCh (see WaitAnyBuiltIn). The reader goroutine posts
	// exactly one value per ReadFramedValue call and never posts again
	// after an error, so once its terminal EOF has been consumed the
	// channel will never become ready again -- selecting on it a second
	// time (e.g. from a later, separate wait_any call over a list that
	// still includes this process) would block forever. recvDone lets
	// wait_any recognise and skip an already-finished process up front
	// instead of adding a permanently-dead case to its select, which
	// previously could make every case dead at once and hang the whole
	// process in an unrecoverable Go runtime deadlock. Only ever touched
	// from WaitAnyBuiltIn on the single VM goroutine, so it needs no
	// synchronisation.
	recvDone bool
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
