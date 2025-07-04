package debug

import (
	"glox/src/core"
)

// VMInspector defines the interface for debug access to the VM
// Only the methods needed for debugging are exposed
// Implemented by *VM in vm.go

type VMInspector interface {
	ShowStack() string
	ShowGlobals() string
	Frame() *core.CallFrame
	FrameAt(depth int) *core.CallFrame
	FrameCount() int
	Script() string
	CurrCode() uint8
}

func TraceOpcode(vm VMInspector) {
	if vm == nil || vm.Frame() == nil {
		return
	}
	core.Log(core.TRACE, "-----------------------------------------------------")
	if core.DebugShowGlobals {
		core.LogFmt(core.TRACE, "Globals:\n%s\n", vm.ShowGlobals())
	}
	core.LogFmt(core.TRACE, "Stack:\n%s\n", vm.ShowStack())
	chunk := vm.Frame().Closure.Function.Chunk
	function := vm.Frame().Closure.Function
	name := function.Name.Get()
	script := vm.Script()
	code := vm.CurrCode()
	depth := vm.Frame().Depth
	offset := vm.Frame().Ip
	_ = DisassembleInstruction(chunk, script, name, depth, code, offset)

}

func TraceCall(vm VMInspector, data any) {

	closure := data.(*core.ClosureObject)
	core.LogFmt(core.TRACE, "Call: %s\n", closure.Function.Name.Get())
}

func TraceReturn(vm VMInspector, data any) {

	value := data.(core.Value)
	core.LogFmt(core.TRACE, "Return: %s\n", value.String())

}

func TraceHook(vmContext interface{}, event core.DebugEvent, data any) {

	vm, _ := vmContext.(VMInspector)
	switch event {
	case core.DebugEventOpcode:
		TraceOpcode(vm)
	case core.DebugEventCall:
		TraceCall(vm, data)
	case core.DebugEventReturn:
		TraceReturn(vm, data)
	}
}

var InstructionCount int

func InstrumentHook(vmContext interface{}, event core.DebugEvent, data any) {

	switch event {
	case core.DebugEventOpcode:
		InstructionCount += 1
	}
}
