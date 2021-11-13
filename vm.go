package main

import (
	"fmt"
	"os"
)

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

func (vm *VM) runTimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprint(os.Stderr, "\n")
	line := vm.chunk.lines[vm.ip-1]
	fmt.Fprintf(os.Stderr, "[line %d] in script \n", line)
	vm.resetStack()

}

func (vm *VM) push(v Value) {
	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

func (vm *VM) pop() Value {
	if vm.stackTop == 0 {
		return MakeNilValue()
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) isFalsey(v Value) bool {
	switch v.(type) {
	case NumberValue:
		return v.(NumberValue).Get() == 0
	case NilValue:
		return true
	case BooleanValue:
		return !v.(BooleanValue).Get()
	}
	return true
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

Loop:
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
			if !vm.unaryNegate() {
				break Loop
			}
		case OP_ADD:
			if !vm.binaryAdd() {
				break Loop
			}
		case OP_SUBTRACT:
			if !vm.binarySubtract() {
				break Loop
			}
		case OP_MULTIPLY:
			if !vm.binaryMultiply() {
				break Loop
			}
		case OP_DIVIDE:
			if !vm.binaryDivide() {
				break Loop
			}
		case OP_NIL:
			vm.push(MakeNilValue())
		case OP_TRUE:
			vm.push(MakeBooleanValue(true))
		case OP_FALSE:
			vm.push(MakeBooleanValue(false))
		case OP_NOT:
			v := vm.pop()
			bv := vm.isFalsey(v)
			vm.push(MakeBooleanValue(bv))
		case OP_EQUAL:
			a := vm.pop()
			b := vm.pop()
			vm.push(MakeBooleanValue(valuesEqual(a, b)))
		case OP_GREATER:
			if !vm.binaryGreater() {
				break Loop
			}
		case OP_LESS:
			if !vm.binaryLess() {
				break Loop
			}
		case OP_PRINT:
			v := vm.pop()
			fmt.Printf("%s\n", v.String())
		case OP_POP:
			_ = vm.pop()
		default:
			vm.runTimeError("Invalid Opcode")
			break Loop
		}
	}
	return INTERPRET_RUNTIME_ERROR, MakeNilValue()
}

func (vm *VM) binaryAdd() bool {
	v2 := vm.pop()
	switch v2.(type) {
	case NumberValue:
		nv2, _ := v2.(NumberValue)
		v1 := vm.pop()
		nv1, ok := v1.(NumberValue)
		if !ok {
			vm.runTimeError("Addition type mismatch")
			return false
		}
		vm.push(MakeNumberValue(nv1.Get() + nv2.Get()))
		return true
	case ObjectValue:
		ov2 := v2.(ObjectValue).value
		if ov2.getType() == OBJECT_STRING {
			v1 := vm.pop()
			o1, ok := v1.(ObjectValue)
			if !ok {
				vm.runTimeError("Addition type mismatch")
				return false
			}
			ov1 := o1.value
			if ov1.getType() == OBJECT_STRING {
				vm.concatenate(ov1.String(), ov2.String())
				return true
			}

		}
	}
	vm.runTimeError("Operands must be numbers or strings")
	return false
}

func (vm *VM) binarySubtract() bool {
	v2 := vm.pop()
	nv2, ok := v2.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	v1 := vm.pop()
	nv1, ok := v1.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(MakeNumberValue(nv1.Get() - nv2.Get()))
	return true
}

func (vm *VM) binaryMultiply() bool {
	v2 := vm.pop()
	nv2, ok := v2.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	v1 := vm.pop()
	nv1, ok := v1.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(MakeNumberValue(nv1.Get() * nv2.Get()))
	return true
}

func (vm *VM) binaryDivide() bool {
	v2 := vm.pop()
	nv2, ok := v2.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	v1 := vm.pop()
	nv1, ok := v1.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(MakeNumberValue(nv1.Get() / nv2.Get()))
	return true
}

func (vm *VM) unaryNegate() bool {
	v := vm.pop()
	nv, ok := v.(NumberValue)
	if !ok {
		vm.runTimeError("Operand must be a number")
		return false
	}
	f := nv.Get()
	vm.push(MakeNumberValue(-f))
	return true
}

func (vm *VM) binaryGreater() bool {
	v2 := vm.pop()
	nv2, ok := v2.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	v1 := vm.pop()
	nv1, ok := v1.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(MakeBooleanValue(nv1.Get() > nv2.Get()))
	return true
}

func (vm *VM) binaryLess() bool {
	v2 := vm.pop()
	nv2, ok := v2.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	v1 := vm.pop()
	nv1, ok := v1.(NumberValue)
	if !ok {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(MakeBooleanValue(nv1.Get() < nv2.Get()))
	return true
}

func (vm *VM) concatenate(s1, s2 string) {

	so := MakeStringObject(s1 + s2)
	vm.push(MakeObjectValue(so))
}
