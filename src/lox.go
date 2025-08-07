package lox

import "glox/src/vm"

// Re-export VM types and constants from the vm package
type VM = vm.VM
type InterpretResult = vm.InterpretResult
type VMRunMode = vm.VMRunMode

const (
	INTERPRET_OK            = vm.INTERPRET_OK
	INTERPRET_COMPILE_ERROR = vm.INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR = vm.INTERPRET_RUNTIME_ERROR

	RUN_TO_COMPLETION    = vm.RUN_TO_COMPLETION
	RUN_CURRENT_FUNCTION = vm.RUN_CURRENT_FUNCTION
)

// Re-export VM constructor and functions
var NewVM = vm.NewVM
