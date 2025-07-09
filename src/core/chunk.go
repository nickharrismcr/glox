package core

import (
	"bytes"
	bin "encoding/binary"
	"glox/src/util"
)

const (
	OP_RETURN uint8 = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD_NUMERIC
	OP_CONCAT
	OP_ADD_VECTOR
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_NOT
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_PRINT
	OP_STR
	OP_POP
	OP_DEFINE_GLOBAL
	OP_DEFINE_GLOBAL_CONST
	OP_GET_GLOBAL
	OP_SET_GLOBAL
	OP_GET_LOCAL
	OP_SET_LOCAL
	OP_JUMP_IF_FALSE
	OP_JUMP
	OP_LOOP
	OP_CALL
	OP_MODULUS
	OP_CREATE_LIST
	OP_CREATE_DICT
	OP_INDEX
	OP_INDEX_ASSIGN
	OP_SLICE
	OP_SLICE_ASSIGN
	OP_CLOSURE
	OP_GET_UPVALUE
	OP_SET_UPVALUE
	OP_CLOSE_UPVALUE
	OP_CLASS
	OP_SET_PROPERTY
	OP_GET_PROPERTY
	OP_METHOD
	OP_STATIC_METHOD
	OP_INVOKE
	OP_INHERIT
	OP_GET_SUPER
	OP_SUPER_INVOKE
	OP_IMPORT
	OP_TRY
	OP_END_TRY
	OP_EXCEPT
	OP_END_EXCEPT
	OP_FINALLY
	OP_RAISE
	OP_FOREACH
	OP_NEXT
	OP_END_FOREACH
	OP_CREATE_TUPLE
	OP_IN
	OP_BREAKPOINT
	OP_UNPACK
	OP_IMPORT_FROM
	OP_ONE
	OP_DUP
	OP_INC_LOCAL
)

func NewChunk(filename string) *Chunk {

	return &Chunk{
		Code:      []uint8{},
		Constants: []Value{},
		Lines:     []int{},
		Filename:  filename,
		LocalVars: []LocalVarInfo{},
	}
}

func MakeChunk(filename string, code []uint8, constants []Value, lines []int) *Chunk {
	return &Chunk{
		Code:      code,
		Constants: constants,
		Lines:     lines,
		Filename:  filename,
	}
}

func (c *Chunk) WriteOpCode(b uint8, line int) {

	c.Code = append(c.Code, b)
	c.Lines = append(c.Lines, line)
}

func (c *Chunk) AddConstant(v Value) uint8 {

	// if constant is already in list, reuse it - but not if a function/method
	ok, idx := c.InConstants(v)
	if ok {
		return idx
	}
	c.Constants = append(c.Constants, v)
	return uint8(len(c.Constants) - 1)
}

func (c *Chunk) InConstants(v Value) (bool, uint8) {

	if v.IsObj() {
		t := v.Obj.GetType()
		if t == OBJECT_BOUNDMETHOD || t == OBJECT_CLOSURE || t == OBJECT_FUNCTION {
			return false, 0
		}
	}

	for i, cv := range c.Constants {
		if ValuesEqual(v, cv, true) {
			return true, uint8(i)
		}
	}
	return false, 0
}

func (c *Chunk) Serialise(b *bytes.Buffer) {

	util.WriteMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.Code)))
	b.Write(c.Code)
	util.WriteMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.Lines)))
	for _, line := range c.Lines {
		bin.Write(b, bin.LittleEndian, uint32(line))
	}
	util.WriteMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.Constants)))
	for _, v := range c.Constants {
		v.Serialise(b)
	}
	util.WriteMarker(b)
	bin.Write(b, bin.LittleEndian, uint32(len(c.Filename)))
	b.Write([]byte(c.Filename))
	util.WriteMarker(b)
}
