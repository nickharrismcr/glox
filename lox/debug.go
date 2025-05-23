package lox

import (
	"fmt"
)

var DebugTraceExecution = false
var DebugPrintCode = false
var DebugShowGlobals = false

func (c *Chunk) disassemble(name string) {

	fmt.Printf("=== %s ===\n", name)
	idx := -1
	s := ""
	for {
		idx++
		if idx == len(c.constants) {
			break
		}
		constant := c.constants[idx]
		s = s + fmt.Sprintf("[ %s ]", constant.String())
	}
	fmt.Println(s)
	offset := 0
	for {
		instr := c.code[offset]
		offset = c.disassembleInstruction(instr, offset)
		if offset >= len(c.code) {
			break
		}
	}
}

var lastoffset int = 0

func (c *Chunk) disassembleInstruction(i uint8, offset int) int {

	fmt.Printf("%04d ", offset)
	if offset > 0 && c.lines[offset] == lastoffset {
		fmt.Printf("   | ")

	} else {
		fmt.Printf("%04d ", c.lines[offset])
	}
	lastoffset = c.lines[offset]

	switch i {
	case OP_RETURN:
		return c.simpleInstruction("OP_RETURN", offset)
	case OP_CONSTANT:
		return c.constantInstruction("OP_CONSTANT", offset)
	case OP_NEGATE:
		return c.simpleInstruction("OP_NEGATE", offset)
	case OP_ADD:
		return c.simpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT:
		return c.simpleInstruction("OP_SUBTRACT", offset)
	case OP_MODULUS:
		return c.simpleInstruction("OP_MODULUS", offset)
	case OP_MULTIPLY:
		return c.simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE:
		return c.simpleInstruction("OP_DIVIDE", offset)
	case OP_NIL:
		return c.simpleInstruction("OP_NIL", offset)
	case OP_TRUE:
		return c.simpleInstruction("OP_TRUE", offset)
	case OP_FALSE:
		return c.simpleInstruction("OP_FALSE", offset)
	case OP_NOT:
		return c.simpleInstruction("OP_NOT", offset)
	case OP_EQUAL:
		return c.simpleInstruction("OP_EQUAL", offset)
	case OP_GREATER:
		return c.simpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return c.simpleInstruction("OP_LESS", offset)
	case OP_PRINT:
		return c.simpleInstruction("OP_PRINT", offset)
	case OP_STR:
		return c.simpleInstruction("OP_STR", offset)
	case OP_POP:
		return c.simpleInstruction("OP_POP", offset)
	case OP_DEFINE_GLOBAL:
		return c.constantInstruction("OP_DEFINE_GLOBAL", offset)
	case OP_DEFINE_GLOBAL_CONST:
		return c.constantInstruction("OP_DEFINE_GLOBAL_CONST", offset)
	case OP_GET_GLOBAL:
		return c.constantInstruction("OP_GET_GLOBAL", offset)
	case OP_SET_GLOBAL:
		return c.constantInstruction("OP_SET_GLOBAL", offset)
	case OP_GET_LOCAL:
		return c.byteInstruction("OP_GET_LOCAL", offset)
	case OP_SET_LOCAL:
		return c.byteInstruction("OP_SET_LOCAL", offset)
	case OP_JUMP_IF_FALSE:
		return c.jumpInstruction("OP_JUMP_IF_FALSE", 1, offset)
	case OP_JUMP:
		return c.jumpInstruction("OP_JUMP", 1, offset)
	case OP_LOOP:
		return c.jumpInstruction("OP_LOOP", -1, offset)
	case OP_CALL:
		return c.byteInstruction("OP_CALL", offset)
	case OP_CREATE_LIST:
		return c.byteInstruction("OP_CREATE_LIST", offset)
	case OP_CREATE_DICT:
		return c.byteInstruction("OP_CREATE_DICT", offset)
	case OP_INDEX:
		return c.simpleInstruction("OP_INDEX", offset)
	case OP_INDEX_ASSIGN:
		return c.simpleInstruction("OP_INDEX_ASSIGN", offset)
	case OP_SLICE:
		return c.simpleInstruction("OP_SLICE", offset)
	case OP_SLICE_ASSIGN:
		return c.simpleInstruction("OP_SLICE_ASSIGN", offset)
	case OP_CLOSURE:

		var s string

		offset++
		constant := c.code[offset]
		offset++
		fmt.Printf("%-16s %04d", "OP_CLOSURE", constant)
		value := c.constants[constant]
		fmt.Printf("  %s\n", value.String())
		function := getFunctionObjectValue(value)
		for j := 0; j < function.upvalueCount; j++ {
			isLocal := c.code[offset]
			offset++
			index := c.code[offset]
			offset++
			if isLocal == 1 {
				s = "local"
			} else {
				s = "upvalue"
			}
			fmt.Printf("%04d      |                     %s %d\n", offset-2, s, index)
		}
		return offset
	case OP_GET_UPVALUE:
		return c.byteInstruction("OP_GET_UPVALUE", offset)
	case OP_SET_UPVALUE:
		return c.byteInstruction("OP_SET_UPVALUE", offset)
	case OP_CLOSE_UPVALUE:
		return c.simpleInstruction("OP_CLOSE_UPVALUE", offset)
	case OP_CLASS:
		return c.constantInstruction("OP_CLASS", offset)
	case OP_GET_PROPERTY:
		return c.constantInstruction("OP_GET_PROPERTY", offset)
	case OP_SET_PROPERTY:
		return c.constantInstruction("OP_SET_PROPERTY", offset)
	case OP_METHOD:
		return c.constantInstruction("OP_METHOD", offset)
	case OP_INVOKE:
		return c.invokeInstruction("OP_INVOKE", offset)
	case OP_INHERIT:
		return c.simpleInstruction("OP_INHERIT", offset)
	case OP_GET_SUPER:
		return c.constantInstruction("OP_INHERIT", offset)
	case OP_SUPER_INVOKE:
		return c.invokeInstruction("OP_SUPER_INVOKE", offset)
	case OP_IMPORT:
		return c.constantInstruction("OP_IMPORT", offset)
	default:
		fmt.Printf("Unknown opcode %d", i)
		return offset + 1
	}
}

func (*Chunk) simpleInstruction(name string, offset int) int {

	fmt.Printf("%s\n", name)
	return offset + 1
}

func (c *Chunk) constantInstruction(name string, offset int) int {

	constant := c.code[offset+1]
	fmt.Printf("%-16s %04d", name, constant)
	value := c.constants[constant]
	fmt.Printf("  %s\n", value.String())
	return offset + 2
}

func (c *Chunk) byteInstruction(name string, offset int) int {

	slot := c.code[offset+1]
	fmt.Printf("%-16s %04d\n", name, slot)
	return offset + 2
}

func (c *Chunk) jumpInstruction(name string, sign int, offset int) int {

	var jump uint16

	jump1 := uint16(c.code[offset+1])
	jump2 := uint16(c.code[offset+2])

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	fmt.Printf("%-16s %04d -> %d \n", name, offset, uint16(offset)+3+(uint16(sign)*jump))
	return offset + 3
}

func (c *Chunk) invokeInstruction(name string, offset int) int {
	constant := c.code[offset+1]
	argCount := c.code[offset+2]
	fmt.Printf("%-16s (%d args) %4d", name, argCount, constant)
	value := c.constants[constant]
	fmt.Printf("  %s\n", value.String())
	return offset + 3
}

func (vm *VM) showGlobals() {
	fmt.Printf("globals:\n")
	for k, v := range vm.globals {
		if v.String() == "<built-in>" {
			continue
		}
		fmt.Printf("%s -> %s  \n", k, v)
	}
}

func (vm *VM) stackTrace() {

	fmt.Printf("                                                         ")
	for i := 0; i < vm.stackTop; i++ {
		v := vm.stack[i]
		s := v.String()
		t := v.typeName()
		if v.Immutable() {
			fmt.Printf("| %s %s(c) |", t, s)
		} else {
			fmt.Printf("| %s %s |", t, s)
		}
	}
	fmt.Printf("\n")
}
