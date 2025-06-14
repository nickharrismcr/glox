package lox

import (
	"bytes"
	"fmt"
	"glox/src/core"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var ForceModuleCompile = false

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
	closure  *core.ClosureObject
	ip       int
	slots    int // start of vm stack for this frame
	handlers *ExceptionHandler
	depth    int
}

type VM struct {
	script       string
	source       string
	stack        [STACK_MAX]core.Value
	stackTop     int
	frames       [FRAMES_MAX]*CallFrame
	frameCount   int
	starttime    time.Time
	lastGC       time.Time
	openUpValues *core.UpvalueObject // head of list
	args         []string
	ErrorMsg     string
	stackTrace   []string
	ModuleImport bool
	builtIns     map[string]core.Value // global built-in functions
}

type ExceptionHandler struct {
	exceptIP uint16
	stackTop int
	prev     *ExceptionHandler
}

//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------

var globalModules = map[string]string{}

func NewVM(script string, defineBuiltIns bool) *VM {

	vm := &VM{
		script:       script,
		starttime:    time.Now(),
		lastGC:       time.Now(),
		openUpValues: nil,
		args:         []string{},
		ErrorMsg:     "",
		stackTrace:   []string{},
		builtIns:     make(map[string]core.Value),
	}
	vm.resetStack()
	if defineBuiltIns {
		vm.defineBuiltIns()
	}
	return vm
}

// func (vm *VM) popEnvironment() {
// 	if vm.Environments.prev != nil {
// 		vm.Environments = vm.Environments.prev
// 	}
// }

func (vm *VM) SetArgs(args []string) {
	vm.args = args
}

func (vm *VM) Interpret(source string, module string) (InterpretResult, string) {

	vm.source = source
	function := vm.compile(source, module)
	if function == nil {
		return INTERPRET_COMPILE_ERROR, ""
	}
	if vm.ModuleImport {
		b := new(bytes.Buffer)
		function.Chunk.Serialise(b)
		writeToLxc(vm, b)
	}
	closure := core.MakeClosureObject(function)
	vm.push(core.MakeObjectValue(closure, false))
	vm.call(closure, 0)
	res, val := vm.run()
	return res, val.String()
}

func (vm *VM) Stack(index int) core.Value {

	if index < 0 || index >= vm.stackTop {
		return core.MakeNilValue()
	}
	return vm.stack[index]
}

func (vm *VM) Args() []string {

	return vm.args
}

func (vm *VM) StartTime() time.Time {

	return vm.starttime
}

func (subvm *VM) callLoadedChunk(name string, newEnv *core.Environment, chunk *core.Chunk) {

	function := core.MakeFunctionObject(name, newEnv)
	function.Chunk = chunk
	function.Name = core.MakeStringObject(name)
	closure := core.MakeClosureObject(function)
	subvm.push(core.MakeObjectValue(closure, false))
	subvm.call(closure, 0)
	_, _ = subvm.run()
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

	return vm.frame().closure.Function.Chunk.Code
}

func (vm *VM) resetStack() {

	vm.stackTop = 0
	vm.frameCount = 0
}

func (vm *VM) RunTimeError(format string, args ...any) {

	vm.ErrorMsg = fmt.Sprintf(format, args...)
}

func (vm *VM) defineBuiltIn(name string, function core.BuiltInFn) {
	vm.builtIns[name] = core.MakeObjectValue(core.MakeBuiltInObject(function), false)
}

func (vm *VM) push(v core.Value) {

	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

func (vm *VM) pop() core.Value {

	if vm.stackTop == 0 {
		return core.MakeNilValue()
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) Peek(dist int) core.Value {

	return vm.stack[(vm.stackTop-1)-dist]
}

/*func (vm *VM) set(dist int, val Value) {

	vm.stack[(vm.stackTop-1)-dist] = val
}*/

func (vm *VM) callValue(callee core.Value, argCount int) bool {

	if callee.Type == core.VAL_OBJ {
		if callee.IsClosureObject() {
			return vm.call(core.GetClosureObjectValue(callee), argCount)

		} else if callee.IsBuiltInObject() {
			nf := callee.AsBuiltIn()
			res := nf.Function(argCount, vm.stackTop-argCount, vm)
			if res.Type == core.VAL_NIL { // error occurred
				return false
			}
			vm.stackTop -= argCount + 1
			vm.push(res)
			return true

		} else if callee.IsClassObject() {
			class := callee.AsClass()
			vm.stack[vm.stackTop-argCount-1] = core.MakeObjectValue(core.MakeInstanceObject(class), false)
			if initialiser, ok := class.Methods["init"]; ok {
				return vm.call(initialiser.AsClosure(), argCount)
			} else if argCount != 0 {
				vm.RunTimeError("Expected 0 arguments but got %d", argCount)
				return false
			}
			return true

		} else if callee.IsBoundMethodObject() {
			bound := callee.AsBoundMethod()
			vm.stack[vm.stackTop-argCount-1] = bound.Receiver
			return vm.call(bound.Method, argCount)
		}
	}
	vm.RunTimeError("Can only call functions and classes.")
	return false
}

// optimised method call/module access
func (vm *VM) invoke(name core.Value, argCount int) bool {
	receiver := vm.Peek(argCount)
	if receiver.Type != core.VAL_OBJ {
		vm.RunTimeError("Invalid use of '.' operator")
		return false
	}
	switch receiver.Obj.GetType() {
	case core.OBJECT_INSTANCE:
		instance := receiver.AsInstance()
		return vm.invokeFromClass(instance.Class, name, argCount)
	case core.OBJECT_MODULE:
		module := receiver.AsModule()
		return vm.invokeFromModule(module, name, argCount)
	case core.OBJECT_FLOAT_ARRAY, core.OBJECT_STRING, core.OBJECT_LIST, core.OBJECT_DICT, core.OBJECT_GRAPHICS:
		return vm.invokeFromBuiltin(receiver.Obj, name, argCount)
	default:
		vm.RunTimeError("Invalid use of '.' operator")
		return false
	}

}

func (vm *VM) invokeFromClass(class *core.ClassObject, name core.Value, argCount int) bool {
	n := core.GetStringValue(name)
	method, ok := class.Methods[n]
	if !ok {
		vm.RunTimeError("Undefined method '%s'.", n)
		return false
	}
	return vm.call(method.AsClosure(), argCount)
}

func (vm *VM) invokeFromModule(module *core.ModuleObject, name core.Value, argCount int) bool {
	n := core.GetStringValue(name)
	fn, ok := module.Environment.GetVar(n)
	if !ok {
		vm.RunTimeError("Undefined module property '%s'.", n)
		return false
	}
	return vm.callValue(fn, argCount)
}

func (vm *VM) invokeFromBuiltin(obj core.Object, name core.Value, argCount int) bool {

	n := core.GetStringValue(name)
	bobj, ok := obj.(core.HasMethods)
	if ok {
		method := bobj.GetMethod(name.AsString().Get())
		if method != nil {
			builtin := method.Function
			res := builtin(argCount, vm.stackTop-argCount, vm)
			vm.stackTop -= argCount + 1
			vm.push(res)
			return true
		}
	}
	vm.RunTimeError("Undefined builtin property '%s'.", n)
	return false

}

func (vm *VM) bindMethod(class *core.ClassObject, name string) bool {
	method, ok := class.Methods[name]
	if !ok {
		vm.RunTimeError("Undefined property '%s'", name)
		return false
	}
	bound := core.MakeBoundMethodObject(vm.Peek(0), method.AsClosure())
	vm.pop()
	vm.push(core.MakeObjectValue(bound, false))
	return true
}

func (vm *VM) captureUpvalue(slot int) *core.UpvalueObject {

	var prevUpvalue *core.UpvalueObject = nil

	upvalue := vm.openUpValues
	for upvalue != nil && upvalue.Slot > slot {
		prevUpvalue = upvalue
		upvalue = upvalue.Next
	}
	if upvalue != nil && upvalue.Slot == slot {
		return upvalue
	}
	new := core.MakeUpvalueObject(&(vm.stack[slot]), slot)
	new.Next = upvalue
	if prevUpvalue == nil {
		vm.openUpValues = new
	} else {
		prevUpvalue.Next = new
	}
	return new
}

func (vm *VM) closeUpvalues(last int) {
	for vm.openUpValues != nil && vm.openUpValues.Slot >= last {
		upvalue := vm.openUpValues
		upvalue.Closed = vm.stack[upvalue.Slot]
		upvalue.Location = &upvalue.Closed
		vm.openUpValues = upvalue.Next
	}
}

func (vm *VM) defineMethod(name string) {
	method := vm.Peek(0)
	class := vm.Peek(1).AsClass()
	class.Methods[name] = method
	vm.pop()
}

func (vm *VM) call(closure *core.ClosureObject, argCount int) bool {

	if argCount != closure.Function.Arity {
		vm.RunTimeError("Expected %d arguments but got %d.", closure.Function.Arity, argCount)
		return false
	}
	frame := &CallFrame{
		closure:  closure,
		ip:       0,
		slots:    vm.stackTop - argCount - 1,
		handlers: nil,
		depth:    vm.frameCount + 1,
	}
	vm.frames[vm.frameCount] = frame
	vm.frameCount++
	if vm.frameCount == FRAMES_MAX {
		vm.RunTimeError("Stack overflow.")
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

func (vm *VM) readByte() uint8 {

	vm.frame().ip += 1
	return vm.getCode()[vm.frame().ip-1]
}

func (vm *VM) isFalsey(v core.Value) bool {

	switch v.Type {
	case core.VAL_FLOAT:
		return v.Float == 0
	case core.VAL_NIL:
		return true
	case core.VAL_BOOL:
		return !v.Bool
	}
	return true
}

// main interpreter loop
func (vm *VM) run() (InterpretResult, core.Value) {

	counter := 0
	vm.ErrorMsg = ""

	for {
		frame := vm.frame()
		function := frame.closure.Function
		chunk := function.Chunk
		constants := chunk.Constants

		counter++
		if counter%100000 == 0 {
			elapsed := time.Since(vm.lastGC).Seconds()
			if elapsed > GC_COLLECT_INTERVAL {
				runtime.GC()
				vm.lastGC = time.Now()
			}
		}

		inst := vm.getCode()[frame.ip]
		if DebugTraceExecution && !DebugSuppress {
			if DebugShowGlobals {
				vm.showGlobals()
			}
			vm.showStack()
			_ = disassembleInstruction(chunk, vm.script, frame, inst, frame.ip)
		}

		frame.ip++
		switch inst {

		case core.OP_INVOKE:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			method := constants[idx]
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			if !vm.invoke(method, int(argCount)) {
				goto End
			}

		case core.OP_CLOSURE:
			// get the function indexed by operand from constants, wrap in a closure object and push onto stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			function := constants[idx]
			closure := core.MakeClosureObject(core.GetFunctionObjectValue(function))
			vm.push(core.MakeObjectValue(closure, false))
			for i := 0; i < closure.UpvalueCount; i++ {
				isLocal := vm.getCode()[frame.ip]
				frame.ip++
				index := int(vm.getCode()[frame.ip])
				frame.ip++
				if isLocal == 1 {
					closure.Upvalues[i] = vm.captureUpvalue(frame.slots + index)
				} else {
					upv := frame.closure.Upvalues[index]
					closure.Upvalues[i] = upv
				}
			}

		case core.OP_RETURN:
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

		case core.OP_GET_UPVALUE:
			slot := vm.getCode()[frame.ip]
			frame.ip++
			valIdx := frame.closure.Upvalues[slot].Location
			vm.push(*valIdx)

		case core.OP_SET_UPVALUE:
			slot := vm.getCode()[frame.ip]
			frame.ip++
			*(frame.closure.Upvalues[slot].Location) = vm.Peek(0)

		case core.OP_CLOSE_UPVALUE:
			vm.closeUpvalues(vm.stackTop - 1)
			vm.pop()

		case core.OP_CONSTANT:
			// get the constant indexed by operand and push it onto the stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			constant := constants[idx]
			vm.push(constant)

		case core.OP_METHOD:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := constants[idx]
			vm.defineMethod(core.GetStringValue(name))

		case core.OP_NEGATE:
			// negate the value at stack top
			if !vm.unaryNegate() {
				goto End
			}

		case core.OP_ADD:
			// pop 2 stack values, add them and push onto the stack
			if !vm.binaryAdd() {
				goto End
			}

		case core.OP_SUBTRACT:
			// pop 2 stack values, subtract and push onto the stack
			if !vm.binarySubtract() {
				goto End
			}

		case core.OP_MULTIPLY:
			// pop 2 stack values, multiply and push onto the stack
			if !vm.binaryMultiply() {
				goto End
			}

		case core.OP_MODULUS:
			// pop 2 stack values, take modulus and push onto the stack
			if !vm.binaryModulus() {
				goto End
			}

		case core.OP_DIVIDE:
			// pop 2 stack values, divide and push onto the stack
			if !vm.binaryDivide() {
				goto End
			}

		case core.OP_NIL:
			// push nil val onto the stack
			vm.push(core.MakeNilValue())

		case core.OP_TRUE:
			// push true bool val onto the stack
			vm.push(core.MakeBooleanValue(true, false))

		case core.OP_FALSE:
			// push false bool val onto the stack
			vm.push(core.MakeBooleanValue(false, false))

		case core.OP_NOT:
			// replace stack top with boolean not of itself
			v := vm.pop()
			bv := vm.isFalsey(v)
			vm.push(core.MakeBooleanValue(bv, false))

		case core.OP_GET_PROPERTY:

			v := vm.Peek(0)
			if v.Type != core.VAL_OBJ {
				vm.RunTimeError("Attempt to access property of non-object.")
				goto End
			}

			idx := vm.getCode()[frame.ip]
			frame.ip++
			nv := constants[idx]
			name := core.GetStringValue(nv)

			switch v.Obj.GetType() {
			case core.OBJECT_INSTANCE:
				ot := v.AsInstance()
				if val, ok := ot.Fields[name]; ok {
					vm.pop()
					vm.push(val)
				} else {
					if !vm.bindMethod(ot.Class, name) {
						goto End
					}
				}
			case core.OBJECT_MODULE:
				ot := v.AsModule()
				if val, ok := ot.Environment.GetVar(name); ok {
					vm.pop()
					vm.push(val)
				} else {
					vm.RunTimeError("Property '%s' not found.", name)
					goto End
				}
			default:
				vm.RunTimeError("Property '%s' not found.", name)
				goto End
			}

		case core.OP_SET_PROPERTY:

			val := vm.Peek(0)
			v := vm.Peek(1)
			if v.Type != core.VAL_OBJ {
				vm.RunTimeError("Property not found.")
				goto End
			}
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			switch v.Obj.GetType() {
			case core.OBJECT_INSTANCE:
				ot := v.AsInstance()
				ot.Fields[name] = val
				tmp := vm.pop()
				vm.pop()
				vm.push(tmp)
			case core.OBJECT_MODULE:
				ot := v.AsModule()
				ot.Environment.SetVar(name, val)
				tmp := vm.pop()
				vm.pop()
				vm.push(tmp)
			default:
				vm.RunTimeError("Property not found.")
				goto End
			}

		case core.OP_EQUAL:
			// pop 2 stack values, stack top = boolean
			a := vm.pop()
			b := vm.pop()
			vm.push(core.MakeBooleanValue(core.ValuesEqual(a, b, false), false))

		case core.OP_GREATER:
			// pop 2 stack values, stack top = boolean
			if !vm.binaryGreater() {
				goto End
			}

		case core.OP_LESS:
			// pop 2 stack values, stack top = boolean
			if !vm.binaryLess() {
				goto End
			}

		case core.OP_PRINT:
			// compiler ensures stack top will be a string object via core.OP_STR
			v := vm.pop()
			fmt.Printf("%s\n", v.AsString().Get())

		case core.OP_POP:
			// pop 1 stack value and discard
			_ = vm.pop()

		case core.OP_DEFINE_GLOBAL:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			value := vm.Peek(0)
			//DumpValue("Define global", value)
			function.Environment.SetVar(name, value)
			vm.pop()

		case core.OP_DEFINE_GLOBAL_CONST:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it and flag as immutable
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			function.Environment.SetVar(name, vm.Peek(0))
			v, _ := function.Environment.GetVar(name)
			function.Environment.SetVar(name, core.Immutable(v))
			vm.pop()

		case core.OP_GET_GLOBAL:
			// name = constant at operand index
			// push globals[name] onto stack
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			value, ok := function.Environment.GetVar(name)
			//DumpValue("Get global", value)
			if !ok {
				value, ok = vm.builtIns[name]
				if !ok {
					vm.RunTimeError("Undefined variable %s", name)
					goto End
				}
			}
			vm.push(value)

		case core.OP_SET_GLOBAL:
			// name = constant at operand index
			// set globals[name] to stack top, key must exist
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			// auto-declare
			// if _, ok := vm.Environments.Vars[name]; !ok {
			// 	vm.RunTimeError("Undefined variable %s\n", name)
			// 	goto End
			// }
			v, _ := function.Environment.GetVar(name)
			if v.Immutable() {
				vm.RunTimeError("Cannot assign to const %s", name)
				goto End
			}
			function.Environment.SetVar(name, core.Mutable(vm.Peek(0))) // in case of assignment of const

		case core.OP_GET_LOCAL:
			// get local from stack at position = operand and push on stack top
			slot_idx := int(vm.getCode()[frame.ip])
			frame.ip++
			vm.push(vm.stack[frame.slots+slot_idx])

		case core.OP_SET_LOCAL:
			// get value at stack top and store it in stack at position = operand
			val := vm.Peek(0)
			slot_idx := int(vm.getCode()[frame.ip])
			frame.ip++
			if vm.stack[frame.slots+slot_idx].Immutable() {
				vm.RunTimeError("Cannot assign to const local.")
				goto End
			}
			vm.stack[frame.slots+slot_idx] = core.Mutable(val)

		case core.OP_JUMP_IF_FALSE:
			// if stack top is falsey, jump by offset ( 2 operands )
			offset := vm.readShort()
			if vm.isFalsey(vm.Peek(0)) {
				frame.ip += int(offset)
			}

		case core.OP_JUMP:
			// jump by offset ( 2 operands )
			offset := vm.readShort()
			frame.ip += int(offset)

		case core.OP_LOOP:
			// jump back by offset ( 2 operands )
			offset := vm.readShort()
			frame.ip -= int(offset)

		// entered a try block, IP of the except block is encoded in the next 2 instructions
		// push an exception handler storing that info
		case core.OP_TRY:
			exceptIP := vm.readShort()
			frame.handlers = &ExceptionHandler{
				exceptIP: exceptIP,
				stackTop: vm.stackTop,
				prev:     frame.handlers,
			}

		// ended a try block OK, so pop the handler block
		case core.OP_END_TRY:
			frame.handlers = frame.handlers.prev

		// marks the start of an exception handler block.  index of exception classname is in next instruction
		case core.OP_EXCEPT:
			frame.ip++

		// marks the end of an exception handler block
		case core.OP_END_EXCEPT:

		// 1 pop the thrown exception instance from the stack
		// 2 get the top frame exception handler - this has the IP of the first handler core.OP_EXCEPT.
		//   next instruction is an index to the exception classname in constants.
		//   if the thrown exception name matches the handler, run the handler
		//   else skip to the next handler if it exists, or unwind the call stack and retry.
		//   we'll either hit a matching hander or exit the vm with an unhandled exception error.
		case core.OP_RAISE:
			err := vm.pop()
			if !vm.raiseException(err) {
				return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()
			}

		case core.OP_CALL:
			// arg count is operand, callable object is on stack after arguments, result will be stack top
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			if !vm.callValue(vm.Peek(int(argCount)), int(argCount)) {
				goto End
			}

		case core.OP_CLASS:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			vm.push(core.MakeObjectValue(core.MakeClassObject(name), false))

		case core.OP_INHERIT:
			superclass := vm.Peek(1)
			subclass := vm.Peek(0).AsClass()
			if superclass.Type == core.VAL_OBJ {
				if superclass.IsClassObject() {
					sco := superclass.AsClass()
					for k, v := range sco.Methods {
						subclass.Methods[k] = v
					}
					subclass.Super = superclass.AsClass()
					vm.pop()
					continue
				}
			}

			vm.RunTimeError("Superclass must be a class.")
			return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()

		case core.OP_GET_SUPER:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			name := core.GetStringValue(constants[idx])
			v := vm.pop()
			superclass := v.AsClass()

			if !vm.bindMethod(superclass, name) {
				return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()
			}

		case core.OP_SUPER_INVOKE:
			idx := vm.getCode()[frame.ip]
			frame.ip++
			method := constants[idx]
			argCount := vm.getCode()[frame.ip]
			frame.ip++
			superclass := vm.pop().AsClass()
			if !vm.invokeFromClass(superclass, method, int(argCount)) {
				return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()
			}

		// NJH added:

		case core.OP_IMPORT:

			idx := vm.getCode()[frame.ip]
			frame.ip++
			mv := constants[idx]
			module := mv.AsString().Get()

			status := vm.importModule(module)
			if status != INTERPRET_OK {
				return status, core.MakeNilValue()
			}

		case core.OP_STR:

			// replace stack top with string repr of it
			v := vm.Peek(0) // may be needed for class toString so don't pop now
			s := v.String()
			switch v.Type {
			case core.VAL_OBJ:
				ov := v.Obj
				switch ov.GetType() {
				case core.OBJECT_STRING:
					ot := ov.(core.StringObject)
					s = ot.Get()
				case core.OBJECT_INSTANCE:
					// get string repr of class by calling AsString().Get() method if present
					ot := ov.(*core.InstanceObject)
					if toString, ok := ot.Class.Methods["toString"]; ok {
						vm.call(toString.AsClosure(), 0)
						continue
					}
					s = v.String()
				}
			}
			vm.pop()
			vm.push(core.MakeObjectValue(core.MakeStringObject(s), false))

		case core.OP_CREATE_LIST:
			// item count is operand, expects items on stack,  list object will be stack top
			vm.createList(frame)

		case core.OP_CREATE_TUPLE:
			// item count is operand, expects items on stack,  list object will be stack top
			vm.createTuple(frame)

		case core.OP_CREATE_DICT:
			// key/pair item count is operand, expects keys/values on stack,  dict object will be stack top
			vm.createDict(frame)

		case core.OP_INDEX:
			// list/map + index on stack,  item at index -> stack top
			if !vm.index() {
				goto End
			}

		case core.OP_INDEX_ASSIGN:
			// list + index + RHS on stack,  list updated in place
			if !vm.indexAssign() {
				goto End
			}

		case core.OP_SLICE:
			// list + from/to on stack. nil indicates from start/end.  new list at index -> stack top
			if !vm.slice() {
				goto End
			}
		case core.OP_SLICE_ASSIGN:
			// list + from/to + RHS on stack.  list updated in place
			if !vm.sliceAssign() {
				goto End
			}

		// local slot, end of foreach in next 3 instructions
		case core.OP_FOREACH:
			slot := vm.readByte()
			iterableSlot := vm.readByte()
			idxSlot := vm.readByte()
			jumpToEnd := vm.readShort()
			iterable := vm.stack[frame.slots+int(iterableSlot)]

			if iterable.Type != core.VAL_OBJ {
				vm.RunTimeError("Iterable in foreach must be list or string.")
				goto End
			}
			idxVal := vm.stack[frame.slots+int(idxSlot)]
			idx := idxVal.Int
			switch iterable.Obj.GetType() {
			case core.OBJECT_LIST:
				t := iterable.AsList()
				if idx >= len(t.Items) {
					frame.ip += int(jumpToEnd - 2)
				} else {
					val := t.Get()[idx]
					vm.stack[frame.slots+int(slot)] = val
				}

			case core.OBJECT_STRING:
				t := iterable.AsString()
				if idx >= len(t.Get()) {
					frame.ip += int(jumpToEnd - 2)

				} else {
					val, _ := t.Index(idx)
					vm.stack[frame.slots+int(slot)] = val
				}

			default:
				vm.RunTimeError("Iterable in foreach must be list or string.")
				goto End
			}

		case core.OP_NEXT:

			jumpToStart := vm.readShort()
			indexSlot := vm.readByte()
			slot := frame.slots + int(indexSlot)
			indexVal := vm.stack[slot]
			vm.stack[slot] = core.MakeIntValue(indexVal.Int+1, false)
			frame.ip -= int(jumpToStart + 1)

		case core.OP_END_FOREACH:

		// stack 1 : string or list
		// stack 2 : key or substring

		case core.OP_IN:

			b := vm.pop()
			a := vm.pop()

			if !(b.IsStringObject() || b.IsListObject()) {
				vm.RunTimeError("'in' requires string or list as right operand.")
				goto End
			}
			switch b.Obj.GetType() {
			case core.OBJECT_STRING:
				if !a.IsStringObject() {
					vm.RunTimeError("'in' requires string as left operand.")
					goto End
				}
				rv := b.AsString().Contains(a)
				vm.push(rv)
			case core.OBJECT_LIST:
				rv := b.AsList().Contains(a)
				vm.push(rv)
			}
		case core.OP_BREAKPOINT:
			// hit a breakpoint, pause execution
			vm.pauseExecution()

		default:
			vm.RunTimeError("Invalid Opcode")
			goto End
		}
	End:

		if vm.ErrorMsg != "" {
			if !vm.RaiseExceptionByName("RunTimeError", vm.ErrorMsg) {
				return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()
			}
		}
	}
	//return INTERPRET_RUNTIME_ERROR, core.MakeNilValue()
}

// natively raise an exception given a name:
// - get the class object for the name from globals
// - make an instance of the class and set the message on it
// - pass the instance to raiseException
// used for vm raising errors that can be handled in lox e.g EOFError when reading a file
func (vm *VM) RaiseExceptionByName(name string, msg string) bool {

	classVal := vm.builtIns[name]
	classObj := classVal.Obj
	instance := core.MakeInstanceObject(classObj.(*core.ClassObject))
	instance.Fields["msg"] = core.MakeObjectValue(core.MakeStringObject(msg), false)
	return vm.raiseException(core.MakeObjectValue(instance, false))
}

func (vm *VM) raiseException(err core.Value) bool {

	for {
		vm.appendStackTrace()
		handler := vm.frame().handlers
		if handler != nil {

			vm.stackTop = handler.stackTop
			vm.push(err)
			// jump to handler IP
			vm.frame().ip = int(handler.exceptIP)
		inner:
			for {
				// get handler classname
				vm.frame().ip += 2
				idx := vm.getCode()[vm.frame().ip-1]
				function := vm.frame().closure.Function
				name := core.GetStringValue(function.Chunk.Constants[idx])
				v, ok := function.Environment.GetVar(name)
				if !ok {
					v, ok = vm.builtIns[name]
					if !ok {
						vm.RunTimeError("Undefined exception handler '%s'.", name)
						return false
					}
				}
				handler_class := v.AsClass()
				err_class := core.GetInstanceObjectValue(err).Class
				// is error a subclass of handler
				if err_class.IsSubclassOf(handler_class) {
					// yes, continue in handler block
					vm.ErrorMsg = ""
					vm.stackTrace = []string{}
					vm.frame().handlers = handler.prev

					return true
				}
				// skip to start of next handler if exists
				if !vm.nextHandler() {
					break inner
				}
			}
		}
		// no more handlers in this call frame. if top level, exit
		// else unwind call stack and repeat

		if !vm.popFrame() {
			exc := err.AsInstance()
			vm.RunTimeError("Uncaught exception: %s : %s ", exc.Class, exc.Fields["msg"])
			return false
		}
	}
}

func (vm *VM) nextHandler() bool {

	for {
		vm.frame().ip++
		if vm.getCode()[vm.frame().ip] == core.OP_END_EXCEPT {
			if vm.getCode()[vm.frame().ip+1] == core.OP_EXCEPT {
				vm.frame().ip += 1
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

func (vm *VM) appendStackTrace() {

	frame := vm.frame()
	function := frame.closure.Function
	where, script := "", ""
	if function.Name.Get() == "" {
		script = vm.script
		where = "<module>"
	} else {
		script = function.Chunk.Filename
		where = function.Name.Get()
	}
	line := function.Chunk.Lines[frame.ip]

	s := fmt.Sprintf("File '%s' , line %d, in %s ", script, line, where)
	vm.stackTrace = append(vm.stackTrace, s)
	codeline := vm.sourceLine(script, line)
	vm.stackTrace = append(vm.stackTrace, codeline)
}

func (vm *VM) PrintStackTrace() {
	for _, v := range vm.stackTrace {
		fmt.Fprintf(os.Stderr, "%s\n", v)
	}
}

func (vm *VM) sourceLine(script string, line int) string {

	source := vm.source
	if script != vm.script {
		module := getModule(script)
		source = globalModules[module]
	}
	lines := strings.Split(source, "\n")
	if line > 0 && line <= len(lines) {
		rv := lines[line-1]
		return rv
	}
	return ""
}

func (vm *VM) importModule(moduleName string) InterpretResult {

	searchPath := getPath(vm.Args(), moduleName) + ".lox"
	bytes, err := os.ReadFile(searchPath)
	if err != nil {
		fmt.Printf("Could not find module %s.", searchPath)
		os.Exit(1)
	}
	globalModules[moduleName] = string(bytes)
	subvm := NewVM(searchPath, false)
	subvm.builtIns = vm.builtIns
	subvm.SetArgs(vm.Args())
	subvm.ModuleImport = true
	// see if we can load lxc bytecode file for the module.
	// if so run it
	if loadedChunk, newEnv, ok := loadLxc(searchPath); ok {
		loadedChunk.Filename = moduleName
		subvm.callLoadedChunk(moduleName, newEnv, loadedChunk)
	} else {
		// if not, load the module source, compile it and run it
		res, _ := subvm.Interpret(string(bytes), moduleName)
		if res != INTERPRET_OK {
			return res
		}
	}
	subfn := subvm.frames[0].closure.Function
	Debugf("subvm environment name = %s", subfn.Environment.Name)
	subvm_globals := subfn.Environment.Vars
	for k, v := range subvm_globals {
		Debugf("Import Found module property '%s' in subvm main func environment", k)
		if v.IsClosureObject() {
			Debugf("Found module property '%s' is a closure", k)
			if v.AsClosure().Function.Environment == nil {
				Debugf("Module property '%s' has no environment", k)
			}
			Debugf("Property %s environment vars count = %d", k, len(v.AsClosure().Function.Environment.Vars))
		}
	}

	mo := core.MakeModuleObject(moduleName, *subfn.Environment)
	v := core.MakeObjectValue(mo, false)
	vm.frame().closure.Function.Environment.SetVar(moduleName, v)
	return INTERPRET_OK
}

func (vm *VM) createList(frame *CallFrame) {

	itemCount := int(vm.getCode()[frame.ip])
	frame.ip++
	list := []core.Value{}

	for i := 0; i < itemCount; i++ {
		list = append([]core.Value{vm.pop()}, list...) // reverse order
	}
	lo := core.MakeListObject(list, false)
	vm.push(core.MakeObjectValue(lo, false))
}

func (vm *VM) createTuple(frame *CallFrame) {

	itemCount := int(vm.getCode()[frame.ip])
	frame.ip++
	list := []core.Value{}

	for i := 0; i < itemCount; i++ {
		list = append([]core.Value{vm.pop()}, list...) // reverse order
	}
	lo := core.MakeListObject(list, true)
	vm.push(core.MakeObjectValue(lo, true))
}

func (vm *VM) createDict(frame *CallFrame) {

	itemCount := int(vm.getCode()[frame.ip])
	frame.ip++
	dict := map[string]core.Value{}

	for i := 0; i < itemCount; i++ {
		value := vm.pop()
		key := vm.pop()
		dict[key.AsString().Get()] = value
	}
	do := core.MakeDictObject(dict)
	vm.push(core.MakeObjectValue(do, false))
}

func (vm *VM) index() bool {

	iv := vm.pop()
	sv := vm.pop()

	if sv.IsObj() {
		switch sv.Obj.GetType() {
		case core.OBJECT_LIST:
			if iv.Type != core.VAL_INT {
				vm.RunTimeError("Subscript must be an integer.")
				return false
			}
			t := sv.AsList()
			idx := iv.Int
			lo, err := t.Index(idx)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.push(lo)
			return true

		case core.OBJECT_STRING:
			if iv.Type != core.VAL_INT {
				vm.RunTimeError("Subscript must be an integer.")
				return false
			}
			idx := iv.Int
			t := sv.AsString()
			so, err := t.Index(idx)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.push(so)
			return true

		case core.OBJECT_DICT:

			key := iv.AsString().Get()
			t := sv.AsDict()
			so, err := t.Get(key)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.push(so)
			return true
		}

	}
	vm.RunTimeError("Invalid type for subscript.")
	return false
}

func (vm *VM) indexAssign() bool {

	// collection, index, RHS on stack
	rhs := vm.pop()
	index := vm.pop()
	collection := vm.Peek(0)
	if collection.Type == core.VAL_OBJ {
		switch collection.Obj.GetType() {
		case core.OBJECT_LIST:
			t := collection.AsList()
			if t.Tuple {
				vm.RunTimeError("Tuples are immutable.")
				return false
			}
			if index.Type == core.VAL_INT {
				if err := t.AssignToIndex(index.Int, rhs); err != nil {
					vm.RunTimeError("%v", err)
					return false
				} else {
					return true
				}
			} else {
				vm.RunTimeError("List index must an integer.")
				return false
			}
		case core.OBJECT_DICT:
			t := collection.AsDict()
			t.Set(index.AsString().Get(), rhs)
			return true
		}
	}
	vm.RunTimeError("Can only assign to collection.")
	return false
}

func (vm *VM) slice() bool {

	var from_idx, to_idx int

	v_to := vm.pop()
	if v_to.Type == core.VAL_NIL {
		to_idx = -1
	} else if v_to.Type != core.VAL_INT {
		vm.RunTimeError("Invalid type in slice expression.")
		return false
	} else {
		to_idx = v_to.Int
	}

	v_from := vm.pop()
	if v_from.Type == core.VAL_NIL {
		from_idx = 0
	} else if v_from.Type != core.VAL_INT {
		vm.RunTimeError("Invalid type in slice expression.")
		return false
	} else {
		from_idx = v_from.Int
	}

	lv := vm.pop()
	if lv.IsObj() {
		if lv.Obj.GetType() == core.OBJECT_LIST {
			lo, err := lv.AsList().Slice(from_idx, to_idx)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.push(lo)
			return true

		} else if lv.Obj.GetType() == core.OBJECT_STRING {
			so, err := lv.AsString().Slice(from_idx, to_idx)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.push(so)
			return true
		}
	}
	vm.RunTimeError("Invalid type for slice.")
	return false
}

func (vm *VM) sliceAssign() bool {

	var from_idx, to_idx int

	val := vm.pop() // RHS

	v_to := vm.pop()
	if v_to.Type == core.VAL_NIL {
		to_idx = -1
	} else if v_to.Type != core.VAL_INT {
		vm.RunTimeError("Invalid type in slice expression.")
		return false
	} else {
		to_idx = v_to.Int
	}

	v_from := vm.pop()
	if v_from.Type == core.VAL_NIL {
		from_idx = 0
	} else if v_from.Type != core.VAL_INT {
		vm.RunTimeError("Invalid type in slice expression.")
		return false
	} else {
		from_idx = v_from.Int
	}

	lv := vm.Peek(0)
	if lv.IsObj() {

		if lv.Obj.GetType() == core.OBJECT_LIST {
			lst := lv.AsList()
			if lst.Tuple {
				vm.RunTimeError("Tuples are immutable")
				return false
			}
			err := lst.AssignToSlice(from_idx, to_idx, val)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			return true
		}
	}
	vm.RunTimeError("Can only assign to list slice.")
	return false
}

// numbers and strings only
func (vm *VM) binaryAdd() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case core.VAL_INT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeIntValue(v1.Int+v2.Int, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float+float64(v2.Int), false))
			return true
		}
		vm.RunTimeError("Addition type mismatch")
		return false

	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeFloatValue(float64(v1.Int)+v2.Float, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float+v2.Float, false))
			return true
		}
		vm.RunTimeError("Addition type mismatch")
		return false

	case core.VAL_OBJ:
		ov2 := v2.Obj
		switch ov2.GetType() {
		case core.OBJECT_STRING:
			if v1.Type != core.VAL_OBJ {
				vm.RunTimeError("Addition type mismatch")
				return false
			}
			ov1 := v1.Obj
			if ov1.GetType() == core.OBJECT_STRING {
				so := core.MakeStringObject(v1.AsString().Get() + v2.AsString().Get())
				vm.push(core.MakeObjectValue(so, false))
				return true
			}

		case core.OBJECT_LIST:

			if v1.Type != core.VAL_OBJ {
				vm.RunTimeError("Addition type mismatch")
				return false
			}
			ov1 := v1.Obj
			if ov1.GetType() == core.OBJECT_LIST {
				lo := ov1.(*core.ListObject).Add(ov2.(*core.ListObject))
				vm.push(core.MakeObjectValue(lo, false))
				return true
			}
		}
	}
	vm.RunTimeError("Operands must be numbers or strings")
	return false
}

func (vm *VM) binarySubtract() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case core.VAL_INT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeIntValue(v1.Int-v2.Int, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float-float64(v2.Int), false))
			return true
		}

	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeFloatValue(float64(v1.Int)-v2.Float, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float-v2.Float, false))
			return true
		}
	}

	vm.RunTimeError("Addition type mismatch")
	return false
}

func (vm *VM) binaryMultiply() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case core.VAL_INT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeIntValue(v1.Int*v2.Int, false))
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float*float64(v2.Int), false))
		case core.VAL_OBJ:
			if !v1.IsStringObject() {
				vm.RunTimeError("Invalid operand for multiply.")
				return false
			}
			s := v1.AsString().Get()
			vm.push(vm.stringMultiply(s, v2.Int))
		default:
			vm.RunTimeError("Invalid operand for multiply.")
			return false
		}
	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeFloatValue(float64(v1.Int)*v2.Float, false))
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float*v2.Float, false))
		default:
			vm.RunTimeError("Invalid operand for multiply.")
			return false
		}
	case core.VAL_OBJ:
		if !v2.IsStringObject() {
			vm.RunTimeError("Invalid operand for multiply.")
			return false
		}
		switch v1.Type {
		case core.VAL_INT:
			s := v2.AsString().Get()
			vm.push(vm.stringMultiply(s, v1.Int))
		default:
			vm.RunTimeError("Invalid operand for multiply.")
			return false
		}

	default:
		vm.RunTimeError("Invalid operand for multiply.")
		return false
	}

	return true
}

func (vm *VM) binaryDivide() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case core.VAL_INT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeIntValue(v1.Int/v2.Int, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float/float64(v2.Int), false))
			return true
		}

	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.push(core.MakeFloatValue(float64(v1.Int)/v2.Float, false))
			return true
		case core.VAL_FLOAT:
			vm.push(core.MakeFloatValue(v1.Float/v2.Float, false))
			return true
		}
	}

	vm.RunTimeError("Addition type mismatch")
	return false
}

func (vm *VM) binaryModulus() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.IsInt() || !v2.IsInt() {
		vm.RunTimeError("Operands must be integers")
		return false
	}
	vm.push(core.MakeIntValue(v1.Int%v2.Int, false))

	return true
}

func (vm *VM) unaryNegate() bool {

	v := vm.pop()
	switch v.Type {
	case core.VAL_FLOAT:
		f := v.Float
		vm.push(core.MakeFloatValue(-f, false))
		return true
	case core.VAL_INT:
		f := v.Int
		vm.push(core.MakeIntValue(-f, false))
		return true
	}

	vm.RunTimeError("Operand must be a number")
	return false

}

func (vm *VM) binaryGreater() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.IsNumber() || !v2.IsNumber() {
		vm.RunTimeError("Operands must be numbers")
		return false
	}

	vm.push(core.MakeBooleanValue(v1.AsFloat() > v2.AsFloat(), false))
	return true
}

func (vm *VM) binaryLess() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	if !v1.IsNumber() || !v2.IsNumber() {
		vm.RunTimeError("Operands must be numbers")
		return false
	}

	vm.push(core.MakeBooleanValue(v1.AsFloat() < v2.AsFloat(), false))
	return true
}

func (vm *VM) stringMultiply(s string, x int) core.Value {

	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return core.MakeObjectValue(core.MakeStringObject(rv), false)
}
func (vm *VM) pauseExecution() {

	fmt.Println("⚠️  BREAKPOINT HIT")
	fmt.Println("Stack:", vm.stack[:vm.stackTop])
	// If you track them
	debug.PrintStack()   // optional, prints Go stack trace
	runtime.Breakpoint() // halt if debugger attached
	// or: pause until user input
	//buf := bufio.NewReader(os.Stdin)
	//fmt.Print("Press Enter to continue...")
	//buf.ReadBytes('\n')
}

// return the path to the given module.
// first, will look in lox/modules in the lox installation directory defined in LOX_PATH environment var.
// if not found will look in the directory containing the main module being run
func getPath(args []string, module string) string {

	lox_inst_dir := os.Getenv("LOX_PATH")

	if lox_inst_dir != "" {
		lox_inst_module_dir := lox_inst_dir + "/src/modules"
		path := lox_inst_module_dir + "/" + module
		_, err := os.Stat(path + ".lox")
		if err == nil {
			return path
		}
	}

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

func getModule(fullPath string) string {
	base := filepath.Base(fullPath)      // "foo.lox"
	ext := filepath.Ext(base)            // ".lox"
	return strings.TrimSuffix(base, ext) // "foo"
}
