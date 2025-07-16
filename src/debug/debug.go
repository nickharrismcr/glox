package debug

import (
	"fmt"
	"glox/src/core"
)

func Disassemble(c *core.Chunk, name string) {

	core.LogFmt(core.TRACE, "=== %s ===\n", name)
	s := ""
	for _, v := range c.Constants {
		s = s + fmt.Sprintf("[ %s ]", v.String())
	}
	core.Log(core.TRACE, s)
	offset := 0
	for {
		instr := c.Code[offset]
		offset = DisassembleInstruction(c, name, "", 0, instr, offset)
		if offset >= len(c.Code) {
			break
		}
	}
}

var lastoffset int = 0

func DisassembleInstruction(c *core.Chunk, name string, function string, depth int, i uint8, offset int) int {

	if function != "" {
		if depth > 1 {
			name = function
		}
	}
	core.LogFmt(core.TRACE, "%02d : [%-10s] : ", depth, name)

	core.LogFmt(core.TRACE, "%04d ", offset)
	if offset > 0 && c.Lines[offset] == lastoffset {
		core.LogFmt(core.TRACE, "   | ")

	} else {
		core.LogFmt(core.TRACE, "%04d ", c.Lines[offset])
	}
	lastoffset = c.Lines[offset]

	switch i {
	case core.OP_ONE:
		return simpleInstruction("OP_ONE", offset)
	case core.OP_NOOP:
		return simpleInstruction("OP_NOOP", offset)
	case core.OP_DUP:
		return simpleInstruction("OP_DUP", offset)
	case core.OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)
	case core.OP_CONSTANT:
		return constantInstruction(c, "OP_CONSTANT", offset)
	case core.OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case core.OP_ADD_NUMERIC:
		return simpleInstruction("OP_ADD_NUMERIC", offset)
	case core.OP_CONCAT:
		return simpleInstruction("OP_CONCAT_STRING", offset)
	case core.OP_ADD_VECTOR:
		return simpleInstruction("OP_ADD_VECTOR", offset)
	case core.OP_SUBTRACT:
		return simpleInstruction("OP_SUBTRACT", offset)
	case core.OP_MODULUS:
		return simpleInstruction("OP_MODULUS", offset)
	case core.OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case core.OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case core.OP_NIL:
		return simpleInstruction("OP_NIL", offset)
	case core.OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case core.OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case core.OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case core.OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case core.OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case core.OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	case core.OP_PRINT:
		return simpleInstruction("OP_PRINT", offset)
	case core.OP_INC_LOCAL:
		return byteInstruction(c, "OP_INC_LOCAL", offset)
	case core.OP_STR:
		return simpleInstruction("OP_STR", offset)
	case core.OP_POP:
		return simpleInstruction("OP_POP", offset)
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
	case core.OP_CREATE_TUPLE:
		return byteInstruction(c, "OP_CREATE_TUPLE", offset)
	case core.OP_CREATE_DICT:
		return byteInstruction(c, "OP_CREATE_DICT", offset)
	case core.OP_INDEX:
		return simpleInstruction("OP_INDEX", offset)
	case core.OP_INDEX_ASSIGN:
		return simpleInstruction("OP_INDEX_ASSIGN", offset)
	case core.OP_SLICE:
		return simpleInstruction("OP_SLICE", offset)
	case core.OP_SLICE_ASSIGN:
		return simpleInstruction("OP_SLICE_ASSIGN", offset)
	case core.OP_FOREACH:
		return foreachInstruction(c, offset)
	case core.OP_NEXT:
		return nextInstruction(c, "OP_NEXT", -1, offset)
	case core.OP_END_FOREACH:
		return simpleInstruction("OP_END_FOREACH", offset)
	case core.OP_CLOSURE:

		var s string

		offset++
		constant := c.Code[offset]
		offset++
		core.LogFmt(core.TRACE, "%-16s %04d", "OP_CLOSURE", constant)
		value := c.Constants[constant]
		core.LogFmt(core.TRACE, "  %s\n", value.String())
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
			core.LogFmt(core.TRACE, "%04d      |                     %s %d\n", offset-2, s, index)
		}
		return offset
	case core.OP_GET_UPVALUE:
		return byteInstruction(c, "OP_GET_UPVALUE", offset)
	case core.OP_SET_UPVALUE:
		return byteInstruction(c, "OP_SET_UPVALUE", offset)
	case core.OP_CLOSE_UPVALUE:
		return simpleInstruction("OP_CLOSE_UPVALUE", offset)
	case core.OP_CLASS:
		return constantInstruction(c, "OP_CLASS", offset)
	case core.OP_GET_PROPERTY:
		return constantInstruction(c, "OP_GET_PROPERTY", offset)
	case core.OP_SET_PROPERTY:
		return constantInstruction(c, "OP_SET_PROPERTY", offset)
	case core.OP_METHOD:
		return constantInstruction(c, "OP_METHOD", offset)
	case core.OP_STATIC_METHOD:
		return constantInstruction(c, "OP_STATIC_METHOD", offset)
	case core.OP_INVOKE:
		return invokeInstruction(c, "OP_INVOKE", offset)
	case core.OP_INHERIT:
		return simpleInstruction("OP_INHERIT", offset)
	case core.OP_GET_SUPER:
		return constantInstruction(c, "OP_INHERIT", offset)
	case core.OP_SUPER_INVOKE:
		return invokeInstruction(c, "OP_SUPER_INVOKE", offset)
	case core.OP_IMPORT:
		return twoConstantInstruction(c, "OP_IMPORT", offset)
	case core.OP_TRY:
		return addressInstruction(c, "OP_TRY", offset)
	case core.OP_END_TRY:
		return jumpInstruction(c, "OP_END_TRY", 1, offset)
	case core.OP_EXCEPT:
		return constantInstruction(c, "OP_EXCEPT", offset)
	case core.OP_RAISE:
		return simpleInstruction("OP_RAISE", offset)
	case core.OP_END_EXCEPT:
		return simpleInstruction("OP_END_EXCEPT", offset)
	case core.OP_BREAKPOINT:
		return simpleInstruction("OP_BREAKPOINT", offset)
	case core.OP_UNPACK:
		return byteInstruction(c, "OP_UNPACK", offset)
	case core.OP_IMPORT_FROM:
		return importFromInstruction(c, "OP_IMPORT_FROM", offset)
	case core.OP_ADD_NN:
		return twoByteInstruction(c, "OP_ADD_NN", offset)
	case core.OP_ADD_II:
		return twoByteInstruction(c, "OP_ADD_II", offset)
	case core.OP_ADD_FF:
		return twoByteInstruction(c, "OP_ADD_FF", offset)
	case core.OP_INCR_CONST_N:
		return byteConstantInstruction(c, "OP_INCR_CONST_N", offset)
	case core.OP_INCR_CONST_I:
		return byteConstantInstruction(c, "OP_INCR_CONST_I", offset)
	case core.OP_INCR_CONST_F:
		return byteConstantInstruction(c, "OP_INCR_CONST_F", offset)
	default:
		core.LogFmt(core.TRACE, "Unknown opcode %d\n", i)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {

	core.LogFmt(core.TRACE, "%s\n", name)
	return offset + 1
}

func constantInstruction(c *core.Chunk, name string, offset int) int {

	constant := c.Code[offset+1]
	core.LogFmt(core.TRACE, "%-16s %04d", name, constant)
	value := c.Constants[constant]
	core.LogFmt(core.TRACE, "  %s\n", value.String())
	return offset + 2
}

func twoConstantInstruction(c *core.Chunk, name string, offset int) int {

	constant1 := c.Code[offset+1]
	constant2 := c.Code[offset+2]
	core.LogFmt(core.TRACE, "%-16s %04d %04d", name, constant1, constant2)
	value1 := c.Constants[constant1]
	value2 := c.Constants[constant2]
	core.LogFmt(core.TRACE, "  %s", value1.String())
	core.LogFmt(core.TRACE, "  %s\n", value2.String())
	return offset + 3
}
func byteConstantInstruction(c *core.Chunk, name string, offset int) int {

	byte_ := c.Code[offset+1]
	constant := c.Code[offset+2]
	core.LogFmt(core.TRACE, "%-16s %04d %04d", name, byte_, constant)
	value := c.Constants[constant]
	core.LogFmt(core.TRACE, "  %s\n", value.String())
	return offset + 3
}

func byteInstruction(c *core.Chunk, name string, offset int) int {

	byte_ := c.Code[offset+1]
	core.LogFmt(core.TRACE, "%-16s %04d\n", name, byte_)
	return offset + 2
}

func twoByteInstruction(c *core.Chunk, name string, offset int) int {

	byte1 := uint32(c.Code[offset+1])
	byte2 := uint32(c.Code[offset+2])

	core.LogFmt(core.TRACE, "%-16s %04d %04d \n", name, byte1, byte2)
	return offset + 4
}

func jumpInstruction(c *core.Chunk, name string, sign int, offset int) int {

	var jump uint16

	jump1 := uint16(c.Code[offset+1])
	jump2 := uint16(c.Code[offset+2])

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	core.LogFmt(core.TRACE, "%-16s %04d -> %d \n", name, offset, uint16(offset)+3+(uint16(sign)*jump))
	return offset + 3
}
func foreachInstruction(c *core.Chunk, offset int) int {

	var jump uint16
	slot := c.Code[offset+1]
	iterslot := c.Code[offset+2]
	jump1 := uint16(c.Code[offset+3])
	jump2 := uint16(c.Code[offset+4])

	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	core.LogFmt(core.TRACE, "%-16s %04d %04d %04d -> %d \n", "OP_FOREACH", slot, iterslot, jump, uint16(offset)+4+jump)
	return offset + 5
}

func nextInstruction(c *core.Chunk, name string, sign int, offset int) int {

	var jump uint16

	jump1 := uint16(c.Code[offset+1])
	jump2 := uint16(c.Code[offset+2])
	iterSlot := c.Code[offset+3]
	jump = uint16(jump1 << 8)
	jump |= uint16(jump2)

	core.LogFmt(core.TRACE, "%-16s %04d %04d -> %d \n", name, iterSlot, offset, uint16(offset)+3+(uint16(sign)*jump))
	return offset + 4
}
func addressInstruction(c *core.Chunk, name string, offset int) int {

	var address uint16

	addr1 := uint16(c.Code[offset+1])
	addr2 := uint16(c.Code[offset+2])

	address = uint16(addr1 << 8)
	address |= uint16(addr2)

	core.LogFmt(core.TRACE, "%-16s %04d -> %d  \n", name, offset, address)
	return offset + 3
}

func invokeInstruction(c *core.Chunk, name string, offset int) int {
	constant := c.Code[offset+1]
	argCount := c.Code[offset+2]
	core.LogFmt(core.TRACE, "%-16s (%d args) %4d", name, argCount, constant)
	value := c.Constants[constant]
	core.LogFmt(core.TRACE, "  %s\n", value.String())
	return offset + 3
}

func importFromInstruction(c *core.Chunk, name string, offset int) int {
	constant := c.Code[offset+1]
	moduleName := c.Constants[constant].String()
	listLength := c.Code[offset+2]
	if listLength == 0 {
		core.LogFmt(core.TRACE, "%-16s %s -> all\n", name, moduleName)
	} else {
		core.LogFmt(core.TRACE, "%-16s %s (%d items) -> ", name, moduleName, listLength)
		for i := 0; i < int(listLength); i++ {
			constant = c.Code[offset+3+i]
			core.LogFmt(core.TRACE, "  %s", c.Constants[constant].String())
		}
		core.LogFmt(core.TRACE, "\n")
	}
	return offset + 3 + int(listLength)
}
