package main

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	chunk    *Chunk
	ip       int
	stack    [256]Value
	stackTop int
}

func NewVM() *VM {
	vm := &VM{}
	vm.resetStack()
	return vm
}

func (vm *VM) resetStack() {
	vm.stackTop = 0
}

func (vm *VM) push(v Value) {
	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

func (vm *VM) pop() Value {
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) interpret(source string) (InterpretResult, string) {

	vm.chunk = NewChunk()
	if !vm.compile(source) {
		return INTERPRET_COMPILE_ERROR, ""
	}
	vm.ip = 0
	res, val := vm.run()
	return res, val.String()
}

func (vm *VM) run() (InterpretResult, Value) {

	for {

		inst := vm.chunk.code[vm.ip]

		if debugTraceExecution {
			vm.stackTrace()
			_ = vm.chunk.disassembleInstruction(inst, vm.ip)
		}

		vm.ip++
		switch inst {
		case OP_RETURN:
			v := vm.pop()
			return INTERPRET_OK, v
		case OP_CONSTANT:
			idx := vm.chunk.code[vm.ip]
			vm.ip++
			constant := vm.chunk.constants[idx]
			vm.push(constant)
		case OP_NEGATE:
			v := vm.pop()
			f := v.(NumberValue).Get()
			vm.push(NumberValue{value: -f})
		case OP_ADD:
			vm.binary_add()
		case OP_SUBTRACT:
			vm.binary_subtract()
		case OP_MULTIPLY:
			vm.binary_multiply()
		case OP_DIVIDE:
			vm.binary_divide()
		}
	}
	return INTERPRET_RUNTIME_ERROR, NumberValue{}
}

func (vm *VM) binary_add() {
	v2 := vm.pop()
	v1 := vm.pop()
	f1 := v1.(NumberValue).Get()
	f2 := v2.(NumberValue).Get()
	vm.push(NumberValue{value: f1 + f2})
}
func (vm *VM) binary_subtract() {
	v2 := vm.pop()
	v1 := vm.pop()
	f1 := v1.(NumberValue).Get()
	f2 := v2.(NumberValue).Get()
	vm.push(NumberValue{value: f1 - f2})
}
func (vm *VM) binary_multiply() {
	v2 := vm.pop()
	v1 := vm.pop()
	f1 := v1.(NumberValue).Get()
	f2 := v2.(NumberValue).Get()
	vm.push(NumberValue{value: f1 * f2})
}
func (vm *VM) binary_divide() {
	v2 := vm.pop()
	v1 := vm.pop()
	f1 := v1.(NumberValue).Get()
	f2 := v2.(NumberValue).Get()
	vm.push(NumberValue{value: f1 / f2})
}
