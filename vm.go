package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

const (
	FRAMES_MAX int = 64
	STACK_MAX  int = FRAMES_MAX * 256
)

type CallFrame struct {
	function *FunctionObject
	ip       int
	slots    int // start of vm stack
}

type VM struct {
	chunk      *Chunk
	ip         int
	stack      [STACK_MAX]Value
	stackTop   int
	globals    map[string]Value
	frames     [FRAMES_MAX]*CallFrame
	frameCount int
	starttime  time.Time
}

func NewVM() *VM {
	vm := &VM{
		globals:   map[string]Value{},
		starttime: time.Now(),
	}
	vm.resetStack()
	vm.defineNative("clock", clockNative)
	vm.defineNative("str", strNative)
	return vm
}

func (vm *VM) interpret(source string) (InterpretResult, string) {

	function := vm.compile(source)
	if function == nil {
		return INTERPRET_COMPILE_ERROR, ""
	}
	vm.push(makeObjectValue(function, false))
	vm.call(function, 0)
	res, val := vm.run()
	return res, val.String()
}

func (vm *VM) frame() *CallFrame {
	return vm.frames[vm.frameCount-1]
}

func (vm *VM) getCode() []uint8 {
	return vm.frame().function.chunk.code
}

func (vm *VM) resetStack() {
	vm.stackTop = 0
}

func (vm *VM) runTimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprint(os.Stderr, "\n")
	line := vm.frame().function.chunk.lines[vm.frame().ip-1]
	fmt.Fprintf(os.Stderr, "[line %d] in script \n", line)

	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := vm.frames[i]
		function := frame.function
		inst := function.chunk.code[frame.ip]
		fmt.Fprintf(os.Stderr, "[line %d] in ", function.chunk.lines[inst])
		if function.name.String() == "" {
			fmt.Fprintf(os.Stderr, "script \n")
		} else {
			fmt.Fprintf(os.Stderr, function.name.String())
		}
	}

	vm.resetStack()

}
func (vm *VM) defineNative(name string, function NativeFn) {
	vm.push(makeObjectValue(MakeStringObject(name), false))
	vm.push(makeObjectValue(makeNativeObject(function), false))
	vm.globals[name] = vm.stack[1]
	vm.pop()
	vm.pop()
}

func (vm *VM) push(v Value) {
	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

func (vm *VM) pop() Value {
	if vm.stackTop == 0 {
		return makeNilValue()
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) peek(dist int) Value {
	return vm.stack[(vm.stackTop-1)-dist]
}

func (vm *VM) callValue(callee Value, argCount int) bool {
	if ov, ok := callee.(ObjectValue); ok {
		if ov.isFunctionObject() {
			return vm.call(ov.Get().(*FunctionObject), argCount)
		}
		if ov.isNativeFunction() {
			nf := ov.Get().(*NativeObject)
			res := nf.function(argCount, vm.stackTop-argCount, vm)
			vm.push(res)
			return true
		}
	}
	vm.runTimeError("Can only call functions and classes.")
	return false
}

func (vm *VM) call(function *FunctionObject, argCount int) bool {

	if argCount != function.arity {
		vm.runTimeError("Expected %d arguments but got %d.", function.arity, argCount)
		return false
	}
	frame := &CallFrame{
		function: function,
		ip:       0,
		slots:    vm.stackTop - argCount - 1,
	}
	vm.frames[vm.frameCount] = frame
	vm.frameCount++
	if vm.frameCount == FRAMES_MAX {
		vm.runTimeError("Stack overflow.")
		return false
	}

	return true
}

func (vm *VM) readShort() uint16 {
	vm.frame().ip += 2
	b1 := uint16(vm.getCode()[vm.frame().ip-2])
	b2 := uint16(vm.getCode()[vm.frame().ip-1])
	return uint16(b1<<8 | b2)
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

// main interpreter loop
func (vm *VM) run() (InterpretResult, Value) {

	frame := vm.frame()
Loop:
	for {
		inst := vm.getCode()[frame.ip]

		if debugTraceExecution {
			vm.stackTrace()
			_ = frame.function.chunk.disassembleInstruction(inst, frame.ip)
		}

		frame.ip++
		switch inst {

		case OP_RETURN:
			// exit, return the value at stack top
			result := vm.pop()
			vm.frameCount--
			if vm.frameCount == 0 {
				vm.pop() // main script function
				runtime.GC()
				return INTERPRET_OK, result
			}
			vm.stackTop = frame.slots
			vm.push(result)
			frame = vm.frames[vm.frameCount-1]

		case OP_CONSTANT:
			// get the constant indexed by operand 2 and push it onto the stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			constant := frame.function.chunk.constants[idx]
			vm.push(constant)

		case OP_NEGATE:
			// negate the value at stack top
			if !vm.unaryNegate() {
				break Loop
			}

		case OP_ADD:
			// pop 2 stack values, add them and push onto the stack
			if !vm.binaryAdd() {
				break Loop
			}

		case OP_SUBTRACT:
			// pop 2 stack values, subtract and push onto the stack
			if !vm.binarySubtract() {
				break Loop
			}

		case OP_MULTIPLY:
			// pop 2 stack values, multiply and push onto the stack
			if !vm.binaryMultiply() {
				break Loop
			}

		case OP_DIVIDE:
			// pop 2 stack values, divide and push onto the stack
			if !vm.binaryDivide() {
				break Loop
			}

		case OP_NIL:
			// push nil val onto the stack
			vm.push(makeNilValue())

		case OP_TRUE:
			// push true bool val onto the stack
			vm.push(makeBooleanValue(true, false))

		case OP_FALSE:
			// push false bool val onto the stack
			vm.push(makeBooleanValue(false, false))

		case OP_NOT:
			// replace stack top with boolean not of itself
			v := vm.pop()
			bv := vm.isFalsey(v)
			vm.push(makeBooleanValue(bv, false))

		case OP_EQUAL:
			// pop 2 stack values, stack top = boolean
			a := vm.pop()
			b := vm.pop()
			vm.push(makeBooleanValue(valuesEqual(a, b), false))

		case OP_GREATER:
			// pop 2 stack values, stack top = boolean
			if !vm.binaryGreater() {
				break Loop
			}

		case OP_LESS:
			// pop 2 stack values, stack top = boolean
			if !vm.binaryLess() {
				break Loop
			}

		case OP_PRINT:
			// pop 1 stack value and print it
			v := vm.pop()
			fmt.Printf("%s\n", v.String())

		case OP_POP:
			// pop 1 stack value and discard
			_ = vm.pop()

		case OP_DEFINE_GLOBAL:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := frame.function.chunk.constants[idx].String()
			vm.globals[name] = vm.peek(0)
			vm.pop()

		case OP_DEFINE_GLOBAL_CONST:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it and flag as immutable
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := frame.function.chunk.constants[idx].String()
			vm.globals[name] = vm.peek(0)
			vm.globals[name] = immutable(vm.globals[name])
			vm.pop()

		case OP_GET_GLOBAL:
			// name = constant at operand index
			// push globals[name] onto stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := frame.function.chunk.constants[idx].String()
			value, ok := vm.globals[name]
			if !ok {
				vm.runTimeError("Undefined variable %s\n", name)
				break Loop
			}
			vm.push(value)

		case OP_SET_GLOBAL:
			// name = constant at operand index
			// set globals[name] to stack top, key must exist
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := frame.function.chunk.constants[idx].String()
			if _, ok := vm.globals[name]; !ok {
				vm.runTimeError("Undefined variable %s\n", name)
				break Loop
			}
			if vm.globals[name].Immutable() {
				vm.runTimeError("Cannot assign to const %s\n", name)
				break Loop
			}
			vm.globals[name] = vm.peek(0)

		case OP_GET_LOCAL:
			// get local from stack at position = operand and push on stack top
			slot_idx := int(vm.getCode()[frame.ip])
			frame.ip++
			vm.push(vm.stack[frame.slots+slot_idx])

		case OP_SET_LOCAL:
			// get value at stack top and store it in stack at position = operand
			val := vm.peek(0)
			slot_idx := int(vm.getCode()[frame.ip])
			frame.ip++
			if vm.stack[frame.slots+slot_idx].Immutable() {
				vm.runTimeError("Cannot assign to const local.\n")
				break Loop
			}
			vm.stack[frame.slots+slot_idx] = val

		case OP_JUMP_IF_FALSE:
			// if stack top is falsey, jump by offset ( 2 operands )
			offset := vm.readShort()
			if vm.isFalsey(vm.peek(0)) {
				frame.ip += int(offset)
			}

		case OP_JUMP:
			// jump by offset ( 2 operands )
			offset := vm.readShort()
			frame.ip += int(offset)

		case OP_LOOP:
			// jump back by offset ( 2 operands )
			offset := vm.readShort()
			frame.ip -= int(offset)

		case OP_CALL:
			// arg count is operand, function object is on stack after arguments, result will be stack top
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			if !vm.callValue(vm.peek(int(argCount)), int(argCount)) {
				return INTERPRET_RUNTIME_ERROR, makeNilValue()
			}
			frame = vm.frame()

		default:
			vm.runTimeError("Invalid Opcode")
			break Loop
		}
	}
	return INTERPRET_RUNTIME_ERROR, makeNilValue()
}

// numbers and strings only
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
		vm.push(makeNumberValue(nv1.Get()+nv2.Get(), false))
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

	vm.push(makeNumberValue(nv1.Get()-nv2.Get(), false))
	return true
}

func (vm *VM) binaryMultiply() bool {
	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.(type) {
	case NumberValue:
		switch v1.(type) {
		case NumberValue:
			vm.push(makeNumberValue(v1.(NumberValue).Get()*v2.(NumberValue).Get(), false))
		case ObjectValue:
			if !v1.(ObjectValue).isStringObject() {
				vm.runTimeError("Invalid operand for multiply.")
				return false
			}
			vm.push(vm.stringMultiply(v1.String(), int(v2.(NumberValue).Get())))
		default:
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}
	case ObjectValue:
		if !v2.(ObjectValue).isStringObject() {
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}
		switch v1.(type) {
		case NumberValue:
			vm.push(vm.stringMultiply(v2.String(), int(v1.(NumberValue).Get())))
		default:
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}

	default:
		vm.runTimeError("Invalid operand for multiply.")
		return false
	}

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

	vm.push(makeNumberValue(nv1.Get()/nv2.Get(), false))
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
	vm.push(makeNumberValue(-f, false))
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

	vm.push(makeBooleanValue(nv1.Get() > nv2.Get(), false))
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

	vm.push(makeBooleanValue(nv1.Get() < nv2.Get(), false))
	return true
}

func (vm *VM) concatenate(s1, s2 string) {

	so := MakeStringObject(s1 + s2)
	vm.push(makeObjectValue(so, false))
}

func (vm *VM) stringMultiply(s string, x int) Value {
	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return makeObjectValue(MakeStringObject(rv), false)
}
