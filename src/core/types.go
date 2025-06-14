package core

type CallFrame struct {
	Closure  *ClosureObject
	Ip       int
	Slots    int // start of vm stack for this frame
	Handlers *ExceptionHandler
	Depth    int
}

type ExceptionHandler struct {
	ExceptIP uint16
	StackTop int
	Prev     *ExceptionHandler
}
