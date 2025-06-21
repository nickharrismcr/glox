package lox

import (
	"bytes"
	"fmt"
	"glox/src/compiler"
	"glox/src/core"
	"glox/src/debug"
	"os"
	"path/filepath"
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

type VM struct {
	script       string
	source       string
	stack        [STACK_MAX]core.Value
	stackTop     int
	frames       [FRAMES_MAX]*core.CallFrame
	frameCount   int
	currCode     []uint8 // current code being executed
	starttime    time.Time
	lastGC       time.Time
	openUpValues *core.UpvalueObject // head of list
	args         []string
	ErrorMsg     string
	stackTrace   []string
	ModuleImport bool
	builtIns     map[int]core.Value   // global built-in functions
	foreachState *core.VMForeachState // state stack for foreach loops

}

var ITER_METHOD = core.MakeStringObjectValue("__iter__", true)
var NEXT_METHOD = core.MakeStringObjectValue("__next__", true)

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
		builtIns:     make(map[int]core.Value),
		foreachState: nil,
	}
	vm.resetStack()
	if defineBuiltIns && !core.DebugCompileOnly {
		vm.defineBuiltIns()
	}
	return vm
}

func NewVMForeachState(prev *core.VMForeachState, localSlot int, iterSlot int, jumpToStart int, jumpToEnd int) *core.VMForeachState {

	return &core.VMForeachState{
		LocalSlot:   localSlot,
		IterSlot:    iterSlot,
		JumpToStart: jumpToStart,
		JumpToEnd:   jumpToEnd,
		Stage:       core.WAITING_FOR_ITER,
		Prev:        prev,
	}
}

func (vm *VM) SetArgs(args []string) {
	vm.args = args
}

func (vm *VM) Interpret(source string, module string) (InterpretResult, string) {

	vm.source = source
	function := compiler.Compile(vm.script, source, module)
	if function == nil {
		return INTERPRET_COMPILE_ERROR, ""
	}
	if core.DebugCompileOnly {
		return INTERPRET_OK, ""
	}

	if vm.ModuleImport {
		b := new(bytes.Buffer)
		function.Chunk.Serialise(b)
		writeToLxc(vm, b)
	}
	closure := core.MakeClosureObject(function)
	vm.stack[vm.stackTop] = core.MakeObjectValue(closure, false)
	vm.stackTop++
	vm.call(closure, 0)
	res, val := vm.run()
	return res, val.String()
}

func (vm *VM) Stack(index int) core.Value {

	if index < 0 || index >= vm.stackTop {
		return core.NIL_VALUE
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

func (vm *VM) frame() *core.CallFrame {

	return vm.frames[vm.frameCount-1]
}

func (vm *VM) getCode() []uint8 {

	return vm.frame().Closure.Function.Chunk.Code
}

func (vm *VM) resetStack() {

	vm.stackTop = 0
	vm.frameCount = 0
}

func (vm *VM) RunTimeError(format string, args ...any) {

	vm.ErrorMsg = fmt.Sprintf(format, args...)
}

func (vm *VM) defineBuiltIn(name string, function core.BuiltInFn) {
	id := core.InternName(name)
	vm.builtIns[id] = core.MakeObjectValue(core.MakeBuiltInObject(function), false)
}

func (vm *VM) push(v core.Value) {

	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

func (vm *VM) pop() core.Value {

	if vm.stackTop == 0 {
		return core.NIL_VALUE
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
			vm.stack[vm.stackTop] = res
			vm.stackTop++
			return true

		} else if callee.IsClassObject() {
			class := callee.AsClass()
			vm.stack[vm.stackTop-argCount-1] = core.MakeObjectValue(core.MakeInstanceObject(class), false)
			if initialiser, ok := class.Methods[core.INIT]; ok {
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
	if receiver.Obj.IsBuiltIn() {
		return vm.invokeFromBuiltin(receiver.Obj, name, argCount)
	}

	switch receiver.Obj.GetType() {
	case core.OBJECT_INSTANCE:
		instance := receiver.AsInstance()
		return vm.invokeFromClass(instance.Class, name, argCount, false)
	case core.OBJECT_CLASS:
		class := receiver.AsClass()
		return vm.invokeFromClass(class, name, argCount, true)
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

func (vm *VM) invokeFromClass(class *core.ClassObject, name core.Value, argCount int, isStatic bool) bool {
	i := name.InternedId
	if isStatic {
		method, ok := class.StaticMethods[i]
		if !ok {
			vm.RunTimeError("Undefined static method '%s'.", core.GetStringValue(name))
			return false
		}
		return vm.call(method.AsClosure(), argCount)
	}
	method, ok := class.Methods[i]
	if !ok {
		vm.RunTimeError("Undefined method '%s'.", core.GetStringValue(name))
		return false
	}
	return vm.call(method.AsClosure(), argCount)
}

func (vm *VM) invokeFromModule(module *core.ModuleObject, name core.Value, argCount int) bool {

	fn, ok := module.Environment.GetVar(name.InternedId)
	if !ok {
		n := core.GetStringValue(name)
		vm.RunTimeError("Undefined module property '%s'.", n)
		return false
	}
	return vm.callValue(fn, argCount)
}

func (vm *VM) invokeFromBuiltin(obj core.Object, name core.Value, argCount int) bool {

	n := core.GetStringValue(name)
	bobj, ok := obj.(core.HasMethods)
	if ok {
		method := bobj.GetMethod(name.InternedId)
		if method != nil {
			builtin := method.Function
			res := builtin(argCount, vm.stackTop-argCount, vm)
			vm.stackTop -= argCount + 1
			vm.stack[vm.stackTop] = res
			vm.stackTop++
			return true
		}
	}
	vm.RunTimeError("Undefined builtin property '%s'.", n)
	return false

}

func (vm *VM) bindMethod(class *core.ClassObject, stringId int) bool {
	method, ok := class.Methods[stringId]
	if !ok {
		vm.RunTimeError("Undefined property '%s'", core.NameFromID(stringId))
		return false
	}
	bound := core.MakeBoundMethodObject(vm.Peek(0), method.AsClosure())
	vm.pop()
	vm.stack[vm.stackTop] = core.MakeObjectValue(bound, false)
	vm.stackTop++
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

func (vm *VM) defineMethod(stringID int, isStatic bool) {
	method := vm.Peek(0)
	class := vm.Peek(1).AsClass()
	if isStatic {
		class.StaticMethods[stringID] = method
	} else {
		class.Methods[stringID] = method
	}
	vm.pop()
}

func (vm *VM) call(closure *core.ClosureObject, argCount int) bool {

	if argCount != closure.Function.Arity {
		vm.RunTimeError("Expected %d arguments but got %d.", closure.Function.Arity, argCount)
		return false
	}
	frame := &core.CallFrame{
		Closure:  closure,
		Ip:       0,
		Slots:    vm.stackTop - argCount - 1,
		Handlers: nil,
		Depth:    vm.frameCount + 1,
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

	vm.frame().Ip += 2
	b1 := uint16(vm.currCode[vm.frame().Ip-2])
	b2 := uint16(vm.currCode[vm.frame().Ip-1])
	return uint16(b1<<8 | b2)
}

func (vm *VM) readByte() uint8 {

	vm.frame().Ip += 1
	return vm.currCode[vm.frame().Ip-1]
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
		function := frame.Closure.Function
		chunk := function.Chunk
		constants := chunk.Constants
		vm.currCode = chunk.Code

		counter++
		if counter%100000 == 0 {
			elapsed := time.Since(vm.lastGC).Seconds()
			if elapsed > GC_COLLECT_INTERVAL {
				runtime.GC()
				vm.lastGC = time.Now()
			}
		}

		inst := vm.currCode[frame.Ip]
		if core.DebugTraceExecution && !core.DebugSuppress {
			if core.DebugShowGlobals {
				vm.showGlobals()
			}
			vm.showStack()
			_ = debug.DisassembleInstruction(chunk, vm.script, function.Name.Get(), frame.Depth, inst, frame.Ip)
		}

		frame.Ip++
		switch inst {

		case core.OP_EQUAL:
			// pop 2 stack values, stack top = boolean
			a := vm.pop()
			b := vm.pop()
			vm.stack[vm.stackTop] = core.MakeBooleanValue(core.ValuesEqual(a, b, false), false)
			vm.stackTop++

		case core.OP_GREATER:
			// pop 2 stack values, stack top = boolean

			v2 := vm.pop()
			v1 := vm.pop()
			if !v1.IsNumber() || !v2.IsNumber() {
				vm.RunTimeError("Operands must be numbers")
				goto End
			}
			vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsFloat() > v2.AsFloat(), false)
			vm.stackTop++

		case core.OP_LESS:
			// pop 2 stack values, stack top = boolean

			v2 := vm.pop()
			v1 := vm.pop()
			if !v1.IsNumber() || !v2.IsNumber() {
				vm.RunTimeError("Operands must be numbers")
				goto End
			}
			vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsFloat() < v2.AsFloat(), false)
			vm.stackTop++

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
			idx := vm.currCode[frame.Ip]
			frame.Ip++

			value := vm.Peek(0)
			//DumpValue("Define global", value)
			function.Environment.SetVar(constants[idx].InternedId, value)
			vm.pop()

		case core.OP_DEFINE_GLOBAL_CONST:
			// name = constant at operand index
			// pop 1 stack value and set globals[name] to it and flag as immutable
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			id := constants[idx].InternedId
			function.Environment.SetVar(id, vm.Peek(0))
			v, _ := function.Environment.GetVar(id)
			function.Environment.SetVar(id, core.Immutable(v))
			vm.pop()

		case core.OP_GET_GLOBAL:
			// name = constant at operand index
			// push globals[name] onto stack
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			id := constants[idx].InternedId

			value, ok := function.Environment.GetVar(id)
			//DumpValue("Get global", value)
			if !ok {
				value, ok = vm.builtIns[id]
				if !ok {
					name := core.GetStringValue(constants[idx])
					vm.RunTimeError("Undefined variable %s", name)
					goto End
				}
			}
			vm.stack[vm.stackTop] = value
			vm.stackTop++

		case core.OP_SET_GLOBAL:
			// name = constant at operand index
			// set globals[name] to stack top, key must exist
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			id := constants[idx].InternedId

			v, _ := function.Environment.GetVar(id)
			if v.Immutable() {
				name := core.GetStringValue(constants[idx])
				vm.RunTimeError("Cannot assign to const %s", name)
				goto End
			}
			function.Environment.SetVar(id, core.Mutable(vm.Peek(0))) // in case of assignment of const

		case core.OP_GET_LOCAL:
			// get local from stack at position = operand and push on stack top
			slot_idx := int(vm.currCode[frame.Ip])
			frame.Ip++
			vm.stack[vm.stackTop] = vm.stack[frame.Slots+slot_idx]
			vm.stackTop++

		case core.OP_SET_LOCAL:
			// get value at stack top and store it in stack at position = operand
			val := vm.Peek(0)
			slot_idx := int(vm.currCode[frame.Ip])
			frame.Ip++
			if vm.stack[frame.Slots+slot_idx].Immutable() {
				vm.RunTimeError("Cannot assign to const local.")
				goto End
			}
			vm.stack[frame.Slots+slot_idx] = core.Mutable(val)

		case core.OP_JUMP_IF_FALSE:
			// if stack top is falsey, jump by offset ( 2 operands )
			offset := vm.readShort()
			if vm.isFalsey(vm.Peek(0)) {
				frame.Ip += int(offset)
			}

		case core.OP_JUMP:
			// jump by offset ( 2 operands )
			offset := vm.readShort()
			frame.Ip += int(offset)

		case core.OP_GET_UPVALUE:
			slot := vm.currCode[frame.Ip]
			frame.Ip++
			valIdx := frame.Closure.Upvalues[slot].Location
			vm.stack[vm.stackTop] = *valIdx
			vm.stackTop++

		case core.OP_SET_UPVALUE:
			slot := vm.currCode[frame.Ip]
			frame.Ip++
			*(frame.Closure.Upvalues[slot].Location) = vm.Peek(0)

		case core.OP_CLOSE_UPVALUE:
			vm.closeUpvalues(vm.stackTop - 1)
			vm.pop()

		case core.OP_CONSTANT:
			// get the constant indexed by operand and push it onto the stack
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			constant := constants[idx]
			vm.stack[vm.stackTop] = constant
			vm.stackTop++

		case core.OP_CALL:
			// arg count is operand, callable object is on stack after arguments, result will be stack top
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			if !vm.callValue(vm.Peek(int(argCount)), int(argCount)) {
				goto End
			}

		case core.OP_ADD:
			// pop 2 stack values, add them and push onto the stack
			v2 := vm.pop()
			v1 := vm.pop()
			switch v2.Type {
			case core.VAL_INT:
				switch v1.Type {
				case core.VAL_INT:
					vm.stack[vm.stackTop] = core.MakeIntValue(v1.Int+v2.Int, false)
					vm.stackTop++
					continue
				case core.VAL_FLOAT:
					vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float+float64(v2.Int), false)
					vm.stackTop++
					continue
				}
				vm.RunTimeError("Addition type mismatch")
				goto End

			case core.VAL_FLOAT:
				switch v1.Type {
				case core.VAL_INT:
					vm.stack[vm.stackTop] = core.MakeFloatValue(float64(v1.Int)+v2.Float, false)
					vm.stackTop++
					continue
				case core.VAL_FLOAT:
					vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float+v2.Float, false)
					vm.stackTop++
					continue
				}
				vm.RunTimeError("Addition type mismatch")
				goto End

			case core.VAL_OBJ:
				ov2 := v2.Obj
				switch ov2.GetType() {
				case core.OBJECT_STRING:
					if v1.Type != core.VAL_OBJ {
						vm.RunTimeError("Addition type mismatch")
						goto End
					}
					ov1 := v1.Obj
					if ov1.GetType() == core.OBJECT_STRING {
						vm.stack[vm.stackTop] = core.MakeStringObjectValue(v1.AsString().Get()+v2.AsString().Get(), false)
						vm.stackTop++

						continue
					}
				case core.OBJECT_LIST:
					if v1.Type != core.VAL_OBJ {
						vm.RunTimeError("Addition type mismatch")
						goto End
					}
					ov1 := v1.Obj
					if ov1.GetType() == core.OBJECT_LIST {
						lo := ov1.(*core.ListObject).Add(ov2.(*core.ListObject))
						vm.stack[vm.stackTop] = core.MakeObjectValue(lo, false)
						vm.stackTop++
						continue
					}
				}
			}
			vm.RunTimeError("Operands must be numbers or strings")
			goto End

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
			vm.stack[vm.stackTop] = core.NIL_VALUE
			vm.stackTop++

		case core.OP_TRUE:
			// push true bool val onto the stack
			vm.stack[vm.stackTop] = core.MakeBooleanValue(true, false)
			vm.stackTop++

		case core.OP_FALSE:
			// push false bool val onto the stack
			vm.stack[vm.stackTop] = core.MakeBooleanValue(false, false)
			vm.stackTop++

		case core.OP_NOT:
			// replace stack top with boolean not of itself
			v := vm.pop()
			bv := vm.isFalsey(v)
			vm.stack[vm.stackTop] = core.MakeBooleanValue(bv, false)
			vm.stackTop++

		case core.OP_LOOP:
			// jump back by offset ( 2 operands )
			offset := vm.readShort()
			frame.Ip -= int(offset)

		case core.OP_INVOKE:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			method := constants[idx]
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			if !vm.invoke(method, int(argCount)) {
				goto End
			}

		case core.OP_CLOSURE:
			// get the function indexed by operand from constants, wrap in a closure object and push onto stack
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			function := constants[idx]
			closure := core.MakeClosureObject(core.GetFunctionObjectValue(function))
			vm.stack[vm.stackTop] = core.MakeObjectValue(closure, false)
			vm.stackTop++
			for i := 0; i < closure.UpvalueCount; i++ {
				isLocal := vm.currCode[frame.Ip]
				frame.Ip++
				index := int(vm.currCode[frame.Ip])
				frame.Ip++
				if isLocal == 1 {
					closure.Upvalues[i] = vm.captureUpvalue(frame.Slots + index)
				} else {
					upv := frame.Closure.Upvalues[index]
					closure.Upvalues[i] = upv
				}
			}

		case core.OP_RETURN:
			// exit, return the value at stack top
			result := vm.pop()
			vm.closeUpvalues(frame.Slots)
			vm.frameCount--
			if vm.frameCount == 0 {
				vm.pop() // drop main script function obj
				runtime.GC()
				return INTERPRET_OK, result
			}
			vm.stackTop = frame.Slots
			vm.stack[vm.stackTop] = result
			vm.stackTop++

			// #### iterable class handling ####
			// returning from an __iter__ instance method call?
			if vm.foreachState != nil && vm.foreachState.Stage == core.WAITING_FOR_ITER {
				// result holds the iterator object, it had better have a __next__ method
				if !result.IsInstanceObject() {
					vm.RunTimeError("Foreach iterator must be a object with a __next__ method.")
					return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
				}
				_, ok := result.AsInstance().Class.Methods[core.NEXT]
				if !ok {
					vm.RunTimeError("Foreach iterator must have a __next__ method.")
					return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
				}
				vm.stack[vm.foreachState.IterSlot] = result // store iterator object in stack
				// call __next__ instance method to get the first item
				vm.foreachState.Stage = core.WAITING_FOR_NEXT
				vm.stack[vm.stackTop] = result
				vm.stackTop++ //TODO needed? is already on stack
				vm.invoke(NEXT_METHOD, 0)
				continue
			}
			// returning from a __next__ instance method call?
			if vm.foreachState != nil && vm.foreachState.Stage == core.WAITING_FOR_NEXT {
				// result holds the next item value
				vm.pop() //TODO needed? - dup push above ?
				frame = vm.frame()
				if result.Type == core.VAL_NIL {
					// we have no more items, so jump to end of foreach loop
					frame.Ip = int(vm.foreachState.JumpToEnd)
					vm.foreachState = vm.foreachState.Prev // pop the foreach state
				} else {
					// we have a value, so set it in the local slot and continue
					vm.stack[vm.foreachState.LocalSlot] = result
					// jump to start of foreach loop
					frame.Ip = int(vm.foreachState.JumpToStart)
				}
			}

		case core.OP_METHOD:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := constants[idx]
			vm.defineMethod(name.InternedId, false)

		case core.OP_STATIC_METHOD:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := constants[idx]
			vm.defineMethod(name.InternedId, true)

		case core.OP_NEGATE:
			// negate the value at stack top
			v := vm.pop()
			switch v.Type {
			case core.VAL_FLOAT:
				f := v.Float
				vm.stack[vm.stackTop] = core.MakeFloatValue(-f, false)
				vm.stackTop++
				continue
			case core.VAL_INT:
				f := v.Int
				vm.stack[vm.stackTop] = core.MakeIntValue(-f, false)
				vm.stackTop++
				continue
			}
			vm.RunTimeError("Operand must be a number")
			goto End

		case core.OP_GET_PROPERTY:

			v := vm.Peek(0)
			if v.Type != core.VAL_OBJ && v.Type != core.VAL_VEC2 && v.Type != core.VAL_VEC3 && v.Type != core.VAL_VEC4 {
				vm.RunTimeError("Attempt to access property of non-object.")
				goto End
			}

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			nv := constants[idx]
			stringId := nv.InternedId

			bobj, ok := v.Obj.(core.HasConstants)
			if ok {
				val := bobj.GetConstant(stringId)
				vm.pop() // pop the object
				vm.stack[vm.stackTop] = val
				vm.stackTop++
				continue
			}

			switch v.Obj.GetType() {
			case core.OBJECT_VEC2:
				// special case for Vec2, which has x and y properties
				switch stringId {
				case core.X:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec2().X, false)
					vm.stackTop++
					continue
				case core.Y:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec2().Y, false)
					vm.stackTop++
					continue
				}
			case core.OBJECT_VEC3:
				// special case for Vec3, which has x, y and z properties
				switch stringId {
				case core.X:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec3().X, false)
					vm.stackTop++
					continue
				case core.Y:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec3().Y, false)
					vm.stackTop++
					continue
				case core.Z:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec3().Z, false)
					vm.stackTop++
					continue
				}
			case core.OBJECT_VEC4:
				// special case for Vec4, which has x, y, z and w properties
				switch stringId {
				case core.X, core.R:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec4().X, false)
					vm.stackTop++
					continue
				case core.Y, core.G:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec4().Y, false)
					vm.stackTop++
					continue
				case core.Z, core.B:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec4().Z, false)
					vm.stackTop++
					continue
				case core.W, core.A:
					vm.pop()
					vm.stack[vm.stackTop] = core.MakeFloatValue(v.AsVec4().W, false)
					vm.stackTop++
					continue
				}

			case core.OBJECT_INSTANCE:
				ot := v.AsInstance()
				if val, ok := ot.Fields[stringId]; ok {
					vm.pop()
					vm.stack[vm.stackTop] = val
					vm.stackTop++
				} else {
					if !vm.bindMethod(ot.Class, stringId) {
						goto End
					}
				}

			case core.OBJECT_MODULE:
				ot := v.AsModule()

				if val, ok := ot.Environment.GetVar(nv.InternedId); ok {
					vm.pop()
					vm.stack[vm.stackTop] = val
					vm.stackTop++
				} else {
					name := core.GetStringValue(nv)
					vm.RunTimeError("Property '%s' not found.", name)
					goto End
				}

			default:
				name := core.GetStringValue(nv)
				vm.RunTimeError("Property '%s' not found.", name)
				goto End
			}

		// stack top is value, byte operand is the index of the property name in constants,
		// stack + 1 is the object to set the property on.
		case core.OP_SET_PROPERTY:

			val := vm.Peek(0)
			v := vm.Peek(1)
			if v.Type != core.VAL_OBJ && v.Type != core.VAL_VEC2 && v.Type != core.VAL_VEC3 && v.Type != core.VAL_VEC4 {
				vm.RunTimeError("Property not found.")
				goto End
			}
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			stringId := constants[idx].InternedId
			switch v.Obj.GetType() {
			case core.OBJECT_VEC2:
				// special case for Vec2, which has x and y properties
				switch stringId {
				case core.X:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec2()
					o.SetX(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.Y:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec2()
					o.SetY(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++

				default:
					vm.RunTimeError("Property '%s' not found.", core.GetStringValue(constants[idx]))
					goto End
				}
			case core.OBJECT_VEC3:
				// special case for Vec3, which has x, y and z properties
				switch stringId {
				case core.X:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec3()
					o.SetX(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.Y:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec3()
					o.SetY(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.Z:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec3()
					o.SetZ(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++

				default:
					vm.RunTimeError("Property '%s' not found.", core.GetStringValue(constants[idx]))
					goto End
				}
			case core.OBJECT_VEC4:
				// special case for Vec4, which has x, y, z and w properties
				switch stringId {
				case core.X, core.R:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec4()
					o.SetX(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.Y, core.G:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec4()
					o.SetY(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.Z, core.B:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec4()
					o.SetZ(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				case core.W, core.A:
					tmp := vm.pop() // pop the value
					vm.pop()        // pop the object
					o := v.AsVec4()
					o.SetW(tmp.AsFloat())
					vm.stack[vm.stackTop] = tmp // push the value back on the stack
					vm.stackTop++
				default:
					vm.RunTimeError("Property '%s' not found.", core.GetStringValue(constants[idx]))
					goto End
				}

			case core.OBJECT_INSTANCE:
				ot := v.AsInstance()
				ot.Fields[stringId] = val
				tmp := vm.pop()
				vm.pop()
				vm.stack[vm.stackTop] = tmp
				vm.stackTop++
			case core.OBJECT_MODULE:
				ot := v.AsModule()
				ot.Environment.SetVar(constants[idx].InternedId, val)
				tmp := vm.pop()
				vm.pop()
				vm.stack[vm.stackTop] = tmp
				vm.stackTop++
			default:
				vm.RunTimeError("Property '%s' not found.", core.GetStringValue(constants[idx]))
				goto End
			}

		// entered a try block, IP of the except block is encoded in the next 2 instructions
		// push an exception handler storing that info
		case core.OP_TRY:
			exceptIP := vm.readShort()
			frame.Handlers = &core.ExceptionHandler{
				ExceptIP: exceptIP,
				StackTop: vm.stackTop,
				Prev:     frame.Handlers,
			}

		// ended a try block OK, so pop the handler block
		case core.OP_END_TRY:
			frame.Handlers = frame.Handlers.Prev

		// marks the start of an exception handler block.  index of exception classname is in next instruction
		case core.OP_EXCEPT:
			frame.Ip++

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
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}

		case core.OP_CLASS:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := core.GetStringValue(constants[idx])
			vm.stack[vm.stackTop] = core.MakeObjectValue(core.MakeClassObject(name), false)
			vm.stackTop++

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
			return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE

		case core.OP_GET_SUPER:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := constants[idx].AsString()
			stringId := name.InternedId
			v := vm.pop()
			superclass := v.AsClass()

			if !vm.bindMethod(superclass, stringId) {
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}

		case core.OP_SUPER_INVOKE:
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			method := constants[idx]
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			superclass := vm.pop().AsClass()
			if !vm.invokeFromClass(superclass, method, int(argCount), false) {
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}

		// NJH added:
		// import a module by name (1) and alias (2)
		case core.OP_IMPORT:

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			mv := constants[idx]
			module := mv.AsString().Get()

			idx = vm.currCode[frame.Ip]
			frame.Ip++
			alv := constants[idx]
			alias := alv.AsString().Get()

			status := vm.importModule(module, alias)
			if status != INTERPRET_OK {
				return status, core.NIL_VALUE
			}

		// import functions from a module, or all functions
		// byte operand 1 is the index of the module name in constants
		// byte operand 2 is the number of functions to import
		// 0 = import all functions
		// byte operands 3..n are the indices of the functions to import
		case core.OP_IMPORT_FROM:

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			mv := constants[idx]
			module := mv.AsString().Get()

			lv := vm.currCode[frame.Ip]
			frame.Ip++
			length := int(lv)

			status := vm.importModule(module, module)
			if status != INTERPRET_OK {
				return status, core.NIL_VALUE
			}

			if length == 0 {
				if !vm.importFunctionFromModule(module, "__all__") {
					vm.RunTimeError("Failed to import function '%s' from module '%s'.", "__all__", module)
					return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
				}
			} else {
				for i := 0; i < length; i++ {
					idx = vm.currCode[frame.Ip]
					frame.Ip++
					fv := constants[idx]
					name := fv.AsString().Get()
					if !vm.importFunctionFromModule(module, name) {
						vm.RunTimeError("Failed to import function '%s' from module '%s'.", name, module)
						return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
					}
				}
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
					if toString, ok := ot.Class.Methods[core.TO_STRING]; ok {
						vm.call(toString.AsClosure(), 0)
						continue
					}
					s = v.String()
				}
			}
			vm.pop()
			vm.stack[vm.stackTop] = core.MakeStringObjectValue(s, false)
			vm.stackTop++

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

		// ### foreach ( var a in iterable ) ###
		// local slot, iterator slot, end of foreach in next 3 instructions
		// can handle native iterable objects (list, string) or lox class instances
		// with __iter__ method returning an iterator object implementing __next__ method

		case core.OP_FOREACH:

			slot := vm.readByte()
			iterableSlot := vm.readByte()
			jumpToEnd := vm.readShort()
			iterable := vm.stack[frame.Slots+int(iterableSlot)]

			if iterable.Type != core.VAL_OBJ {
				vm.RunTimeError("Foreach requires an iterable object.")
				goto End
			}
			// native iterable object (list, string )
			o, ok := iterable.Obj.(core.Iterable)
			if ok {
				iterval, _ := o.GetIterator()
				vm.stack[frame.Slots+int(iterableSlot)] = iterval
				val := iterval.AsIterator().Next()
				if val.Type == core.VAL_NIL {
					// empty iterable, jump to end
					frame.Ip += int(jumpToEnd - 2)
					continue
				}
				vm.stack[frame.Slots+int(slot)] = val

			} else if iterable.IsInstanceObject() {
				// lox class instance with iterator method?
				// we need to call it to get an iterator object
				// so set a new foreach state on the vm indicting we are running a method
				// and waiting for an iterator
				_, ok := iterable.AsInstance().Class.Methods[core.ITER]
				if ok {
					vm.invoke(ITER_METHOD, 0)
					vm.foreachState = NewVMForeachState(vm.foreachState, frame.Slots+int(slot), frame.Slots+int(iterableSlot), int(frame.Ip), frame.Ip+int(jumpToEnd-2))
					continue
				}
			} else {
				vm.RunTimeError("Foreach requires an iterable object.")
				goto End
			}
		case core.OP_NEXT:

			jumpToStart := vm.readShort()
			iterSlot := frame.Slots + int(vm.readByte())
			iterVal := vm.stack[iterSlot]
			if vm.foreachState == nil {
				val := iterVal.AsIterator().Next()
				if val.Type != core.VAL_NIL {
					vm.stack[iterSlot-1] = val
					frame.Ip -= int(jumpToStart + 1)
				}
			} else {
				vm.stack[vm.stackTop] = iterVal
				vm.stackTop++ // push the iterator object onto the stack
				vm.invoke(NEXT_METHOD, 0)
			}

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
				vm.stack[vm.stackTop] = rv
				vm.stackTop++
			case core.OBJECT_LIST:
				rv := b.AsList().Contains(a)
				vm.stack[vm.stackTop] = rv
				vm.stackTop++
			}
		case core.OP_BREAKPOINT:
			// hit a breakpoint, pause execution
			vm.pauseExecution()

		//unpack a tuple or list on stack top.
		// byte will be the number of items to unpack
		case core.OP_UNPACK:

			count := int(vm.readByte())
			if count == 0 {
				vm.RunTimeError("Unpack count cannot be zero.")
				goto End
			}
			top := vm.Peek(0)
			if top.Type != core.VAL_OBJ {
				vm.RunTimeError("Unpack requires a list or tuple on stack top.")
				goto End
			}
			if top.Obj.GetType() != core.OBJECT_LIST {
				vm.RunTimeError("Unpack requires a list or tuple on stack top.")
				goto End
			}
			// check we have enough items in the list or tuple
			lo := top.AsList()
			if lo.GetLength() != int(count) {
				vm.RunTimeError("Unpack count %d not the same as list/tuple size %d.", count, lo.GetLength())
				goto End
			}
			vm.pop() // pop the list/tuple from the stack
			// now push the items onto the stack in reverse order
			for i := range count {
				item, _ := lo.Index(int(i))
				vm.stack[vm.stackTop] = item
				vm.stackTop++
			}

		default:
			vm.RunTimeError("Invalid Opcode")
			goto End
		}
	End:

		if vm.ErrorMsg != "" {
			if !vm.RaiseExceptionByName("RunTimeError", vm.ErrorMsg) {
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}
		}
	}
	//return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
}

// natively raise an exception given a name:
// - get the class object for the name from globals
// - make an instance of the class and set the message on it
// - pass the instance to raiseException
// used for vm raising errors that can be handled in lox e.g EOFError when reading a file
func (vm *VM) RaiseExceptionByName(name string, msg string) bool {

	classVal := vm.builtIns[core.InternName(name)]
	classObj := classVal.Obj
	instance := core.MakeInstanceObject(classObj.(*core.ClassObject))
	instance.Fields[core.MSG] = core.MakeStringObjectValue(msg, false)
	return vm.raiseException(core.MakeObjectValue(instance, false))
}

func (vm *VM) raiseException(err core.Value) bool {

	for {
		vm.appendStackTrace()
		handler := vm.frame().Handlers
		if handler != nil {

			vm.stackTop = handler.StackTop
			vm.stack[vm.stackTop] = err
			vm.stackTop++
			// jump to handler IP
			vm.frame().Ip = int(handler.ExceptIP)
		inner:
			for {
				// get handler classname
				vm.frame().Ip += 2
				idx := vm.getCode()[vm.frame().Ip-1]
				function := vm.frame().Closure.Function
				id := function.Chunk.Constants[idx].InternedId
				v, ok := function.Environment.GetVar(id)
				if !ok {
					v, ok = vm.builtIns[id]
					if !ok {
						name := core.GetStringValue(function.Chunk.Constants[idx])
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
					vm.frame().Handlers = handler.Prev

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
			vm.RunTimeError("Uncaught exception: %s : %s ", exc.Class, exc.Fields[core.MSG])
			return false
		}
	}
}

func (vm *VM) nextHandler() bool {

	code := vm.getCode()
	for {
		vm.frame().Ip++
		if code[vm.frame().Ip] == core.OP_END_EXCEPT {
			if code[vm.frame().Ip+1] == core.OP_EXCEPT {
				vm.frame().Ip += 1
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
	vm.stackTop = vm.frames[vm.frameCount].Slots
	return true
}

func (vm *VM) appendStackTrace() {

	frame := vm.frame()
	function := frame.Closure.Function
	where, script := "", ""
	if function.Name.Get() == "" {
		script = vm.script
		where = "<module>"
	} else {
		script = function.Chunk.Filename
		where = function.Name.Get()
	}
	line := function.Chunk.Lines[frame.Ip]

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

func (vm *VM) importModule(moduleName string, alias string) InterpretResult {

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
	subfn := subvm.frames[0].Closure.Function
	mo := core.MakeModuleObject(moduleName, *subfn.Environment)
	v := core.MakeObjectValue(mo, false)
	vm.frame().Closure.Function.Environment.SetVar(core.InternName(alias), v)
	return INTERPRET_OK
}

func (vm *VM) importFunctionFromModule(module string, name string) bool {

	moduleId := core.InternName(module)
	nameId := core.InternName(name)

	moduleVal, ok := vm.frame().Closure.Function.Environment.GetVar(moduleId)
	if !ok {
		vm.RunTimeError("Module '%s' is not imported.", module)
		return false
	}
	if name == "__all__" {
		// import all functions from the module
		moduleObj := moduleVal.AsModule()
		for k, v := range moduleObj.Environment.Vars {
			if v.Type == core.VAL_OBJ && v.Obj.GetType() == core.OBJECT_CLOSURE {
				vm.frame().Closure.Function.Environment.SetVar(k, v)
			}
		}
		return true
	} else {

		moduleObj := moduleVal.AsModule()
		fn, ok := moduleObj.Environment.GetVar(nameId)
		if !ok {
			vm.RunTimeError("Function '%s' not found in module '%s'.", name, module)
			return false
		}
		if fn.Type != core.VAL_OBJ || fn.Obj.GetType() != core.OBJECT_CLOSURE {
			vm.RunTimeError("Function '%s' not found in module '%s'.", name, module)
			return false
		}
		vm.frame().Closure.Function.Environment.SetVar(nameId, fn)
		return true
	}

}

func (vm *VM) createList(frame *core.CallFrame) {

	itemCount := int(vm.currCode[frame.Ip])
	frame.Ip++
	list := []core.Value{}

	for i := 0; i < itemCount; i++ {
		list = append([]core.Value{vm.pop()}, list...) // reverse order
	}
	lo := core.MakeListObject(list, false)
	vm.stack[vm.stackTop] = core.MakeObjectValue(lo, false)
	vm.stackTop++
}

func (vm *VM) createTuple(frame *core.CallFrame) {

	itemCount := int(vm.currCode[frame.Ip])
	frame.Ip++
	list := []core.Value{}

	for i := 0; i < itemCount; i++ {
		list = append([]core.Value{vm.pop()}, list...) // reverse order
	}
	lo := core.MakeListObject(list, true)
	vm.stack[vm.stackTop] = core.MakeObjectValue(lo, true)
	vm.stackTop++
}

func (vm *VM) createDict(frame *core.CallFrame) {

	itemCount := int(vm.currCode[frame.Ip])
	frame.Ip++
	dict := map[string]core.Value{}

	for i := 0; i < itemCount; i++ {
		value := vm.pop()
		key := vm.pop()
		dict[key.AsString().Get()] = value
	}
	do := core.MakeDictObject(dict)
	vm.stack[vm.stackTop] = core.MakeObjectValue(do, false)
	vm.stackTop++
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
			vm.stack[vm.stackTop] = lo
			vm.stackTop++
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
			vm.stack[vm.stackTop] = so
			vm.stackTop++
			return true

		case core.OBJECT_DICT:

			key := iv.AsString().Get()
			t := sv.AsDict()
			so, err := t.Get(key)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.stack[vm.stackTop] = so
			vm.stackTop++
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
			vm.stack[vm.stackTop] = lo
			vm.stackTop++
			return true

		} else if lv.Obj.GetType() == core.OBJECT_STRING {
			so, err := lv.AsString().Slice(from_idx, to_idx)
			if err != nil {
				vm.RunTimeError("%v", err)
				return false
			}
			vm.stack[vm.stackTop] = so
			vm.stackTop++
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

func (vm *VM) binarySubtract() bool {

	v2 := vm.pop()
	v1 := vm.pop()

	switch v2.Type {
	case core.VAL_INT:
		switch v1.Type {
		case core.VAL_INT:
			vm.stack[vm.stackTop] = core.MakeIntValue(v1.Int-v2.Int, false)
			vm.stackTop++
			return true
		case core.VAL_FLOAT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float-float64(v2.Int), false)
			vm.stackTop++
			return true
		}

	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(float64(v1.Int)-v2.Float, false)
			vm.stackTop++
			return true
		case core.VAL_FLOAT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float-v2.Float, false)
			vm.stackTop++
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
			vm.stack[vm.stackTop] = core.MakeIntValue(v1.Int*v2.Int, false)
			vm.stackTop++
		case core.VAL_FLOAT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float*float64(v2.Int), false)
			vm.stackTop++
		case core.VAL_OBJ:
			if !v1.IsStringObject() {
				vm.RunTimeError("Invalid operand for multiply.")
				return false
			}
			s := v1.AsString().Get()
			vm.stack[vm.stackTop] = vm.stringMultiply(s, v2.Int)
			vm.stackTop++
		default:
			vm.RunTimeError("Invalid operand for multiply.")
			return false
		}
	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(float64(v1.Int)*v2.Float, false)
			vm.stackTop++
		case core.VAL_FLOAT:
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float*v2.Float, false)
			vm.stackTop++
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
			vm.stack[vm.stackTop] = vm.stringMultiply(s, v1.Int)
			vm.stackTop++
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
			if v2.Int == 0 {
				vm.RunTimeError("Division by zero")
				return false
			}
			vm.stack[vm.stackTop] = core.MakeIntValue(v1.Int/v2.Int, false)
			vm.stackTop++
			return true
		case core.VAL_FLOAT:
			if v2.Int == 0 {
				vm.RunTimeError("Division by zero")
				return false
			}
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float/float64(v2.Int), false)
			vm.stackTop++
			return true
		}

	case core.VAL_FLOAT:
		switch v1.Type {
		case core.VAL_INT:
			if v2.Float == 0.0 {
				vm.RunTimeError("Division by zero")
				return false
			}
			vm.stack[vm.stackTop] = core.MakeFloatValue(float64(v1.Int)/v2.Float, false)
			vm.stackTop++
			return true
		case core.VAL_FLOAT:
			if v2.Float == 0.0 {
				vm.RunTimeError("Division by zero")
				return false
			}
			vm.stack[vm.stackTop] = core.MakeFloatValue(v1.Float/v2.Float, false)
			vm.stackTop++
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
	vm.stack[vm.stackTop] = core.MakeIntValue(v1.Int%v2.Int, false)
	vm.stackTop++

	return true
}

func (vm *VM) stringMultiply(s string, x int) core.Value {

	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return core.MakeStringObjectValue(rv, false)
}
func (vm *VM) pauseExecution() {

	fmt.Println("⚠️  BREAKPOINT HIT")
	fmt.Println("Stack:", vm.stack[:vm.stackTop])
	// If you track them
	// runtime.Stack can be used to print the Go stack trace if desired
	// buf := make([]byte, 1<<16)
	// runtime.Stack(buf, true)
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

func (vm *VM) showStack() {

	core.LogFmt(core.TRACE, "                                                         ")
	for i := 1; i < vm.stackTop; i++ {
		v := vm.stack[i]
		s := v.String()

		im := ""
		if v.Immutable() {
			im = "(c)"
		}
		if i > vm.frame().Slots {
			core.LogFmt(core.TRACE, "%% %s%s %%", s, im)
		} else {
			core.LogFmt(core.TRACE, "| %s%s |", s, im)
		}
	}
	core.LogFmt(core.TRACE, "\n")
}
func (vm *VM) showGlobals() {
	if vm.frame().Closure.Function.Environment == nil {
		fmt.Println("No globals (nil environment)")
		return
	}
	core.LogFmt(core.TRACE, "globals: %s \n", vm.frame().Closure.Function.Environment.Name)
	for k, v := range vm.frame().Closure.Function.Environment.Vars {
		core.LogFmt(core.TRACE, "%s -> %s  \n", core.NameFromID(k), v)
	}
	//for k, v := range vm.Environments.builtins {
	//	core.LogFmt(core.TRACE,"%s -> %s  \n", k, v)
	//}
}
