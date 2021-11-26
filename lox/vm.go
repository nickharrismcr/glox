package lox

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
	FRAMES_MAX       int = 64
	STACK_MAX        int = FRAMES_MAX * 256
	GC_COLLECT_POINT int = 10000
)

type CallFrame struct {
	closure *ClosureObject
	ip      int
	slots   int // start of vm stack for this frame
}

type VM struct {
	chunk        *Chunk
	ip           int
	stack        [STACK_MAX]Value
	stackTop     int
	globals      map[string]Value
	frames       [FRAMES_MAX]*CallFrame
	frameCount   int
	starttime    time.Time
	counter      int
	openUpValues *UpvalueObject // head of list
}

func NewVM() *VM {

	vm := &VM{
		globals:      map[string]Value{},
		starttime:    time.Now(),
		openUpValues: nil,
	}
	vm.resetStack()
	vm.defineNatives()

	return vm
}

func (vm *VM) Interpret(source string) (InterpretResult, string) {

	function := vm.compile(source)
	if function == nil {
		return INTERPRET_COMPILE_ERROR, ""
	}
	closure := makeClosureObject(function)
	vm.push(makeObjectValue(closure, false))
	vm.call(closure, 0)
	res, val := vm.run()
	return res, val.String()
}

func (vm *VM) frame() *CallFrame {

	return vm.frames[vm.frameCount-1]
}

func (vm *VM) getCode() []uint8 {

	return vm.frame().closure.function.chunk.code
}

func (vm *VM) resetStack() {

	vm.stackTop = 0
	vm.frameCount = 0
}

func (vm *VM) runTimeError(format string, args ...interface{}) {

	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprint(os.Stderr, "\n")
	//line := vm.frame().function.chunk.lines[vm.frame().ip-1]
	//fmt.Fprintf(os.Stderr, "[line %d] in script \n", line)

	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := vm.frames[i]
		function := frame.closure.function

		fmt.Fprintf(os.Stderr, "[line %d] in ", function.chunk.lines[frame.ip])
		if function.name.get() == "" {
			fmt.Fprintf(os.Stderr, "script \n")
		} else {
			fmt.Fprintf(os.Stderr, "%s \n", function.name.get())
		}
	}

	vm.resetStack()

}

func (vm *VM) defineNative(name string, function NativeFn) {

	vm.push(makeObjectValue(makeStringObject(name), false))
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

		if ov.isClosureObject() {
			return vm.call(getClosureObjectValue(callee), argCount)

		} else if ov.isNativeFunction() {
			nf := ov.asNative()
			res := nf.function(argCount, vm.stackTop-argCount, vm)
			if _, ok := res.(NilValue); ok { // error occurred
				return false
			}
			vm.stackTop -= argCount + 1
			vm.push(res)
			return true

		} else if ov.isClassObject() {
			class := ov.asClass()
			vm.stack[vm.stackTop-argCount-1] = makeObjectValue(makeInstanceObject(class), false)
			if initialiser, ok := class.methods["init"]; ok {
				return vm.call(initialiser.(ObjectValue).asClosure(), argCount)
			} else if argCount != 0 {
				vm.runTimeError("Expected 0 arguments but got %d", argCount)
				return false
			}
			return true

		} else if ov.isBoundMethodObject() {
			bound := ov.asBoundMethod()
			vm.stack[vm.stackTop-argCount-1] = bound.receiver
			return vm.call(bound.method, argCount)
		}
	}
	vm.runTimeError("Can only call functions and classes.")
	return false
}

func (vm *VM) bindMethod(class *ClassObject, name string) bool {
	method, ok := class.methods[name]
	if !ok {
		vm.runTimeError("Undefined property '%s'", name)
		return false
	}
	bound := makeBoundMethodObject(vm.peek(0), method.(ObjectValue).asClosure())
	vm.pop()
	vm.push(makeObjectValue(bound, false))
	return true
}

func (vm *VM) captureUpvalue(slot int) *UpvalueObject {

	var prevUpvalue *UpvalueObject = nil

	upvalue := vm.openUpValues
	for upvalue != nil && upvalue.slot > slot {
		prevUpvalue = upvalue
		upvalue = upvalue.next
	}
	if upvalue != nil && upvalue.slot == slot {
		return upvalue
	}
	new := makeUpvalueObject(&(vm.stack[slot]), slot)
	new.next = upvalue
	if prevUpvalue == nil {
		vm.openUpValues = new
	} else {
		prevUpvalue.next = new
	}
	return new
}

func (vm *VM) closeUpvalues(last int) {
	for vm.openUpValues != nil && vm.openUpValues.slot >= last {
		upvalue := vm.openUpValues
		upvalue.closed = vm.stack[upvalue.slot]
		upvalue.location = &upvalue.closed
		vm.openUpValues = upvalue.next
	}
}

func (vm *VM) defineMethod(name string) {
	method := vm.peek(0)
	class := vm.peek(1).(ObjectValue).asClass()
	class.methods[name] = method
	vm.pop()
}

func (vm *VM) call(closure *ClosureObject, argCount int) bool {

	if argCount != closure.function.arity {
		vm.runTimeError("Expected %d arguments but got %d.", closure.function.arity, argCount)
		return false
	}
	frame := &CallFrame{
		closure: closure,
		ip:      0,
		slots:   vm.stackTop - argCount - 1,
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
		return v.(NumberValue).get() == 0
	case NilValue:
		return true
	case BooleanValue:
		return !v.(BooleanValue).get()
	}
	return true
}

// main interpreter loop
func (vm *VM) run() (InterpretResult, Value) {

	frame := vm.frame()
Loop:
	for {
		vm.counter++
		if vm.counter == GC_COLLECT_POINT {
			runtime.GC()
			vm.counter = 0
		}
		inst := vm.getCode()[frame.ip]

		if DebugTraceExecution {
			vm.stackTrace()
			_ = frame.closure.function.chunk.disassembleInstruction(inst, frame.ip)
		}

		frame.ip++
		switch inst {

		case OP_CLOSURE:
			// get the function indexed by operand from constants, wrap in a closure object and push onto stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			function := frame.closure.function.chunk.constants[idx]
			closure := makeClosureObject(getFunctionObjectValue(function))
			vm.push(makeObjectValue(closure, false))
			for i := 0; i < closure.upvalueCount; i++ {
				isLocal := vm.getCode()[frame.ip]
				frame.ip++
				index := int(vm.getCode()[frame.ip])
				frame.ip++
				if isLocal == 1 {
					closure.upvalues[i] = vm.captureUpvalue(frame.slots + index)
				} else {
					upv := frame.closure.upvalues[index]
					closure.upvalues[i] = upv
				}
			}

		case OP_RETURN:
			// exit, return the value at stack top
			result := vm.pop()
			vm.closeUpvalues(frame.slots)
			vm.frameCount--
			if vm.frameCount == 0 {
				vm.pop() // drop main script function obj
				runtime.GC()
				return INTERPRET_OK, result
			}
			vm.stackTop = frame.slots
			vm.push(result)
			frame = vm.frames[vm.frameCount-1]

		case OP_GET_UPVALUE:
			slot := vm.getCode()[frame.ip]
			frame.ip++
			valIdx := frame.closure.upvalues[slot].location
			vm.push(*valIdx)

		case OP_SET_UPVALUE:
			slot := vm.getCode()[frame.ip]
			frame.ip++
			*(frame.closure.upvalues[slot].location) = vm.peek(0)

		case OP_CLOSE_UPVALUE:
			vm.closeUpvalues(vm.stackTop - 1)
			vm.pop()

		case OP_CONSTANT:
			// get the constant indexed by operand and push it onto the stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			constant := frame.closure.function.chunk.constants[idx]
			vm.push(constant)

		case OP_METHOD:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := frame.closure.function.chunk.constants[idx]
			vm.defineMethod(getStringValue(name))

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

		case OP_MODULUS:
			// pop 2 stack values, take modulus and push onto the stack
			if !vm.binaryModulus() {
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

		case OP_GET_PROPERTY:
			v := vm.peek(0)
			ov, ok := v.(ObjectValue)
			if !ok || !ov.isInstanceObject() {
				vm.runTimeError("Only instances have properties.")
				break Loop
			}

			instance := getInstanceObjectValue(v)
			idx := vm.getCode()[frame.ip]
			frame.ip++
			nv := frame.closure.function.chunk.constants[idx]
			name := getStringValue(nv)
			if v, ok := instance.fields[name]; ok {
				vm.pop()
				vm.push(v)
			} else {
				if !vm.bindMethod(instance.class, name) {
					break Loop
				}

			}

		case OP_SET_PROPERTY:

			val := vm.peek(0)
			v := vm.peek(1)
			ov, ok := v.(ObjectValue)
			if !ok || !ov.isInstanceObject() {
				vm.runTimeError("Only instances have fields.")
				break Loop
			}
			instance := getInstanceObjectValue(v)
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			instance.fields[name] = val
			tmp := vm.pop()
			vm.pop()
			vm.push(tmp)

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
			s := ""
			switch v.(type) {
			case NumberValue:
				s = v.String()
			case BooleanValue:
				s = v.String()
			case NilValue:
				s = v.String()
			case ObjectValue:
				a := v.(ObjectValue).get()
				switch a.(type) {
				case StringObject:
					s = v.(ObjectValue).asString()
				case *FunctionObject:
					s = v.String()
				case *ListObject:
					s = v.String()
				case *ClassObject:
					s = v.String()
				case *InstanceObject:
					s = v.String()
				}

			}
			fmt.Printf("%s\n", s)

		case OP_POP:
			// pop 1 stack value and discard
			_ = vm.pop()

		case OP_DEFINE_GLOBAL:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			vm.globals[name] = vm.peek(0)
			vm.pop()

		case OP_DEFINE_GLOBAL_CONST:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it and flag as immutable
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			vm.globals[name] = vm.peek(0)
			vm.globals[name] = immutable(vm.globals[name])
			vm.pop()

		case OP_GET_GLOBAL:
			// name = constant at operand index
			// push globals[name] onto stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
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
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			if _, ok := vm.globals[name]; !ok {
				vm.runTimeError("Undefined variable %s\n", name)
				break Loop
			}
			if vm.globals[name].Immutable() {
				vm.runTimeError("Cannot assign to const %s\n", name)
				break Loop
			}
			vm.globals[name] = mutable(vm.peek(0)) // in case of assignment of const

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
			vm.stack[frame.slots+slot_idx] = mutable(val)

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
			// arg count is operand, callable object is on stack after arguments, result will be stack top
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			if !vm.callValue(vm.peek(int(argCount)), int(argCount)) {
				return INTERPRET_RUNTIME_ERROR, makeNilValue()
			}
			frame = vm.frame()

		case OP_CLASS:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			vm.push(makeObjectValue(makeClassObject(name), false))

		// NJH added:

		case OP_CREATE_LIST:
			// item count is operand, expects items on stack,  list object will be stack top

			vm.createList(frame)

		case OP_INDEX:
			// list + index on stack,  item at index -> stack top
			if !vm.index(frame) {
				break Loop
			}

		case OP_SLICE:
			// list + from/to on stack. nil indicates from start/end.  new list at index -> stack top
			if !vm.slice(frame) {
				break Loop
			}

		default:
			vm.runTimeError("Invalid Opcode")
			break Loop
		}
	}
	return INTERPRET_RUNTIME_ERROR, makeNilValue()
}

func (vm *VM) createList(frame *CallFrame) {

	itemCount := int(vm.getCode()[frame.ip])
	frame.ip++
	list := []Value{}

	for i := 0; i < itemCount; i++ {
		list = append([]Value{vm.pop()}, list...) // reverse order
	}
	lo := makeListObject(list)
	vm.push(makeObjectValue(lo, false))
}

func (vm *VM) index(frame *CallFrame) bool {

	var nv NumberValue
	var ov ObjectValue
	var ok bool

	v := vm.pop()
	if nv, ok = v.(NumberValue); !ok {
		vm.runTimeError("Subscript must be a number.")
		return false
	}
	idx := int(nv.get())
	lv := vm.pop()
	if ov, ok = lv.(ObjectValue); ok {
		t := ov.get().getType()
		if t == OBJECT_LIST {
			lo, err := ov.get().(*ListObject).index(idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(lo)
			return true

		} else if t == OBJECT_STRING {
			so, err := ov.get().(StringObject).index(idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(so)
			return true
		}
	}
	vm.runTimeError("Invalid type for subscript.")
	return false
}

func (vm *VM) slice(frame *CallFrame) bool {

	var nv_from NumberValue
	var nv_to NumberValue
	var ov ObjectValue
	var from_idx, to_idx int
	var ok bool

	v_to := vm.pop()
	if _, ok = v_to.(NilValue); ok {
		to_idx = -1
	} else if nv_to, ok = v_to.(NumberValue); !ok {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		to_idx = int(nv_to.get())
	}

	v_from := vm.pop()
	if _, ok = v_from.(NilValue); ok {
		from_idx = 0
	} else if nv_from, ok = v_from.(NumberValue); !ok {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		from_idx = int(nv_from.get())
	}

	lv := vm.pop()
	if ov, ok = lv.(ObjectValue); ok {

		if ov.get().getType() == OBJECT_LIST {

			lo, err := ov.get().(*ListObject).slice(from_idx, to_idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(lo)
			return true

		} else if ov.get().getType() == OBJECT_STRING {
			so, err := ov.get().(StringObject).slice(from_idx, to_idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(so)
			return true
		}
	}
	vm.runTimeError("Invalid type for slice.")
	return false
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
		vm.push(makeNumberValue(nv1.get()+nv2.get(), false))
		return true

	case ObjectValue:
		o2 := v2.(ObjectValue)
		ov2 := o2.value
		switch ov2.getType() {
		case OBJECT_STRING:
			v1 := vm.pop()
			o1, ok := v1.(ObjectValue)
			if !ok {
				vm.runTimeError("Addition type mismatch")
				return false
			}
			ov1 := o1.value
			if ov1.getType() == OBJECT_STRING {
				so := makeStringObject(o1.asString() + o2.asString())
				vm.push(makeObjectValue(so, false))
				return true
			}

		case OBJECT_LIST:
			v1 := vm.pop()
			o1, ok := v1.(ObjectValue)
			if !ok {
				vm.runTimeError("Addition type mismatch")
				return false
			}
			ov1 := o1.value
			if ov1.getType() == OBJECT_LIST {
				lo := ov1.(*ListObject).add(ov2.(*ListObject))
				vm.push(makeObjectValue(lo, false))
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

	vm.push(makeNumberValue(nv1.get()-nv2.get(), false))
	return true
}

func (vm *VM) binaryMultiply() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.(type) {
	case NumberValue:
		switch v1.(type) {
		case NumberValue:
			vm.push(makeNumberValue(v1.(NumberValue).get()*v2.(NumberValue).get(), false))
		case ObjectValue:
			if !v1.(ObjectValue).isStringObject() {
				vm.runTimeError("Invalid operand for multiply.")
				return false
			}
			s := v1.(ObjectValue).get().(StringObject).get()
			vm.push(vm.stringMultiply(s, int(v2.(NumberValue).get())))
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
			s := v2.(ObjectValue).get().(StringObject).get()
			vm.push(vm.stringMultiply(s, int(v1.(NumberValue).get())))
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

	vm.push(makeNumberValue(nv1.get()/nv2.get(), false))
	return true
}

func (vm *VM) binaryModulus() bool {

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
	iv1 := int(nv1.get())
	iv2 := int(nv2.get())
	ret := float64(iv1 % iv2)
	vm.push(makeNumberValue(ret, false))
	return true
}

func (vm *VM) unaryNegate() bool {

	v := vm.pop()
	nv, ok := v.(NumberValue)
	if !ok {
		vm.runTimeError("Operand must be a number")
		return false
	}
	f := nv.get()
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

	vm.push(makeBooleanValue(nv1.get() > nv2.get(), false))
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

	vm.push(makeBooleanValue(nv1.get() < nv2.get(), false))
	return true
}

func (vm *VM) concatenate(s1, s2 string) {

}

func (vm *VM) stringMultiply(s string, x int) Value {

	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return makeObjectValue(makeStringObject(rv), false)
}
