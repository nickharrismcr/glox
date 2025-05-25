package lox

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"
)

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

const (
	FRAMES_MAX          int     = 64
	STACK_MAX           int     = FRAMES_MAX * 256
	GC_COLLECT_INTERVAL float64 = 5
)

type CallFrame struct {
	closure  *ClosureObject
	ip       int
	slots    int // start of vm stack for this frame
	handlers *ExceptionHandler
}

type VM struct {
	script       string
	stack        [STACK_MAX]Value
	stackTop     int
	globals      map[string]Value
	frames       [FRAMES_MAX]*CallFrame
	frameCount   int
	starttime    time.Time
	lastGC       time.Time
	openUpValues *UpvalueObject // head of list
	args         []string
}

type ExceptionHandler struct {
	exceptIP  uint16
	stackTop  int
	className string
	prev      *ExceptionHandler
}

//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------

var globalModules = map[string]bool{}

func NewVM(script string) *VM {

	vm := &VM{
		script:       script,
		globals:      map[string]Value{},
		starttime:    time.Now(),
		lastGC:       time.Now(),
		openUpValues: nil,
		args:         []string{},
	}
	vm.resetStack()
	vm.defineBuiltIns()
	return vm
}

func (vm *VM) SetArgs(args []string) {
	vm.args = args
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

//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------

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

	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := vm.frames[i]
		function := frame.closure.function

		fmt.Fprintf(os.Stderr, "[line %d] in ", function.chunk.lines[frame.ip])
		if function.name.get() == "" {
			fmt.Fprintf(os.Stderr, "%s \n", vm.script)
		} else {
			fmt.Fprintf(os.Stderr, "%s \n", function.name.get())
		}
	}

	vm.resetStack()

}

func (vm *VM) defineBuiltIn(name string, function BuiltInFn) {

	vm.push(makeObjectValue(makeStringObject(name), false))
	vm.push(makeObjectValue(makeBuiltInObject(function), false))
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

	if callee.Type == VAL_OBJ {
		if callee.isClosureObject() {
			return vm.call(getClosureObjectValue(callee), argCount)

		} else if callee.isBuiltInFunction() {
			nf := callee.asBuiltIn()
			res := nf.function(argCount, vm.stackTop-argCount, vm)
			if res.Type == VAL_NIL { // error occurred
				return false
			}
			vm.stackTop -= argCount + 1
			vm.push(res)
			return true

		} else if callee.isClassObject() {
			class := callee.asClass()
			vm.stack[vm.stackTop-argCount-1] = makeObjectValue(makeInstanceObject(class), false)
			if initialiser, ok := class.methods["init"]; ok {
				return vm.call(initialiser.asClosure(), argCount)
			} else if argCount != 0 {
				vm.runTimeError("Expected 0 arguments but got %d", argCount)
				return false
			}
			return true

		} else if callee.isBoundMethodObject() {
			bound := callee.asBoundMethod()
			vm.stack[vm.stackTop-argCount-1] = bound.receiver
			return vm.call(bound.method, argCount)
		}
	}
	vm.runTimeError("Can only call functions and classes.")
	return false
}

// optimised method call/module access
func (vm *VM) invoke(name Value, argCount int) bool {
	receiver := vm.peek(argCount)
	if receiver.Type != VAL_OBJ {
		vm.runTimeError("Invalid use of '.' operator")
		return false
	}
	switch receiver.Obj.getType() {
	case OBJECT_INSTANCE:
		instance := receiver.asInstance()
		return vm.invokeFromClass(instance.class, name, argCount)
	case OBJECT_MODULE:
		module := receiver.asModule()
		return vm.invokeFromModule(module, name, argCount)
	default:
		vm.runTimeError("Invalid use of '.' operator")
		return false
	}
	return true
}

func (vm *VM) invokeFromClass(class *ClassObject, name Value, argCount int) bool {
	n := getStringValue(name)
	method, ok := class.methods[n]
	if !ok {
		vm.runTimeError("Undefined method '%s'.", n)
		return false
	}
	return vm.call(method.asClosure(), argCount)
}

func (vm *VM) invokeFromModule(module *ModuleObject, name Value, argCount int) bool {
	n := getStringValue(name)
	fn, ok := module.globals[n]
	if !ok {
		vm.runTimeError("Undefined module property '%s'.", n)
		return false
	}
	return vm.callValue(fn, argCount)
}

func (vm *VM) bindMethod(class *ClassObject, name string) bool {
	method, ok := class.methods[name]
	if !ok {
		vm.runTimeError("Undefined property '%s'", name)
		return false
	}
	bound := makeBoundMethodObject(vm.peek(0), method.asClosure())
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
	class := vm.peek(1).asClass()
	class.methods[name] = method
	vm.pop()
}

func (vm *VM) call(closure *ClosureObject, argCount int) bool {

	if argCount != closure.function.arity {
		vm.runTimeError("Expected %d arguments but got %d.", closure.function.arity, argCount)
		return false
	}
	frame := &CallFrame{
		closure:  closure,
		ip:       0,
		slots:    vm.stackTop - argCount - 1,
		handlers: nil,
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

	switch v.Type {
	case VAL_FLOAT:
		return v.Float == 0
	case VAL_NIL:
		return true
	case VAL_BOOL:
		return !v.Bool
	}
	return true
}

// main interpreter loop
func (vm *VM) run() (InterpretResult, Value) {

	counter := 0
	frame := vm.frame()
Loop:
	for {

		counter++
		if counter%100000 == 0 {
			elapsed := time.Since(vm.lastGC).Seconds()
			if elapsed > GC_COLLECT_INTERVAL {
				runtime.GC()
				vm.lastGC = time.Now()
			}
		}

		inst := vm.getCode()[frame.ip]
		if DebugTraceExecution {
			if DebugShowGlobals {
				vm.showGlobals()
			}
			vm.stackTrace()
			_ = frame.closure.function.chunk.disassembleInstruction(inst, frame.ip)
		}

		frame.ip++
		switch inst {

		case OP_INVOKE:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			method := frame.closure.function.chunk.constants[idx]
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			if !vm.invoke(method, int(argCount)) {
				break Loop
			}
			frame = vm.frames[vm.frameCount-1]

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
			if v.Type != VAL_OBJ {
				vm.runTimeError("Property not found.")
				break Loop
			}

			idx := vm.getCode()[frame.ip]
			frame.ip++
			nv := frame.closure.function.chunk.constants[idx]
			name := getStringValue(nv)

			switch v.Obj.getType() {
			case OBJECT_INSTANCE:
				ot := v.Obj.(*InstanceObject)
				if val, ok := ot.fields[name]; ok {
					vm.pop()
					vm.push(val)
				} else {
					if !vm.bindMethod(ot.class, name) {
						break Loop
					}
				}
			case OBJECT_MODULE:
				ot := v.Obj.(*ModuleObject)
				if val, ok := ot.globals[name]; ok {
					vm.pop()
					vm.push(val)
				} else {
					vm.runTimeError("Property %s not found.", name)
					break Loop
				}
			default:
				vm.runTimeError("Property not found.")
				break Loop
			}

		case OP_SET_PROPERTY:

			val := vm.peek(0)
			v := vm.peek(1)
			if v.Type != VAL_OBJ {
				vm.runTimeError("Property not found.")
				break Loop
			}
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			switch v.Obj.getType() {
			case OBJECT_INSTANCE:
				ot := v.Obj.(*InstanceObject)
				ot.fields[name] = val
				tmp := vm.pop()
				vm.pop()
				vm.push(tmp)
			case OBJECT_MODULE:
				ot := v.Obj.(*ModuleObject)
				ot.globals[name] = val
				tmp := vm.pop()
				vm.pop()
				vm.push(tmp)
			default:
				vm.runTimeError("Property not found.")
				break Loop
			}

		case OP_EQUAL:
			// pop 2 stack values, stack top = boolean
			a := vm.pop()
			b := vm.pop()
			vm.push(makeBooleanValue(valuesEqual(a, b, false), false))

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
			// compiler ensures stack top will be a string object via OP_STR
			v := vm.pop()
			fmt.Printf("%s\n", v.asString())

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

		// entered a try block, IP of the except block is encoded in the next 2 instructions
		// push an exception handler storing that info
		case OP_TRY:
			exceptIP := vm.readShort()
			frame.handlers = &ExceptionHandler{
				exceptIP: exceptIP,
				stackTop: vm.stackTop,
				prev:     frame.handlers,
			}

		// ended a try block OK, so pop the handler block
		case OP_END_TRY:
			frame.handlers = frame.handlers.prev

		// marks the start of an exception handler block.  index of exception classname is in next instruction
		case OP_EXCEPT:

		// marks the end of an exception handler block
		case OP_END_EXCEPT:

		// 1 pop the thrown exception instance from the stack
		// 2 get the top frame exception handler - this has the IP of the first handler OP_EXCEPT.
		//   next instruction is an index to the exception classname in constants.
		//
		case OP_RAISE:
			err := vm.pop()
		outer:
			for {
				handler := frame.handlers
				if handler != nil {
					vm.stackTop = handler.stackTop
					vm.push(err)
					// jump to handler
					frame.ip = int(handler.exceptIP)
				inner:
					for {
						// get handler classname
						frame.ip += 2
						idx := vm.getCode()[frame.ip-1]
						name := getStringValue(frame.closure.function.chunk.constants[idx])
						handler_class := vm.globals[name].asClass()
						err_class := getInstanceObjectValue(err).class
						// is error a subclass of handler
						if err_class.IsSubclassOf(handler_class) {
							// yes, continue in handler block
							frame.handlers = handler.prev
							break outer
						}
						if !vm.nextHandler() {
							break inner
						}
					}
				}
				if !vm.popFrame() {
					vm.runTimeError("Uncaught exception: %s", err.String())
					return INTERPRET_RUNTIME_ERROR, makeNilValue()
				}
				frame = vm.frame()
			}

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

		case OP_INHERIT:
			superclass := vm.peek(1)
			subclass := vm.peek(0).asClass()
			if superclass.Type == VAL_OBJ {
				if superclass.isClassObject() {
					sco := superclass.asClass()
					for k, v := range sco.methods {
						subclass.methods[k] = v
					}
					subclass.super = superclass.asClass()
					vm.pop()
					continue
				}
			}

			vm.runTimeError("Superclass must be a class.")
			return INTERPRET_RUNTIME_ERROR, makeNilValue()

		case OP_GET_SUPER:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := getStringValue(frame.closure.function.chunk.constants[idx])
			v := vm.pop()
			superclass := v.asClass()

			if !vm.bindMethod(superclass, name) {
				return INTERPRET_RUNTIME_ERROR, makeNilValue()
			}

		case OP_SUPER_INVOKE:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			method := frame.closure.function.chunk.constants[idx]
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			superclass := vm.pop().asClass()
			if !vm.invokeFromClass(superclass, method, int(argCount)) {
				return INTERPRET_RUNTIME_ERROR, makeNilValue()
			}
			frame = vm.frames[vm.frameCount-1]

		// NJH added:

		case OP_IMPORT:

			idx := vm.getCode()[frame.ip]
			frame.ip++
			mv := frame.closure.function.chunk.constants[idx]
			module := mv.asString()
			// if already imported do nothing
			if ok := globalModules[module]; ok {
				panic("Import cycle detected.")
			}
			status := vm.importModule(module)
			if status != INTERPRET_OK {
				return status, makeNilValue()
			}

		case OP_STR:

			// replace stack top with string repr of it
			v := vm.peek(0) // may be needed for class toString so don't pop now
			s := v.String()
			switch v.Type {
			case VAL_OBJ:
				ov := v.Obj
				switch ov.getType() {
				case OBJECT_STRING:
					ot := ov.(StringObject)
					s = ot.get()
				case OBJECT_INSTANCE:
					// get string repr of class by calling asString() method if present
					ot := ov.(*InstanceObject)
					if toString, ok := ot.class.methods["toString"]; ok {
						vm.call(toString.asClosure(), 0)
						frame = vm.frame()
						continue
					}
					s = v.String()
				}
			}
			vm.pop()
			vm.push(makeObjectValue(makeStringObject(s), false))

		case OP_CREATE_LIST:
			// item count is operand, expects items on stack,  list object will be stack top
			vm.createList(frame)

		case OP_CREATE_DICT:
			// kay/pair item count is operand, expects keys/values on stack,  dict object will be stack top
			vm.createDict(frame)

		case OP_INDEX:
			// list/map + index on stack,  item at index -> stack top
			if !vm.index(frame) {
				break Loop
			}

		case OP_INDEX_ASSIGN:
			// list + index + RHS on stack,  list updated in place
			if !vm.indexAssign() {
				break Loop
			}

		case OP_SLICE:
			// list + from/to on stack. nil indicates from start/end.  new list at index -> stack top
			if !vm.slice() {
				break Loop
			}
		case OP_SLICE_ASSIGN:
			// list + from/to + RHS on stack.  list updated in place
			if !vm.sliceAssign() {
				break Loop
			}

		default:
			vm.runTimeError("Invalid Opcode")
			break Loop
		}
	}
	return INTERPRET_RUNTIME_ERROR, makeNilValue()
}

func (vm *VM) nextHandler() bool {

	for {
		vm.frame().ip++
		if vm.getCode()[vm.frame().ip] == OP_END_EXCEPT {
			if vm.getCode()[vm.frame().ip+1] == OP_EXCEPT {
				vm.frame().ip++
				return true
			}
			break
		}
	}
	return false
}

func (vm *VM) popFrame() bool {
	if vm.frameCount == 1 {
		return false
	}
	vm.frameCount--
	vm.stackTop = vm.frames[vm.frameCount].slots
	return true
}

func (vm *VM) importModule(module string) InterpretResult {

	globalModules[module] = true
	searchPath := getPath(vm.args, module) + ".lox"
	bytes, err := ioutil.ReadFile(searchPath)
	if err != nil {
		fmt.Printf("Could not find module %s.", searchPath)
		os.Exit(1)
	}
	subvm := NewVM(searchPath)
	subvm.SetArgs(vm.args)
	res, _ := subvm.Interpret(string(bytes))
	if res != INTERPRET_OK {
		return res
	}

	mo := makeModuleObject(module, subvm.globals)
	v := makeObjectValue(mo, false)
	vm.globals[module] = v

	return INTERPRET_OK
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

func (vm *VM) createDict(frame *CallFrame) {

	itemCount := int(vm.getCode()[frame.ip])
	frame.ip++
	dict := map[string]Value{}

	for i := 0; i < itemCount; i++ {
		value := vm.pop()
		key := vm.pop()
		dict[key.String()] = value
	}
	do := makeDictObject(dict)
	vm.push(makeObjectValue(do, false))
}

func (vm *VM) index(frame *CallFrame) bool {

	v := vm.pop()
	sv := vm.pop()

	if sv.Type == VAL_OBJ {
		switch sv.Obj.getType() {
		case OBJECT_LIST:
			if v.Type != VAL_INT {
				vm.runTimeError("Subscript must be an integer.")
				return false
			}
			t := sv.Obj.(*ListObject)
			idx := v.Int
			lo, err := t.index(idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(lo)
			return true

		case OBJECT_STRING:
			if v.Type != VAL_INT {
				vm.runTimeError("Subscript must be an integer.")
				return false
			}
			idx := v.Int
			t := sv.Obj.(StringObject)
			so, err := t.index(idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(so)
			return true

		case OBJECT_DICT:

			key := v.String()
			t := sv.Obj.(*DictObject)
			so, err := t.get(key)
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

func (vm *VM) indexAssign() bool {

	// collection, index, RHS on stack
	rhs := vm.pop()
	index := vm.pop()
	collection := vm.peek(0)
	if collection.Type == VAL_OBJ {
		switch collection.Obj.getType() {
		case OBJECT_LIST:
			if index.Type == VAL_INT {
				t := collection.Obj.(*ListObject)
				if err := t.assignToIndex(index.Int, rhs); err != nil {
					vm.runTimeError(err.Error())
					return false
				} else {
					return true
				}
			} else {
				vm.runTimeError("List index must an integer.")
				return false
			}
		case OBJECT_DICT:
			t := collection.Obj.(*DictObject)
			t.set(index.String(), rhs)
			return true
		}
	}
	vm.runTimeError("Can only assign to collection.")
	return false
}

func (vm *VM) slice() bool {

	var from_idx, to_idx int

	v_to := vm.pop()
	if v_to.Type == VAL_NIL {
		to_idx = -1
	} else if v_to.Type != VAL_INT {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		to_idx = v_to.Int
	}

	v_from := vm.pop()
	if v_from.Type == VAL_NIL {
		from_idx = 0
	} else if v_from.Type != VAL_INT {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		from_idx = v_from.Int
	}

	lv := vm.pop()
	if lv.Type == VAL_OBJ {
		if lv.Obj.getType() == OBJECT_LIST {
			lo, err := lv.Obj.(*ListObject).slice(from_idx, to_idx)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			vm.push(lo)
			return true

		} else if lv.Obj.getType() == OBJECT_STRING {
			so, err := lv.Obj.(StringObject).slice(from_idx, to_idx)
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

func (vm *VM) sliceAssign() bool {

	var from_idx, to_idx int

	val := vm.pop() // RHS

	v_to := vm.pop()
	if v_to.Type == VAL_NIL {
		to_idx = -1
	} else if v_to.Type != VAL_INT {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		to_idx = v_to.Int
	}

	v_from := vm.pop()
	if v_from.Type == VAL_NIL {
		from_idx = 0
	} else if v_from.Type != VAL_INT {
		vm.runTimeError("Invalid type in slice expression.")
		return false
	} else {
		from_idx = v_from.Int
	}

	lv := vm.peek(0)
	if lv.Type == VAL_OBJ {

		if lv.Obj.getType() == OBJECT_LIST {

			err := lv.Obj.(*ListObject).assignToSlice(from_idx, to_idx, val)
			if err != nil {
				vm.runTimeError(err.Error())
				return false
			}
			return true
		}
	}
	vm.runTimeError("Can only assign to list slice.")
	return false
}

// numbers and strings only
func (vm *VM) binaryAdd() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case VAL_INT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeIntValue(v1.Int+v2.Int, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float+float64(v2.Int), false))
			return true
		}
		vm.runTimeError("Addition type mismatch")
		return false

	case VAL_FLOAT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeFloatValue(float64(v1.Int)+v2.Float, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float+v2.Float, false))
			return true
		}
		vm.runTimeError("Addition type mismatch")
		return false

	case VAL_OBJ:
		ov2 := v2.Obj
		switch ov2.getType() {
		case OBJECT_STRING:
			if v1.Type != VAL_OBJ {
				vm.runTimeError("Addition type mismatch")
				return false
			}
			ov1 := v1.Obj
			if ov1.getType() == OBJECT_STRING {
				so := makeStringObject(v1.asString() + v2.asString())
				vm.push(makeObjectValue(so, false))
				return true
			}

		case OBJECT_LIST:

			if v1.Type != VAL_OBJ {
				vm.runTimeError("Addition type mismatch")
				return false
			}
			ov1 := v1.Obj
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
	v1 := vm.pop()

	switch v2.Type {
	case VAL_INT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeIntValue(v1.Int-v2.Int, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float-float64(v2.Int), false))
			return true
		}

	case VAL_FLOAT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeFloatValue(float64(v1.Int)-v2.Float, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float-v2.Float, false))
			return true
		}
	}

	vm.runTimeError("Addition type mismatch")
	return false
}

func (vm *VM) binaryMultiply() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case VAL_INT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeIntValue(v1.Int*v2.Int, false))
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float*float64(v2.Int), false))
		case VAL_OBJ:
			if !v1.isStringObject() {
				vm.runTimeError("Invalid operand for multiply.")
				return false
			}
			s := v1.Obj.(StringObject).get()
			vm.push(vm.stringMultiply(s, v2.Int))
		default:
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}
	case VAL_FLOAT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeFloatValue(float64(v1.Int)*v2.Float, false))
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float*v2.Float, false))
		default:
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}
	case VAL_OBJ:
		if !v2.isStringObject() {
			vm.runTimeError("Invalid operand for multiply.")
			return false
		}
		switch v1.Type {
		case VAL_INT:
			s := v2.asString()
			vm.push(vm.stringMultiply(s, v1.Int))
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
	v1 := vm.pop()

	switch v2.Type {
	case VAL_INT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeIntValue(v1.Int/v2.Int, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float/float64(v2.Int), false))
			return true
		}

	case VAL_FLOAT:
		switch v1.Type {
		case VAL_INT:
			vm.push(makeFloatValue(float64(v1.Int)/v2.Float, false))
			return true
		case VAL_FLOAT:
			vm.push(makeFloatValue(v1.Float/v2.Float, false))
			return true
		}
	}

	vm.runTimeError("Addition type mismatch")
	return false
}

func (vm *VM) binaryModulus() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.isInt() || !v2.isInt() {
		vm.runTimeError("Operands must be integers")
		return false
	}
	vm.push(makeIntValue(v1.Int%v2.Int, false))

	return true
}

func (vm *VM) unaryNegate() bool {

	v := vm.pop()
	switch v.Type {
	case VAL_FLOAT:
		f := v.Float
		vm.push(makeFloatValue(-f, false))
		return true
	case VAL_INT:
		f := v.Int
		vm.push(makeIntValue(-f, false))
		return true
	}

	vm.runTimeError("Operand must be a number")
	return false

}

func (vm *VM) binaryGreater() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.isNumber() || !v2.isNumber() {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(makeBooleanValue(v1.asFloat() > v2.asFloat(), false))
	return true
}

func (vm *VM) binaryLess() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.isNumber() || !v2.isNumber() {
		vm.runTimeError("Operands must be numbers")
		return false
	}

	vm.push(makeBooleanValue(v1.asFloat() < v2.asFloat(), false))
	return true
}

func (vm *VM) stringMultiply(s string, x int) Value {

	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return makeObjectValue(makeStringObject(rv), false)
}

func getPath(args []string, module string) string {
	if len(args) == 0 {
		return module
	}
	path := args[0]
	if strings.Contains(path, "/") {
		list := strings.Split(path, "/")
		searchPath := list[0 : len(list)-1]
		return strings.Join(searchPath, "/") + "/" + module
	}
	if strings.Contains(path, "\\") {
		list := strings.Split(path, "\\")
		searchPath := list[0 : len(list)-1]
		return strings.Join(searchPath, "\\") + "\\" + module
	}
	return module
}
