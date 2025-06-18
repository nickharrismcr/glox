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

type VMForeachStage int

const (
	WAITING_FOR_ITER VMForeachStage = iota
	WAITING_FOR_NEXT
	DONE
)

type VMForeachState struct {
	LocalSlot   int
	IterSlot    int
	JumpToStart int
	JumpToEnd   int
	Stage       VMForeachStage
	Prev        *VMForeachState
}
