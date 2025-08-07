package vm

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

type VMRunMode int

const (
	RUN_TO_COMPLETION VMRunMode = iota
	RUN_CURRENT_FUNCTION
)

const (
	FRAMES_MAX          int     = 64
	STACK_MAX           int     = FRAMES_MAX * 256
	GC_COLLECT_INTERVAL float64 = 5
)

type VM struct {
	script         string
	source         string
	stack          [STACK_MAX]core.Value
	stackTop       int
	Frames         [FRAMES_MAX]*core.CallFrame
	frameCount     int
	currCode       []uint8 // current code being executed
	Starttime      time.Time
	lastGC         time.Time
	openUpValues   *core.UpvalueObject // head of list
	args           []string
	ErrorMsg       string
	stackTrace     []string
	ModuleImport   bool
	BuiltIns       map[int]core.Value         // global built-in functions
	BuiltInModules map[int]*core.ModuleObject // global built-in modules - need to be imported before use

	// Debug hook: called with (vm, event, data) at opcode, call, return
	// opcode events will have data as the opcode byte,
	// call events will have data as the closure object being called,
	// return events will have data as the return Value.
	DebugHook func(vm interface{}, event core.DebugEvent, data any)
}

var _ debug.VMInspector = (*VM)(nil)

var ITER_METHOD = core.MakeStringObjectValue("__iter__", true)
var NEXT_METHOD = core.MakeStringObjectValue("__next__", true)

//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------

var globalModuleSource = map[string]string{}
var globalModules = map[string]*core.ModuleObject{}

// NewVM creates and initializes a new virtual machine instance for executing Lox scripts.
// It sets up the initial state including stack, frames, and optionally defines built-in functions.
func NewVM(script string, defineBuiltIns bool) *VM {

	vm := &VM{
		script:         script,
		Starttime:      time.Now(),
		lastGC:         time.Now(),
		openUpValues:   nil,
		args:           []string{},
		ErrorMsg:       "",
		stackTrace:     []string{},
		BuiltIns:       make(map[int]core.Value),
		BuiltInModules: make(map[int]*core.ModuleObject),
	}
	vm.resetStack()
	if defineBuiltIns && !core.DebugCompileOnly {
		DefineBuiltIns(vm)
	}
	return vm
}

//------------------------------------------------------------------------------------------

// SetArgs sets the command-line arguments that will be available to the running Lox script.
func (vm *VM) SetArgs(args []string) {
	vm.args = args
}

//------------------------------------------------------------------------------------------

// Interpret compiles and executes the given Lox source code, returning the result and any output.
// It handles the full lifecycle from compilation through execution, including module import preparation.
func (vm *VM) Interpret(source string, module string) (InterpretResult, string) {

	core.LogFmtLn(core.INFO, "VM %s starting execution\n", vm.script)
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
	res, val := vm.run(RUN_TO_COMPLETION)
	core.LogFmtLn(core.INFO, "VM %s finished execution\n", vm.script)

	return res, val.String()
}

//------------------------------------------------------------------------------------------

// Stack returns the value at the specified index in the VM's stack, or NIL_VALUE if index is invalid.
// Used for debugging and inspection purposes.
func (vm *VM) Stack(index int) core.Value {

	if index < 0 || index >= vm.stackTop {
		return core.NIL_VALUE
	}
	return vm.stack[index]
}

//------------------------------------------------------------------------------------------

// Args returns the command-line arguments that were set for this VM instance.
func (vm *VM) Args() []string {

	return vm.args
}

//------------------------------------------------------------------------------------------

// StartTime returns the timestamp when this VM instance was created and started execution.
func (vm *VM) StartTime() time.Time {

	return vm.Starttime
}

// FileName extracts and returns the base filename of the script being executed by this VM.
func (vm *VM) FileName() string {

	// returns the script name
	if vm.script == "" {
		return "<unknown>"
	}
	return filepath.Base(vm.script)
}

//------------------------------------------------------------------------------------------

// RunTimeError stores a runtime error message in the VM for later exception handling.
// This is typically called when an operation fails during bytecode execution.
func (vm *VM) RunTimeError(format string, args ...any) {

	vm.ErrorMsg = fmt.Sprintf(format, args...)
}

//------------------------------------------------------------------------------------------

// Peek looks at a value on the stack at the specified distance from the top without removing it.
// Distance 0 means the top of the stack, 1 means one below the top, etc.
func (vm *VM) Peek(dist int) core.Value {

	return vm.stack[(vm.stackTop-1)-dist]
}

//------------------------------------------------------------------------------------------

// Frame returns the current call frame (the topmost frame on the call stack).
// Exported Frame method
func (vm *VM) Frame() *core.CallFrame {
	return vm.Frames[vm.frameCount-1]
}

// FrameCount returns the number of active call frames on the call stack.
func (vm *VM) FrameCount() int {
	return vm.frameCount
}

//------------------------------------------------------------------------------------------

// FrameAt returns the call frame at the specified index, or nil if the index is invalid.
// Used for debugging and stack trace generation.
func (vm *VM) FrameAt(index int) *core.CallFrame {
	if index < 0 || index >= vm.frameCount {
		return nil
	}
	return vm.Frames[index]
}

//------------------------------------------------------------------------------------------

// StackTop returns the current stack pointer (number of values currently on the stack).
func (vm *VM) StackTop() int {
	return vm.stackTop
}

//------------------------------------------------------------------------------------------

// CurrCode returns the current bytecode instruction being executed at the instruction pointer.
func (vm *VM) CurrCode() uint8 {
	return vm.currCode[vm.frame().Ip]
}

//------------------------------------------------------------------------------------------

// ShowStack returns a formatted string representation of the current stack contents.
// Exported ShowStack returns stack as string
func (vm *VM) ShowStack() string {

	var sb strings.Builder
	for i := 1; i < vm.stackTop; i++ {
		v := vm.stack[i]
		s := v.String()
		im := ""
		if v.Immutable() {
			im = "[const]"
		}
		localname := vm.LocalName(i-1, vm.frame().Ip)
		if localname != "" {
			localname = fmt.Sprintf(" (%s)", localname)
		}
		if i >= vm.frame().Slots {
			slot := i - vm.frame().Slots
			sb.WriteString(fmt.Sprintf("%04d->%s%s%s\n", slot, s, im, localname))
		} else {
			sb.WriteString(fmt.Sprintf("      %s%s%s\n", s, im, localname))
		}
	}
	return sb.String()
}

// LocalName looks up the name of a local variable at the given slot and instruction pointer.
// Returns empty string if no local variable name is found for the given position.
// ------------------------------------------------------------------------------------------
func (vm *VM) LocalName(slot int, ip int) string {
	for _, info := range vm.frame().Closure.Function.Chunk.LocalVars {
		if info.Slot == slot && ip >= info.StartIp && (info.EndIp == -1 || ip < info.EndIp) {
			return info.Name
		}
	}
	return ""
}

//------------------------------------------------------------------------------------------

// Script returns the name/path of the script file being executed by this VM.
func (vm *VM) Script() string {
	// returns the script name
	return vm.script
}

// GetGlobals returns the global environment/scope of the currently executing function.
// ------------------------------------------------------------------------------------------
func (vm *VM) GetGlobals() *core.Environment {
	if vm.frame().Closure.Function.Environment == nil {
		return nil
	}
	return vm.frame().Closure.Function.Environment
}

//------------------------------------------------------------------------------------------

// frame returns the current call frame (internal helper function).
// This is the private version of Frame() for internal VM use.
func (vm *VM) frame() *core.CallFrame {

	return vm.Frames[vm.frameCount-1]
}

//------------------------------------------------------------------------------------------

// getCode returns the bytecode array of the currently executing function.
func (vm *VM) getCode() []uint8 {

	return vm.frame().Closure.Function.Chunk.Code
}

//------------------------------------------------------------------------------------------

// resetStack resets the VM's execution stack and call frames to their initial empty state.
func (vm *VM) resetStack() {

	vm.stackTop = 0
	vm.frameCount = 0
}

//------------------------------------------------------------------------------------------

// push adds a value to the top of the VM's execution stack.
func (vm *VM) push(v core.Value) {

	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

//------------------------------------------------------------------------------------------

// pop removes and returns the value from the top of the VM's execution stack.
// Returns NIL_VALUE if the stack is empty.
func (vm *VM) pop() core.Value {

	if vm.stackTop == 0 {
		return core.NIL_VALUE
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

//------------------------------------------------------------------------------------------

// run executes the main bytecode interpretation loop, processing instructions until completion or error.
// The mode parameter controls whether to run to completion or just the current function.
// main interpreter loop
func (vm *VM) run(mode VMRunMode) (InterpretResult, core.Value) {
	counter := 0
	vm.ErrorMsg = ""
	startFrame := vm.frameCount

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
		if vm.DebugHook != nil {
			vm.DebugHook(vm, core.DebugEventOpcode, inst)
		}

		frame.Ip++
		switch inst {

		case core.OP_NOOP:

		case core.OP_EQUAL:
			// Pop two values from stack, compare for equality, push boolean result

			a := vm.pop()
			b := vm.pop()
			vm.stack[vm.stackTop] = core.MakeBooleanValue(core.ValuesEqual(a, b, false), false)
			vm.stackTop++

		case core.OP_ONE:
			// Push integer constant 1 onto the stack
			vm.stack[vm.stackTop] = core.MakeIntValue(1, false)
			vm.stackTop++

		case core.OP_GREATER:
			// Pop two values, compare if first > second, push boolean result

			v2 := vm.pop()
			v1 := vm.pop()

			if v1.IsStringObject() && v2.IsStringObject() {
				vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsString().Get() > v2.AsString().Get(), false)
				vm.stackTop++
				continue
			}

			if !v1.IsNumber() || !v2.IsNumber() {
				vm.RunTimeError("Operands must be numbers")
				goto End
			}
			vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsFloat() > v2.AsFloat(), false)
			vm.stackTop++

		case core.OP_LESS:
			// Pop two values, compare if first < second, push boolean result

			v2 := vm.pop()
			v1 := vm.pop()

			if v1.IsStringObject() && v2.IsStringObject() {
				vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsString().Get() < v2.AsString().Get(), false)
				vm.stackTop++
				continue
			}

			if !v1.IsNumber() || !v2.IsNumber() {
				vm.RunTimeError("Operands must be numbers")
				goto End
			}
			vm.stack[vm.stackTop] = core.MakeBooleanValue(v1.AsFloat() < v2.AsFloat(), false)
			vm.stackTop++

		case core.OP_INC_LOCAL:
			// Increment local variable at specified slot by 1 (handles int and float types)

			slot := int(vm.currCode[frame.Ip])
			frame.Ip++
			if vm.stack[frame.Slots+slot].Immutable() {
				vm.RunTimeError("Cannot increment const local.")
				goto End
			}
			v := vm.stack[frame.Slots+slot]
			if v.IsInt() {
				vm.stack[frame.Slots+slot] = core.MakeIntValue(vm.stack[frame.Slots+slot].AsInt()+1, false)
				continue
			}
			if v.IsFloat() {
				vm.stack[frame.Slots+slot] = core.MakeFloatValue(vm.stack[frame.Slots+slot].AsFloat()+1, false)
				continue
			}
			vm.RunTimeError("Cannot increment non-number local variable.")
			goto End

		case core.OP_PRINT:
			// Pop value from stack and print it to stdout
			// compiler ensures stack top will be a string object via core.OP_STR
			v := vm.pop()
			fmt.Printf("%s\n", v.AsString().Get())

		case core.OP_POP:
			// Pop and discard the top value from the stack

			_ = vm.pop()

		case core.OP_DEFINE_GLOBAL:
			// Define a new global variable with name from constants and value from stack
			// name = constant at operand index

			idx := vm.currCode[frame.Ip]
			frame.Ip++

			value := vm.Peek(0)
			//DumpValue("Define global", value)
			function.Environment.SetVar(constants[idx].InternedId, core.Mutable(value)) // make sure variable is mutable
			vm.pop()

		case core.OP_DEFINE_GLOBAL_CONST:
			// Define a new global constant with name from constants and value from stack
			// name = constant at operand index

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			id := constants[idx].InternedId
			function.Environment.SetVar(id, vm.Peek(0))
			v, _ := function.Environment.GetVar(id)
			function.Environment.SetVar(id, core.Immutable(v))
			vm.pop()

		case core.OP_GET_GLOBAL:
			// Look up global variable by name from constants and push its value onto stack
			// name = constant at operand index

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			id := constants[idx].InternedId

			value, ok := function.Environment.GetVar(id)
			//DumpValue("Get global", value)
			if !ok {
				value, ok = vm.BuiltIns[id]
				if !ok {
					name := core.GetStringValue(constants[idx])
					vm.RunTimeError("Undefined variable %s", name)
					goto End
				}
			}
			vm.stack[vm.stackTop] = value
			vm.stackTop++

		case core.OP_SET_GLOBAL:
			// Assign value from stack top to existing global variable (must exist and be mutable)
			// name = constant at operand index

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
			// Get local variable from stack at specified slot and push onto stack top

			slot_idx := int(vm.currCode[frame.Ip])
			frame.Ip++
			vm.stack[vm.stackTop] = vm.stack[frame.Slots+slot_idx]
			vm.stackTop++

		case core.OP_SET_LOCAL:
			// Assign value from stack top to local variable at specified slot (must be mutable)

			val := vm.Peek(0)
			slot_idx := int(vm.currCode[frame.Ip])
			frame.Ip++
			if vm.stack[frame.Slots+slot_idx].Immutable() {
				vm.RunTimeError("Cannot assign to const local.")
				goto End
			}
			vm.stack[frame.Slots+slot_idx] = core.Mutable(val)

		case core.OP_JUMP_IF_FALSE:
			// Conditional jump: if stack top is falsy, jump forward by offset amount

			offset := vm.readShort()
			if vm.isFalsey(vm.Peek(0)) {
				frame.Ip += int(offset)
			}

		case core.OP_JUMP:
			// Unconditional jump forward by offset amount (used for control flow)

			offset := vm.readShort()
			core.LogFmtLn(core.DEBUG, "Jumping %d from %d to %d\n", offset, frame.Ip, frame.Ip+int(offset))
			frame.Ip += int(offset)

		case core.OP_GET_UPVALUE:
			// Get upvalue (closed-over variable) at specified slot and push onto stack
			slot := vm.currCode[frame.Ip]
			frame.Ip++
			valIdx := frame.Closure.Upvalues[slot].Location
			vm.stack[vm.stackTop] = *valIdx
			vm.stackTop++

		case core.OP_SET_UPVALUE:
			// Set upvalue (closed-over variable) at specified slot to stack top value
			slot := vm.currCode[frame.Ip]
			frame.Ip++
			*(frame.Closure.Upvalues[slot].Location) = vm.Peek(0)

		case core.OP_CLOSE_UPVALUE:
			// Close upvalue at specified stack position and pop the value
			vm.closeUpvalues(vm.stackTop - 1)
			vm.pop()

		case core.OP_CONSTANT:
			// Load constant at specified index from constants table and push onto stack

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			constant := constants[idx]
			vm.stack[vm.stackTop] = constant
			vm.stackTop++

		case core.OP_CALL:
			// Call function/method with specified argument count (callable object is after args on stack)
			// arg count is operand, callable object is on stack after arguments, result will be stack top
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			if !vm.callValue(vm.Peek(int(argCount)), int(argCount)) {
				goto End
			}

		case core.OP_ADD_NUMERIC:
			// Pop two values from stack, add them (handles int, float), push result

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
				vm.RunTimeError("Addition type mismatch: %s + %s", v1.String(), v2.String())
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
				vm.RunTimeError("Addition type mismatch: %s + %s", v1.String(), v2.String())
				goto End
			}

			vm.RunTimeError("Invalid operands for addition: %s + %s", v1.String(), v2.String())
			goto End

		case core.OP_ADD_VECTOR:
			// Pop two vector values from stack, add them (handles vec2, vec3, vec4), push result
			v2 := vm.pop()
			v1 := vm.pop()
			switch v2.Type {
			case core.VAL_VEC2:
				if v1.Type != core.VAL_VEC2 {
					vm.RunTimeError("Addition type mismatch: %s + %s", v1.String(), v2.String())
					goto End
				}
				v3 := v1.AsVec2().Add(v2.AsVec2())
				vm.stack[vm.stackTop] = core.MakeVec2Value(v3.X, v3.Y, false)
				vm.stackTop++
				continue

			case core.VAL_VEC3:
				if v1.Type != core.VAL_VEC3 {
					vm.RunTimeError("Addition type mismatch: %s + %s", v1.String(), v2.String())
					goto End
				}
				v3 := v1.AsVec3().Add(v2.AsVec3())
				vm.stack[vm.stackTop] = core.MakeVec3Value(v3.X, v3.Y, v3.Z, false)
				vm.stackTop++
				continue

			case core.VAL_VEC4:
				if v1.Type != core.VAL_VEC4 {
					vm.RunTimeError("Addition type mismatch: %s + %s", v1.String(), v2.String())
					goto End
				}
				v3 := v1.AsVec4().Add(v2.AsVec4())
				vm.stack[vm.stackTop] = core.MakeVec4Value(v3.X, v3.Y, v3.Z, v3.W, false)
				vm.stackTop++
				continue
			}

			vm.RunTimeError("Invalid operands for vector addition: %s + %s", v1.String(), v2.String())
			goto End

		case core.OP_ADD_NN:
			// optimised x = x + y for numbers: byte 1, byte 2, numbers to add
			slotDest := vm.readByte()
			slotInc := vm.readByte()
			base := frame.Slots
			valA := vm.stack[base+int(slotDest)]
			valB := vm.stack[base+int(slotInc)]

			// Immediate specializations for common cases
			if valA.Type == core.VAL_INT && valB.Type == core.VAL_INT {
				// Patch and execute specialized version immediately
				vm.patchInstruction(frame.Ip-3, core.OP_ADD_II)
				vm.stack[base+int(slotDest)] = core.MakeIntValue(valA.Int+valB.Int, false)
				continue
			}
			if valA.Type == core.VAL_FLOAT && valB.Type == core.VAL_FLOAT {
				// Patch and execute specialized version immediately
				vm.patchInstruction(frame.Ip-3, core.OP_ADD_FF)
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(valA.Float+valB.Float, false)
				continue
			}

			switch valB.Type {
			case core.VAL_INT:
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(valA.Float+float64(valB.Int), false)

			case core.VAL_FLOAT:
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(float64(valA.Int)+valB.Float, false)
			}

		case core.OP_ADD_II:
			// optimised x=x+y for local ints: byte 1, byte 2, numbers to add

			frm := vm.Frames[vm.frameCount-1]
			frm.Ip += 2
			slotDest := vm.currCode[frm.Ip-2]
			slotInc := vm.currCode[frm.Ip-1]

			base := frm.Slots
			vm.stack[base+int(slotDest)] = core.Value{
				Type:  core.VAL_INT,
				Int:   vm.stack[base+int(slotDest)].Int + vm.stack[base+int(slotInc)].Int,
				Immut: false,
			}
			continue

		case core.OP_ADD_FF:
			// optimised x=x+y for local floats: byte 1, byte 2, numbers to add
			frm := vm.Frames[vm.frameCount-1]
			frm.Ip += 2
			slotDest := vm.currCode[frm.Ip-2]
			slotInc := vm.currCode[frm.Ip-1]

			base := frm.Slots
			vm.stack[base+int(slotDest)] = core.Value{
				Type:  core.VAL_FLOAT,
				Float: vm.stack[base+int(slotDest)].Float + vm.stack[base+int(slotInc)].Float,
				Immut: false,
			}
			continue

		case core.OP_INCR_CONST_N:
			// optimised x = x + c for numbers: byte 1 local, byte 2 constant, numbers to add
			slotDest := vm.readByte()
			slotIncIndex := vm.readByte()
			base := frame.Slots
			valDest := vm.stack[base+int(slotDest)]
			constVal := frame.Closure.Function.Chunk.Constants[slotIncIndex]

			core.LogFmtLn(core.DEBUG, "incr_const_n: dest tpe %d, const type %d\n", valDest.Type, constVal.Type)
			// Immediate specializations for common cases
			if valDest.Type == core.VAL_INT && constVal.Type == core.VAL_INT {
				// Patch and execute specialized version immediately
				vm.patchInstruction(frame.Ip-3, core.OP_INCR_CONST_I)
				vm.stack[base+int(slotDest)] = core.MakeIntValue(valDest.Int+constVal.Int, false)
				continue
			}
			if valDest.Type == core.VAL_FLOAT && constVal.Type == core.VAL_FLOAT {
				// Patch and execute specialized version immediately
				vm.patchInstruction(frame.Ip-3, core.OP_INCR_CONST_F)
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(valDest.Float+constVal.Float, false)
				continue
			}

			switch constVal.Type {
			case core.VAL_INT:
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(valDest.Float+float64(constVal.Int), false)

			case core.VAL_FLOAT:
				vm.stack[base+int(slotDest)] = core.MakeFloatValue(float64(valDest.Int)+constVal.Float, false)
			}

		case core.OP_INCR_CONST_I:
			// optimised x=x+c for local ints: byte 1 local, byte 2 constant, numbers to add
			frm := vm.Frames[vm.frameCount-1]
			frm.Ip += 2
			slotVar := vm.currCode[frm.Ip-2]
			constIndex := vm.currCode[frm.Ip-1]

			base := frm.Slots
			constVal := frm.Closure.Function.Chunk.Constants[constIndex].Int

			// Direct integer increment
			vm.stack[base+int(slotVar)] = core.Value{
				Type:  core.VAL_INT,
				Int:   vm.stack[base+int(slotVar)].Int + constVal,
				Immut: false,
			}
			continue

		case core.OP_INCR_CONST_F:
			// optimised x=x+c for local ints: byte 1 local, byte 2 constant, numbers to add
			frm := vm.Frames[vm.frameCount-1]
			frm.Ip += 2
			slotVar := vm.currCode[frm.Ip-2]
			constIndex := vm.currCode[frm.Ip-1]

			base := frm.Slots
			constVal := frm.Closure.Function.Chunk.Constants[constIndex].Float

			// Direct integer increment
			vm.stack[base+int(slotVar)] = core.Value{
				Type:  core.VAL_FLOAT,
				Float: vm.stack[base+int(slotVar)].Float + constVal,
				Immut: false,
			}
			continue

		case core.OP_CONCAT:
			v2 := vm.pop()
			v1 := vm.pop()
			switch v2.Type {

			case core.VAL_OBJ:
				ov2 := v2.Obj
				switch ov2.GetType() {

				case core.OBJECT_STRING:
					if v1.Type != core.VAL_OBJ {
						vm.RunTimeError("Concatenation type mismatch: %s + %s", v1.String(), v2.String())
						goto End
					}
					ov1 := v1.Obj
					if ov1.GetType() == core.OBJECT_STRING {
						vm.stack[vm.stackTop] = core.MakeStringObjectValue(v1.AsString().Get()+v2.AsString().Get(), false)
						vm.stackTop++
						continue
					}
					vm.RunTimeError("Concatenation type mismatch: %s + %s", v1.String(), v2.String())
					goto End
				case core.OBJECT_LIST:
					if v1.Type != core.VAL_OBJ {
						vm.RunTimeError("Concatenation type mismatch: %s + %s", v1.String(), v2.String())
						goto End
					}
					ov1 := v1.Obj
					if ov1.GetType() == core.OBJECT_LIST {
						lo := ov1.(*core.ListObject).Add(ov2.(*core.ListObject))
						vm.stack[vm.stackTop] = core.MakeObjectValue(lo, false)
						vm.stackTop++
						continue
					}
					vm.RunTimeError("Concatenation type mismatch: %s + %s", v1.String(), v2.String())
					goto End
				}
			}
			vm.RunTimeError("Invalid operands for concatenation: %s + %s", v1.String(), v2.String())
			goto End

		case core.OP_SUBTRACT:
			// Pop two values from stack, subtract second from first, push result

			if !vm.binarySubtract() {
				goto End
			}

		case core.OP_MULTIPLY:
			// Pop two values from stack, multiply them (handles numbers, vectors, string repetition), push result

			if !vm.binaryMultiply() {
				goto End
			}

		case core.OP_MODULUS:
			// Pop two values from stack, compute modulus (first % second), push result

			if !vm.binaryModulus() {
				goto End
			}

		case core.OP_DIVIDE:
			// Pop two values from stack, divide first by second, push result

			if !vm.binaryDivide() {
				goto End
			}

		case core.OP_DUP:
			// Duplicate the value at the top of the stack

			vm.stack[vm.stackTop] = vm.stack[vm.stackTop-1]
			vm.stackTop++

		case core.OP_NIL:
			// Push the nil value onto the stack

			vm.stack[vm.stackTop] = core.NIL_VALUE
			vm.stackTop++

		case core.OP_TRUE:
			// Push the boolean true value onto the stack

			vm.stack[vm.stackTop] = core.MakeBooleanValue(true, false)
			vm.stackTop++

		case core.OP_FALSE:
			// Push the boolean false value onto the stack

			vm.stack[vm.stackTop] = core.MakeBooleanValue(false, false)
			vm.stackTop++

		case core.OP_NOT:
			// Pop value from stack, apply logical NOT, push boolean result

			v := vm.pop()
			bv := vm.isFalsey(v)
			vm.stack[vm.stackTop] = core.MakeBooleanValue(bv, false)
			vm.stackTop++

		case core.OP_LOOP:
			// Jump backward by offset amount (used for loop constructs)

			offset := vm.readShort()
			frame.Ip -= int(offset)

		case core.OP_INVOKE:
			// Optimized method call: directly invoke method by name with argument count
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			method := constants[idx]
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			if !vm.invoke(method, int(argCount)) {
				goto End
			}

		case core.OP_CLOSURE:
			// Create closure from function constant, capturing upvalues as specified

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
			// Return from current function with value from stack top, unwinding call frame

			result := vm.pop()
			vm.closeUpvalues(frame.Slots)
			vm.frameCount--
			if vm.DebugHook != nil {
				vm.DebugHook(vm, core.DebugEventReturn, result)
			}
			core.LogFmtLn(core.DEBUG, "vm.FrameCount: %d, startFrame: %d", vm.frameCount, startFrame)
			if mode == RUN_CURRENT_FUNCTION && vm.frameCount+1 == startFrame {
				vm.stackTop = frame.Slots
				vm.stack[vm.stackTop] = result
				vm.stackTop++
				core.Log(core.DEBUG, "run return")
				return INTERPRET_OK, result
			}
			if vm.frameCount == 0 {
				vm.pop() // drop main script function obj
				runtime.GC()
				return INTERPRET_OK, result
			}
			vm.stackTop = frame.Slots
			vm.stack[vm.stackTop] = result
			vm.stackTop++

		case core.OP_METHOD:
			// Define method on a class using name from constants
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := constants[idx]
			vm.defineMethod(name.InternedId, false)

		case core.OP_STATIC_METHOD:
			// Define static method on a class using name from constants
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := constants[idx]
			vm.defineMethod(name.InternedId, true)

		case core.OP_NEGATE:
			// Pop numeric value from stack, negate it, push result (handles int and float)

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
			// Get property/field from object using name from constants (handles various object types)

			v := vm.Peek(0)
			if v.Type != core.VAL_OBJ && v.Type != core.VAL_VEC2 && v.Type != core.VAL_VEC3 && v.Type != core.VAL_VEC4 {
				vm.RunTimeError("Attempt to access property of non-object.")
				goto End
			}

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			nv := constants[idx]
			stringId := nv.InternedId

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
			case core.OBJECT_NATIVE:
				// built-in objects can have constants, so check for that
				bobj, ok := v.Obj.(core.HasConstants)
				if ok {
					val := bobj.GetConstant(stringId)
					vm.pop() // pop the object
					vm.stack[vm.stackTop] = val
					vm.stackTop++
					continue
				} else {
					name := core.GetStringValue(nv)
					vm.RunTimeError("Get property '%s' not found.", name)
					goto End
				}

			case core.OBJECT_MODULE:
				ot := v.AsModule()

				if val, ok := ot.Environment.GetVar(nv.InternedId); ok {
					vm.pop()
					vm.stack[vm.stackTop] = val
					vm.stackTop++
				} else {
					name := core.GetStringValue(nv)
					vm.RunTimeError("Get property '%s' not found.", name)
					goto End
				}

			default:
				name := core.GetStringValue(nv)
				vm.RunTimeError("Get property : '%s' not found.", name)
				goto End
			}

		// stack top is value, byte operand is the index of the property name in constants,
		// stack + 1 is the object to set the property on.
		case core.OP_SET_PROPERTY:

			val := vm.Peek(0)
			v := vm.Peek(1)
			if v.Type != core.VAL_OBJ && v.Type != core.VAL_VEC2 && v.Type != core.VAL_VEC3 && v.Type != core.VAL_VEC4 {
				vm.RunTimeError("Set property : not found.")
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
					vm.RunTimeError("Set property : '%s' not found.", core.GetStringValue(constants[idx]))
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
					vm.RunTimeError("Set property : '%s' not found.", core.GetStringValue(constants[idx]))
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
					vm.RunTimeError("Set property : '%s' not found.", core.GetStringValue(constants[idx]))
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
				vm.RunTimeError("Set property : '%s' not found.", core.GetStringValue(constants[idx]))
				goto End
			}

		// entered a try block, IP of the except block is encoded in the next 2 instructions
		// push an exception handler storing that info
		case core.OP_TRY:
			// Begin try block: push exception handler with address of except block
			exceptIP := vm.readShort()
			frame.Handlers = &core.ExceptionHandler{
				ExceptIP: exceptIP,
				StackTop: vm.stackTop,
				Prev:     frame.Handlers,
			}

		// ended a try block OK, so pop the handler block
		case core.OP_END_TRY:
			// End try block successfully: remove exception handler from stack
			frame.Handlers = frame.Handlers.Prev

		// marks the start of an exception handler block.  index of exception classname is in next instruction
		case core.OP_EXCEPT:
			// Begin except block: exception handler start marker
			frame.Ip++

		// marks the end of an exception handler block
		case core.OP_END_EXCEPT:
			// End except block: exception handler end marker

		// 1 pop the thrown exception instance from the stack
		// 2 get the top frame exception handler - this has the IP of the first handler core.OP_EXCEPT.
		//   next instruction is an index to the exception classname in constants.
		//   if the thrown exception name matches the handler, run the handler
		//   else skip to the next handler if it exists, or unwind the call stack and retry.
		//   we'll either hit a matching hander or exit the vm with an unhandled exception error.
		case core.OP_RAISE:
			// Raise/throw an exception: pop exception object and start exception handling
			err := vm.pop()
			if !vm.raiseException(err) {
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}

		case core.OP_CLASS:
			// Create new class object using name from constants and push onto stack
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			name := core.GetStringValue(constants[idx])
			vm.stack[vm.stackTop] = core.MakeObjectValue(core.MakeClassObject(name), false)
			vm.stackTop++

		case core.OP_INHERIT:
			// Set up class inheritance: subclass inherits methods from superclass
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
			// Get method from superclass and bind it to current instance
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
			// Optimized super method call: invoke superclass method directly
			idx := vm.currCode[frame.Ip]
			frame.Ip++
			method := constants[idx]
			argCount := vm.currCode[frame.Ip]
			frame.Ip++
			superclass := vm.pop().AsClass()
			if !vm.invokeFromClass(superclass, method, int(argCount), false) {
				return INTERPRET_RUNTIME_ERROR, core.NIL_VALUE
			}

		case core.OP_IMPORT:
			// Import module: load and register module by name with optional alias

			idx := vm.currCode[frame.Ip]
			frame.Ip++
			mv := constants[idx]
			module := mv.AsString().Get()

			idx = vm.currCode[frame.Ip]
			frame.Ip++
			alv := constants[idx]
			alias := alv.AsString().Get()

			sID := core.InternName(module)
			// check if module is in builtins
			moduleObj, ok := vm.BuiltInModules[sID]
			if ok {
				// copy built-in module to the current environment
				vm.frame().Closure.Function.Environment.SetVar(sID, core.MakeObjectValue(moduleObj, false))
				continue
			}

			status := vm.importModule(module, alias)
			if status != INTERPRET_OK {
				core.LogFmtLn(core.ERROR, "Failed to import module '%s' as '%s'.\n", module, alias)
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
				vm.RunTimeError("Failed to import module '%s'.", module)
				return status, core.NIL_VALUE
			}

			if length == 0 {
				if !vm.importFunctionFromModule(module, "__all__") {
					vm.RunTimeError("Failed to import '%s' from module '%s'.", "__all__", module)
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
			// Convert stack top value to string representation (handles class toString methods)

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
			// Create list object from values on stack: pop item count, create list with those items
			// item count is operand, expects items on stack,  list object will be stack top
			vm.createList(frame)

		case core.OP_CREATE_TUPLE:
			// Create tuple object from values on stack: pop item count, create immutable tuple
			// item count is operand, expects items on stack,  list object will be stack top
			vm.createTuple(frame)

		case core.OP_CREATE_DICT:
			// Create dictionary object from key-value pairs on stack
			// key/pair item count is operand, expects keys/values on stack,  dict object will be stack top
			vm.createDict(frame)

		case core.OP_INDEX:
			// Index into list/string/dict: pop index and container, push element at index
			// list/map + index on stack,  item at index -> stack top
			if !vm.index() {
				goto End
			}

		case core.OP_INDEX_ASSIGN:
			// Assign to index in list/dict: pop value, index, and container, update in place
			// list + index + RHS on stack,  list updated in place
			if !vm.indexAssign() {
				goto End
			}

		case core.OP_SLICE:
			// Create slice of list/string: pop from/to indices and container, push new slice
			// list + from/to on stack. nil indicates from start/end.  new list at index -> stack top
			if !vm.slice() {
				goto End
			}
		case core.OP_SLICE_ASSIGN:
			// Assign slice to list: pop slice, from/to indices, and list, update in place
			// list + from/to + RHS on stack.  list updated in place
			if !vm.sliceAssign() {
				goto End
			}

		// ### foreach ( var a in iterable ) ###
		// local slot, iterator slot, end of foreach in next 3 instructions
		// can handle native iterable objects (list, string) or lox class instances
		// with __iter__ method returning an iterator object implementing __next__ method

		case core.OP_FOREACH:
			// Begin foreach loop: set up iteration over iterable object (list, string, or custom iterator)
			slot := vm.readByte()
			iterableSlot := vm.readByte()
			jumpToEnd := vm.readShort()
			iterable := vm.stack[frame.Slots+int(iterableSlot)]
			if iterable.Type != core.VAL_OBJ {
				vm.RunTimeError("Foreach requires an iterable object, got %s", iterable.String())
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
				// we need to call it to get an iterator object which has a __next__ method
				// so we can iterate over it.
				_, ok := iterable.AsInstance().Class.Methods[core.ITER]
				if ok {
					vm.stack[vm.stackTop] = iterable // push the iterator object onto the stack for call
					vm.stackTop++

					// Assert: stack should have the iterable object at the top
					expectedStackTop := vm.stackTop
					if vm.stackTop == 0 || !core.ValuesEqual(vm.stack[vm.stackTop-1], iterable, false) {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Expected iterable at stack top before iter call. stackTop=%d", vm.stackTop)
					}

					vm.invoke(ITER_METHOD, 0)
					iok, result := vm.run(RUN_CURRENT_FUNCTION)

					// Assert: stack top should be same after vm.run since invoke/run should manage stack properly
					if vm.stackTop != expectedStackTop {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top changed unexpectedly after iter vm.run. Expected=%d, Actual=%d", expectedStackTop, vm.stackTop)
					}

					if iok != INTERPRET_OK {
						goto End
					}
					frame = vm.frame()
					core.Log(core.DEBUG, "iter pop")
					vm.pop()

					// Assert: stack top should be back to original level after pop
					if vm.stackTop != expectedStackTop-1 {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top incorrect after iter pop. Expected=%d, Actual=%d", expectedStackTop-1, vm.stackTop)
					}
					if !result.IsInstanceObject() {
						vm.RunTimeError("Foreach iterator must be a object with a __next__ method.")
						goto End
					}

					_, ok := result.AsInstance().Class.Methods[core.NEXT]
					if !ok {
						vm.RunTimeError("Foreach iterator must have a __next__ method.")
						goto End
					}
					core.LogFmtLn(core.DEBUG, "set iterator object in slot %d", frame.Slots+int(iterableSlot))
					vm.stack[frame.Slots+int(iterableSlot)] = result // store iterator object in stack				vm.stack[vm.stackTop] = result                   // push the iterator object onto the stack for call
					vm.stackTop++

					// Assert: stack should have the result (iterator) object at the top
					expectedStackTop2 := vm.stackTop
					if vm.stackTop == 0 || !core.ValuesEqual(vm.stack[vm.stackTop-1], result, false) {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Expected result at stack top before next call. stackTop=%d", vm.stackTop)
					}

					vm.invoke(NEXT_METHOD, 0)
					iok, result = vm.run(RUN_CURRENT_FUNCTION)

					// Assert: stack top should be same after vm.run since invoke/run should manage stack properly
					if vm.stackTop != expectedStackTop2 {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top changed unexpectedly after next vm.run. Expected=%d, Actual=%d", expectedStackTop2, vm.stackTop)
					}

					if iok != INTERPRET_OK {
						goto End
					}
					frame = vm.frame()
					core.Log(core.DEBUG, "next pop in foreach")
					vm.pop()

					// Assert: stack top should be back to original level after pop
					if vm.stackTop != expectedStackTop2-1 {
						core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top incorrect after next pop. Expected=%d, Actual=%d", expectedStackTop2-1, vm.stackTop)
					}
					if result.Type == core.VAL_NIL {
						// we have no items, so jump to end of foreach loop
						frame.Ip += int(jumpToEnd - 2)
					} else {
						core.LogFmtLn(core.DEBUG, "set result in local slot %d", frame.Slots+int(slot))
						vm.stack[frame.Slots+int(slot)] = result // set result in the local slot
					}

					continue
				}
			} else {
				vm.RunTimeError("Foreach requires an iterable object.")
				goto End
			}
		case core.OP_NEXT:
			// Continue foreach loop: get next item from iterator, jump back if more items available

			jumpToStart := vm.readShort()
			iterSlot := frame.Slots + int(vm.readByte())
			iterVal := vm.stack[iterSlot]
			if iterVal.Obj.GetType() != core.OBJECT_INSTANCE {
				val := iterVal.AsIterator().Next()
				if val.Type != core.VAL_NIL {
					vm.stack[iterSlot-1] = val
					frame.Ip -= int(jumpToStart + 1)
				}
			} else {
				vm.stack[vm.stackTop] = iterVal
				vm.stackTop++ // push the iterator object onto the stack for call

				// Assert: stack should have the iterVal object at the top
				expectedStackTop3 := vm.stackTop
				if vm.stackTop == 0 || !core.ValuesEqual(vm.stack[vm.stackTop-1], iterVal, false) {
					core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Expected iterVal at stack top before next call. stackTop=%d", vm.stackTop)
				}

				vm.invoke(NEXT_METHOD, 0)
				ok, rv := vm.run(RUN_CURRENT_FUNCTION)

				// Assert: stack top should be same after vm.run since invoke/run should manage stack properly
				if vm.stackTop != expectedStackTop3 {
					core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top changed unexpectedly after next vm.run (OP_NEXT). Expected=%d, Actual=%d", expectedStackTop3, vm.stackTop)
				}

				if ok != INTERPRET_OK {
					goto End
				}
				core.Log(core.DEBUG, "next pop")
				vm.pop()

				// Assert: stack top should be back to original level after pop
				if vm.stackTop != expectedStackTop3-1 {
					core.LogFmtLn(core.ERROR, "ASSERTION FAILED: Stack top incorrect after next pop (OP_NEXT). Expected=%d, Actual=%d", expectedStackTop3-1, vm.stackTop)
				}

				frame = vm.frame()
				if rv.Type != core.VAL_NIL {
					vm.stack[iterSlot-1] = rv
					frame.Ip -= int(jumpToStart + 1)
				}
				//vm.pop()
			}

		case core.OP_END_FOREACH:
			// End foreach loop marker (no operation needed)

		// stack 1 : string or list
		// stack 2 : key or substring

		case core.OP_IN:
			// Check membership: test if value is in string/list, push boolean result

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
			// Debug breakpoint: pause execution for debugging

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
			// Unknown/invalid opcode encountered
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

//------------------------------------------------------------------------------------------

// callValue attempts to call a value with the specified number of arguments.
// Handles closures, built-in functions, classes (constructors), and bound methods.
func (vm *VM) callValue(callee core.Value, argCount int) bool {

	//core.LogFmtLn(core.DEBUG, "Calling value %s with %d args", callee.String(), argCount)

	if callee.Type == core.VAL_OBJ {
		if callee.IsClosureObject() {
			//core.LogFmtLn(core.DEBUG, "Calling closure %s with %d args", callee.Obj.String(), argCount)
			return vm.call(core.GetClosureObjectValue(callee), argCount)

		} else if callee.IsBuiltInObject() {
			//core.LogFmtLn(core.DEBUG, "Calling built-in function %s with %d args", callee.Obj.String(), argCount)
			nf := callee.AsBuiltIn()
			res := nf.Function(argCount, vm.stackTop-argCount, vm)
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
	core.LogFmtLn(core.DEBUG, "Cannot call value %s", callee.String())
	vm.RunTimeError("Can only call functions and classes.")
	return false
}

//------------------------------------------------------------------------------------------

// invoke performs optimized method calls and module access without separate property lookup.
// optimised method call/module access
func (vm *VM) invoke(name core.Value, argCount int) bool {
	receiver := vm.Peek(argCount)

	if receiver.Type == core.VAL_VEC2 ||
		receiver.Type == core.VAL_VEC3 ||
		receiver.Type == core.VAL_VEC4 {
		return vm.VectorMethodCall(receiver, name, argCount)
	}

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
	case core.OBJECT_NATIVE, core.OBJECT_LIST, core.OBJECT_DICT, core.OBJECT_STRING:
		return vm.invokeFromBuiltin(receiver.Obj, name, argCount)
	default:
		vm.RunTimeError("Invalid use of '.' operator")
		return false
	}

}

//------------------------------------------------------------------------------------------

// invokeFromClass calls a method from a specific class, handling both static and instance methods.
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

//------------------------------------------------------------------------------------------

// invokeFromModule calls a function from a loaded module by name.
func (vm *VM) invokeFromModule(module *core.ModuleObject, name core.Value, argCount int) bool {

	fn, ok := module.Environment.GetVar(name.InternedId)
	if !ok {
		n := core.GetStringValue(name)
		vm.RunTimeError("Undefined module property '%s'.", n)
		return false
	}
	return vm.callValue(fn, argCount)
}

//------------------------------------------------------------------------------------------

// invokeFromBuiltin calls a method on a built-in object (native Go object with exposed methods).
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

//------------------------------------------------------------------------------------------

// VectorMethodCall handles method calls on vector types (Vec2, Vec3, Vec4) with optimized operations.
func (vm *VM) VectorMethodCall(receiver core.Value, name core.Value, argCount int) bool {
	switch receiver.Type {
	case core.VAL_VEC2:
		if name.InternedId == core.ADD && argCount == 1 {
			// special case for Vec2 addition
			other := vm.Peek(0)
			if other.Obj.GetType() == core.OBJECT_VEC2 {
				v2 := other.AsVec2()
				receiver.AsVec2().AddInPlace(v2)
				vm.pop() // pop the other vector
				return true
			}
		}
	case core.VAL_VEC3:
		if name.InternedId == core.ADD && argCount == 1 {
			// special case for Vec3 addition
			other := vm.Peek(0)
			if other.Obj.GetType() == core.OBJECT_VEC3 {
				v3 := other.AsVec3()
				receiver.AsVec3().AddInPlace(v3)
				vm.pop() // pop the other vector
				return true
			}
		}
	case core.VAL_VEC4:
		if name.InternedId == core.ADD && argCount == 1 {
			// special case for Vec4 addition
			other := vm.Peek(0)
			if other.Obj.GetType() == core.OBJECT_VEC4 {
				v4 := other.AsVec4()
				receiver.AsVec4().AddInPlace(v4)
				vm.pop() // pop the other vector
				return true
			}
		}
	}

	vm.RunTimeError("Invalid use of '.' operator")
	return false
}

// bindMethod creates a bound method object that combines an instance with a method from its class.
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

//------------------------------------------------------------------------------------------

// captureUpvalue creates or finds an upvalue for a local variable at the specified stack slot.
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

//------------------------------------------------------------------------------------------

// closeUpvalues closes all open upvalues at or above the specified stack position.
func (vm *VM) closeUpvalues(last int) {
	for vm.openUpValues != nil && vm.openUpValues.Slot >= last {
		upvalue := vm.openUpValues
		upvalue.Closed = vm.stack[upvalue.Slot]
		upvalue.Location = &upvalue.Closed
		vm.openUpValues = upvalue.Next
	}
}

//------------------------------------------------------------------------------------------

// defineMethod adds a method to a class, handling both static and instance methods.
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

//------------------------------------------------------------------------------------------

// call sets up a new call frame for executing a closure with the specified argument count.
func (vm *VM) call(closure *core.ClosureObject, argCount int) bool {
	if vm.DebugHook != nil {
		vm.DebugHook(vm, core.DebugEventCall, closure)
	}

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
	vm.Frames[vm.frameCount] = frame
	vm.frameCount++
	if vm.frameCount == FRAMES_MAX {
		vm.RunTimeError("Stack overflow.")
		return false
	}

	return true
}

//------------------------------------------------------------------------------------------

// readShort reads a 16-bit value from the current instruction stream (big-endian format).
func (vm *VM) readShort() uint16 {

	vm.frame().Ip += 2
	b1 := uint16(vm.currCode[vm.frame().Ip-2])
	b2 := uint16(vm.currCode[vm.frame().Ip-1])
	return uint16(b1<<8 | b2)
}

//------------------------------------------------------------------------------------------

// readByte reads a single byte from the current instruction stream and advances the instruction pointer.
func (vm *VM) readByte() uint8 {

	frame := vm.Frames[vm.frameCount-1]
	frame.Ip += 1
	return vm.currCode[frame.Ip-1]
}

//------------------------------------------------------------------------------------------

// isFalsey determines if a value should be considered false in a boolean context.
// Only nil and false are falsy in Lox.
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

//------------------------------------------------------------------------------------------

// natively raise an exception given a name:
// - get the class object for the name from globals
// - make an instance of the class and set the message on it
// - pass the instance to raiseException
// used for vm raising errors that can be handled in lox e.g EOFError when reading a file
// RaiseExceptionByName creates and raises an exception with the specified name and message.
func (vm *VM) RaiseExceptionByName(name string, msg string) bool {

	classVal := vm.BuiltIns[core.InternName(name)]
	classObj := classVal.Obj
	instance := core.MakeInstanceObject(classObj.(*core.ClassObject))
	instance.Fields[core.MSG] = core.MakeStringObjectValue(msg, false)
	return vm.raiseException(core.MakeObjectValue(instance, false))
}

//------------------------------------------------------------------------------------------

// raiseException handles exception propagation through the call stack and exception handlers.
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
					v, ok = vm.BuiltIns[id]
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

//------------------------------------------------------------------------------------------

// nextHandler moves to the next exception handler in the current frame.
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

//------------------------------------------------------------------------------------------

// popFrame removes the current call frame and continues exception handling in the previous frame.
func (vm *VM) popFrame() bool {
	if vm.frameCount == 1 {
		return false
	}
	vm.frameCount--
	vm.stackTop = vm.Frames[vm.frameCount].Slots
	return true
}

//------------------------------------------------------------------------------------------

// appendStackTrace adds the current function call information to the stack trace.
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
	line := function.Chunk.Lines[frame.Ip-1]

	s := fmt.Sprintf("File '%s' , line %d, in %s ", script, line, where)
	vm.stackTrace = append(vm.stackTrace, s)
	codeline := vm.sourceLine(script, line)
	vm.stackTrace = append(vm.stackTrace, codeline)
}

//------------------------------------------------------------------------------------------

// PrintStackTrace outputs the current stack trace to stderr for debugging.
func (vm *VM) PrintStackTrace() {
	for _, v := range vm.stackTrace {
		fmt.Fprintf(os.Stderr, "%s\n", v)
	}
}

//------------------------------------------------------------------------------------------

// sourceLine extracts a specific line from the source code for error reporting.
func (vm *VM) sourceLine(script string, line int) string {

	source := vm.source
	if script != vm.script {
		module := getModule(script)
		source = globalModuleSource[module]
	}
	lines := strings.Split(source, "\n")
	if line > 0 && line <= len(lines) {
		rv := lines[line-1]
		return rv
	}
	return ""
}

//------------------------------------------------------------------------------------------

// importModule loads and executes a Lox module, adding it to the current environment.
func (vm *VM) importModule(moduleName string, alias string) InterpretResult {

	core.LogFmtLn(core.DEBUG, "Importing module %s as %s\n", moduleName, alias)
	searchPath := getPath(vm.Args(), moduleName) + ".lox"
	bytes, err := os.ReadFile(searchPath)
	if err != nil {
		fmt.Printf("Could not find module %s.", searchPath)
		os.Exit(1)
	}
	module, ok := globalModules[moduleName]
	if ok {
		// module already loaded, just add to the current environment
		core.LogFmtLn(core.DEBUG, "Module %s already loaded, adding to current environment.\n", moduleName)
		v := core.MakeObjectValue(module, false)
		vm.frame().Closure.Function.Environment.SetVar(core.InternName(alias), v)
		return INTERPRET_OK
	}
	globalModuleSource[moduleName] = string(bytes)
	subvm := NewVM(searchPath, false)
	subvm.BuiltIns = vm.BuiltIns
	subvm.BuiltInModules = vm.BuiltInModules
	subvm.SetArgs(vm.Args())
	subvm.ModuleImport = true
	// see if we can load lxc bytecode file for the module.
	// if so run it
	if loadedChunk, newEnv, ok := loadLxc(searchPath); ok {
		core.LogFmtLn(core.DEBUG, "Loaded module %s from bytecode.\n", moduleName)
		loadedChunk.Filename = moduleName
		subvm.callLoadedChunk(moduleName, newEnv, loadedChunk)
		core.LogFmtLn(core.DEBUG, "Completed run of module %s.\n", moduleName)
	} else {
		// if not, load the module source, compile it and run it
		core.LogFmtLn(core.DEBUG, "Compiling module %s from source.\n", moduleName)
		res, _ := subvm.Interpret(string(bytes), moduleName)
		if res != INTERPRET_OK {
			return res
		}
		core.LogFmtLn(core.DEBUG, "Completed compile/run of module %s.\n", moduleName)
	}
	core.LogFmtLn(core.DEBUG, "Created module object for %s.\n", moduleName)
	subfn := subvm.Frames[0].Closure.Function
	mo := core.MakeModuleObject(moduleName, *subfn.Environment)

	globalModules[moduleName] = mo
	v := core.MakeObjectValue(mo, false)
	debug.TraceDumpValue("Dump:", v)
	vm.frame().Closure.Function.Environment.SetVar(core.InternName(alias), v)
	core.LogFmtLn(core.DEBUG, "ImportModule %s as %s return\n", moduleName, alias)
	return INTERPRET_OK
}

//------------------------------------------------------------------------------------------

// callLoadedChunk executes a compiled chunk in a new environment with module isolation.
func (subvm *VM) callLoadedChunk(name string, newEnv *core.Environment, chunk *core.Chunk) {

	function := core.MakeFunctionObject(name, newEnv)
	function.Chunk = chunk
	function.Name = core.MakeStringObject(name)
	closure := core.MakeClosureObject(function)
	subvm.push(core.MakeObjectValue(closure, false))
	subvm.call(closure, 0)
	_, _ = subvm.run(RUN_TO_COMPLETION)
}

//------------------------------------------------------------------------------------------

// importFunctionFromModule imports a specific function from a module into the current environment.
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
		t := fn.Obj.GetType()
		if t != core.OBJECT_CLOSURE && t != core.OBJECT_CLASS {
			vm.RunTimeError("'%s' not found in module '%s'.", name, module)
			return false
		}

		vm.frame().Closure.Function.Environment.SetVar(nameId, fn)
		return true
	}

}

//------------------------------------------------------------------------------------------

// createList creates a list object from the specified number of values on the stack.
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

//------------------------------------------------------------------------------------------

// createTuple creates an immutable tuple object from the specified number of values on the stack.
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

//------------------------------------------------------------------------------------------

// createDict creates a dictionary object from key-value pairs on the stack.
func (vm *VM) createDict(frame *core.CallFrame) {

	itemCount := int(vm.currCode[frame.Ip])
	frame.Ip++
	dict := map[int]core.Value{}

	for i := 0; i < itemCount; i++ {
		value := vm.pop()
		key := vm.pop()

		var keyStr string
		if key.IsStringObject() {
			keyStr = key.AsString().Get()
		} else {
			keyStr = key.String()
		}

		dict[core.InternName(keyStr)] = value
	}
	do := core.MakeDictObject(dict)
	vm.stack[vm.stackTop] = core.MakeObjectValue(do, false)
	vm.stackTop++
}

//------------------------------------------------------------------------------------------

// index performs indexing operation on lists, strings, and dictionaries.
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

			var key string
			if iv.IsStringObject() {
				key = iv.AsString().Get()
			} else {
				key = iv.String()
			}

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

//------------------------------------------------------------------------------------------

// indexAssign performs assignment to an index in lists and dictionaries.
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

			var key string
			if index.IsStringObject() {
				key = index.AsString().Get()
			} else {
				key = index.String()
			}

			t.Set(key, rhs)
			return true
		}
	}
	vm.RunTimeError("Can only assign to collection.")
	return false
}

//------------------------------------------------------------------------------------------

// slice creates a new list/string from a slice of an existing list/string.
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

//------------------------------------------------------------------------------------------

// sliceAssign assigns a slice of values to a range in a list.
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

//------------------------------------------------------------------------------------------

// binarySubtract performs subtraction operation on numeric values and vectors.
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

	case core.VAL_VEC2:
		if v1.Type != core.VAL_VEC2 {
			vm.RunTimeError("Subtraction type mismatch: %s - %s", v1.String(), v2.String())
			return false
		}
		vec1 := v1.AsVec2()
		vec2 := v2.AsVec2()
		vm.stack[vm.stackTop] = core.MakeVec2Value(vec1.X-vec2.X, vec1.Y-vec2.Y, false)
		vm.stackTop++
		return true

	case core.VAL_VEC3:
		if v1.Type != core.VAL_VEC3 {
			vm.RunTimeError("Subtraction type mismatch: %s - %s", v1.String(), v2.String())
			return false
		}
		vec1 := v1.AsVec3()
		vec2 := v2.AsVec3()
		vm.stack[vm.stackTop] = core.MakeVec3Value(vec1.X-vec2.X, vec1.Y-vec2.Y, vec1.Z-vec2.Z, false)
		vm.stackTop++
		return true

	case core.VAL_VEC4:
		if v1.Type != core.VAL_VEC4 {
			vm.RunTimeError("Subtraction type mismatch: %s - %s", v1.String(), v2.String())
			return false
		}
		vec1 := v1.AsVec4()
		vec2 := v2.AsVec4()
		vm.stack[vm.stackTop] = core.MakeVec4Value(vec1.X-vec2.X, vec1.Y-vec2.Y, vec1.Z-vec2.Z, vec1.W-vec2.W, false)
		vm.stackTop++
		return true
	}

	vm.RunTimeError("Subtraction type mismatch: %s - %s", v1.String(), v2.String())
	return false
}

//------------------------------------------------------------------------------------------

// binaryMultiply performs multiplication operation on numbers, vectors, and string repetition.
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

//------------------------------------------------------------------------------------------

// binaryDivide performs division operation on numeric values and vectors.
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

	vm.RunTimeError("Division type mismatch  %s / %s", v1.String(), v2.String())
	return false
}

//------------------------------------------------------------------------------------------

// binaryModulus performs modulus operation on integer values.
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

//------------------------------------------------------------------------------------------

// stringMultiply creates a new string by repeating the input string x times.
func (vm *VM) stringMultiply(s string, x int) core.Value {

	rv := ""
	for i := 0; i < x; i++ {
		rv += s
	}
	return core.MakeStringObjectValue(rv, false)
}

// ------------------------------------------------------------------------------------------
// pauseExecution handles breakpoint debugging by pausing VM execution.
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

// ------------------------------------------------------------------------------------------
// patchInstruction replaces an instruction at the specified instruction pointer with a new operation code.
// used to specialise optimised addition to int or float addition
func (vm *VM) patchInstruction(ip int, newOp byte) {
	vm.currCode[ip] = newOp
}

//------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------

// return the path to the given module.
// first, will look in lox/modules in the lox installation directory defined in LOX_PATH environment var.
// if not found will look in the directory containing the main module being run
// getPath constructs the full file path for a module, handling both absolute and relative paths.
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

//------------------------------------------------------------------------------------------

// getModule extracts the module name from a file path by removing the directory and extension.
func getModule(fullPath string) string {
	base := filepath.Base(fullPath)      // "foo.lox"
	ext := filepath.Ext(base)            // ".lox"
	return strings.TrimSuffix(base, ext) // "foo"
}

//------------------------------------------------------------------------------------------
