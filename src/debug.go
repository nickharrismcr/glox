package lox

import (
	"fmt"
	"glox/src/core"
)

var DebugSuppress = false
var DebugTraceExecution = false
var DebugPrintCode = false
var DebugShowGlobals = false
var DebugSkipBuiltins = false

func Debug(s string) {
	if DebugTraceExecution {
		println(s)
	}
}

func Debugf(format string, args ...interface{}) {
	if DebugTraceExecution {
		fmt.Printf(format+"\n", args...)
	}
}

func disassemble(c *core.Chunk, name string) {

	fmt.Printf("=== %s ===\n", name)
	s := ""
	for _, v := range c.Constants {
		s = s + fmt.Sprintf("[ %s ]", v.String())
	}
	fmt.Println(s)
	offset := 0
	for {
		instr := c.Code[offset]
		offset = disassembleInstruction(c, name, nil, instr, offset)
		if offset >= len(c.Code) {
			break
		}
	}
}

var lastoffset int = 0

func disassembleInstruction(c *core.Chunk, name string, frame *CallFrame, i uint8, offset int) int {

	if frame != nil {
		if frame.depth > 1 {
			name = frame.closure.Function.Name.Get()
		}

		fmt.Printf("%02d : [%-10s] : ", frame.depth, name)
	}
	fmt.Printf("%04d ", offset)
	if offset > 0 && c.Lines[offset] == lastoffset {
		fmt.Printf("   | ")

	} else {
		fmt.Printf("%04d ", c.Lines[offset])
	}
	lastoffset = c.Lines[offset]

	switch i {
	case core.OP_RETURN:
		return simpleInstruction(c, "OP_RETURN", offset)
	case core.OP_CONSTANT:
		return constantInstruction(c, "OP_CONSTANT", offset)
	case core.OP_NEGATE:
		return simpleInstruction(c, "OP_NEGATE", offset)
	case core.OP_ADD:
		return simpleInstruction(c, "OP_ADD", offset)
	case core.OP_SUBTRACT:
		return simpleInstruction(c, "OP_SUBTRACT", offset)
	case core.OP_MODULUS:
		return simpleInstruction(c, "OP_MODULUS", offset)
	case core.OP_MULTIPLY:
		return simpleInstruction(c, "OP_MULTIPLY", offset)
	case core.OP_DIVIDE:
		return simpleInstruction(c, "OP_DIVIDE", offset)
	case core.OP_NIL:
		return simpleInstruction(c, "OP_NIL", offset)
	case core.OP_TRUE:
		return simpleInstruction(c, "OP_TRUE", offset)
	case core.OP_FALSE:
		return simpleInstruction(c, "OP_FALSE", offset)
	case core.OP_NOT:
		return simpleInstruction(c, "OP_NOT", offset)
	case core.OP_EQUAL:
		return simpleInstruction(c, "OP_EQUAL", offset)
	case core.OP_GREATER:
		return simpleInstruction(c, "OP_GREATER", offset)
	case core.OP_LESS:
		return simpleInstruction(c, "OP_LESS", offset)
	case core.OP_PRINT:
		return simpleInstruction(c, "OP_PRINT", offset)
	case core.OP_STR:
		return simpleInstruction(c, "OP_STR", offset)
	case core.OP_POP:
		return simpleInstruction(c, "OP_POP", offset)
	case core.OP_DEFINE_GLOBAL:
		return constantInstruction(c, "OP_DEFINE_GLOBAL", offset)
	case core.OP_DEFINE_GLOBAL_CONST:
		return constantInstruction(c, "OP_DEFINE_GLOBAL_CONST", offset)
	case core.OP_GET_GLOBAL:
		return constantInstruction(c, "OP_GET_GLOBAL", offset)
	case core.OP_SET_GLOBAL:
		return constantInstruction(c, "OP_SET_GLOBAL", offset)
	case core.OP_GET_LOCAL:
		return byteInstruction(c, "OP_GET_LOCAL", offset)
	case core.OP_SET_LOCAL:
		return byteInstruction(c, "OP_SET_LOCAL", offset)
	case core.OP_JUMP_IF_FALSE:
		return jumpInstruction(c, "OP_JUMP_IF_FALSE", 1, offset)
	case core.OP_JUMP:
		return jumpInstruction(c, "OP_JUMP", 1, offset)
	case core.OP_LOOP:
		return jumpInstruction(c, "OP_LOOP", -1, offset)
	case core.OP_CALL:
		return byteInstruction(c, "OP_CALL", offset)
	case core.OP_CREATE_LIST:
		return byteInstruction(c, "OP_CREATE_LIST", offset)
	case core.OP_CREATE_DICT:
		return byteInstruction(c, "OP_CREATE_DICT", offset)
	case core.OP_INDEX:
		return simpleInstruction(c, "OP_INDEX", offset)
	case core.OP_INDEX_ASSIGN:
		return simpleInstruction(c, "OP_INDEX_ASSIGN", offset)
	case core.OP_SLICE:
		return simpleInstruction(c, "OP_SLICE", offset)
	case core.OP_SLICE_ASSIGN:
		return simpleInstruction(c, "OP_SLICE_ASSIGN", offset)
	case core.OP_FOREACH:
		return foreachInstruction(c, offset)
	case core.OP_NEXT:
		return nextInstruction(c, "OP_NEXT", -1, offset)
	case core.OP_END_FOREACH:
		return simpleInstruction(c, "OP_END_FOREACH", offset)
	case core.OP_CLOSURE:

		var s string

		offset++
		constant := c.Code[offset]
		offset++
		fmt.Printf("%-16s %04d", "OP_CLOSURE", constant)
		value := c.Constants[constant]
		fmt.Printf("  %s\n", value.String())
		function := core.GetFunctionObjectValue(value)
		for j := 0; j < function.UpvalueCount; j++ {
			isLocal := c.Code[offset]
			offset++
			index := c.Code[offset]
			offset++
			if isLocal == 1 {
				s = "local"
			} else {
				s = "upvalue"
			}
			fmt.Printf("%04d      |                     %s %d\n", offset-2, s, index)
		}
		return offset
	case core.OP_GET_UPVALUE:
		return byteInstruction(c, "OP_GET_UPVALUE", offset)
	case core.OP_SET_UPVALUE:
		return byteInstruction(c, "OP_SET_UPVALUE", offset)
	case core.OP_CLOSE_UPVALUE:
		return simpleInstruction(c, "OP_CLOSE_UPVALUE", offset)
	case core.OP_CLASS:
		return constantInstruction(c, "OP_CLASS", offset)
	case core.OP_GET_PROPERTY:
		return constantInstruction(c, "OP_GET_PROPERTY", offset)
	case core.OP_SET_PROPERTY:
		return constantInstruction(c, "OP_SET_PROPERTY", offset)
	case core.OP_METHOD:
		return constantInstruction(c, "OP_METHOD", offset)
	case core.OP_INVOKE:
		return invokeInstruction(c, "OP_INVOKE", offset)
	case core.OP_INHERIT:
		return simpleInstruction(c, "OP_INHERIT", offset)
	case core.OP_GET_SUPER:
		return constantInstruction(c, "OP_INHERIT", offset)
	case core.OP_SUPER_INVOKE:
		return invokeInstruction(c, "OP_SUPER_INVOKE", offset)
	case core.OP_IMPORT:
		return constantInstruction(c, "OP_IMPORT", offset)
	case core.OP_TRY:
		return addressInstruction(c, "OP_TRY", offset)
	case core.OP_END_TRY:
		return jumpInstruction(c, "OP_END_TRY", 1, offset)
	case core.OP_EXCEPT:
		return constantInstruction(c, "OP_EXCEPT", offset)
	case core.OP_RAISE:
		return simpleInstruction(c, "OP_RAISE", offset)
	case core.OP_END_EXCEPT:
		return simpleInstruction(c, "OP_END_EXCEPT", offset)
	case core.OP_BREAKPOINT:
		return simpleInstruction(c, "OP_BREAKPOINT", offset)
	default:
		fmt.Printf("Unknown opcode %d", i)
		return offset + 1
	}
}

func simpleInstruction(c *core.Chunk, name string, offset int) int {

	fmt.Printf("%s\n", name)
	return offset + 1
}

func constantInstruction(c *core.Chunk, name string, offset int) int {

	constant := c.Code[offset+1]
	fmt.Printf("%-16s %04d", name, constant)
	value := c.Constants[constant]
	fmt.Printf("  %s\n", value.String())
	return offset + 2
}

func byteInstruction(c *core.Chunk, name string, offset int) int {

	slot := c.Code[offset+1]
	fmt.Printf("%-16s %04d\n", name, slot)
	return offset + 2
}

func jumpInstruction(c *core.Chunk, name string, sign int, offset int) int {

	var jump uint16

	jump1 := uint16(c.Code[offset+1])
	jump2 := uint16(c.Code[offset+2])

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	fmt.Printf("%-16s %04d -> %d \n", name, offset, uint16(offset)+3+(uint16(sign)*jump))
	return offset + 3
}
func foreachInstruction(c *core.Chunk, offset int) int {

	var jump uint16
	slot := c.Code[offset+1]
	iterslot := c.Code[offset+2]
	idxslot := c.Code[offset+3]
	jump1 := uint16(c.Code[offset+4])
	jump2 := uint16(c.Code[offset+5])

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	fmt.Printf("%-16s %04d %04d %04d %04d -> %d \n", "OP_FOREACH", slot, iterslot, idxslot, jump, uint16(offset)+4+jump)
	return offset + 6
}

func nextInstruction(c *core.Chunk, name string, sign int, offset int) int {

	var jump uint16

	jump1 := uint16(c.Code[offset+1])
	jump2 := uint16(c.Code[offset+2])
	idx := c.Code[offset+3]

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	fmt.Printf("%-16s %04d %04d -> %d \n", name, idx, offset, uint16(offset)+3+(uint16(sign)*jump))
	return offset + 4
}
func addressInstruction(c *core.Chunk, name string, offset int) int {

	var address uint16

	addr1 := uint16(c.Code[offset+1])
	addr2 := uint16(c.Code[offset+2])

	address = uint16(addr1 << 8)
	address |= uint16(addr2)

	fmt.Printf("%-16s %04d -> %d  \n", name, offset, address)
	return offset + 3
}

func invokeInstruction(c *core.Chunk, name string, offset int) int {
	constant := c.Code[offset+1]
	argCount := c.Code[offset+2]
	fmt.Printf("%-16s (%d args) %4d", name, argCount, constant)
	value := c.Constants[constant]
	fmt.Printf("  %s\n", value.String())
	return offset + 3
}

func (vm *VM) showGlobals() {
	if vm.frame().closure.function.environment == nil {
		fmt.Println("No globals (nil environment)")
		return
	}
	fmt.Printf("globals: %s \n", vm.frame().closure.function.environment.name)
	for k, v := range vm.frame().closure.function.environment.vars {
		fmt.Printf("%s -> %s  \n", k, v)
	}
	//for k, v := range vm.environments.builtins {
	//	fmt.Printf("%s -> %s  \n", k, v)
	//}
}

func (vm *VM) showStack() {

	fmt.Printf("                                                         ")
	for i := 1; i < vm.stackTop; i++ {
		v := vm.stack[i]
		s := v.String()

		im := ""
		if v.Immutable() {
			im = "(c)"
		}
		if i > vm.frame().slots {
			fmt.Printf("%% %s%s %%", s, im)
		} else {
			fmt.Printf("| %s%s |", s, im)
		}
	}
	fmt.Printf("\n")
}
