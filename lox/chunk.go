package lox

const (
	OP_RETURN uint8 = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
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
)

type Chunk struct {
	code      []uint8
	constants []Value
	lines     []int
}

func newChunk() *Chunk {

	return &Chunk{
		code:      []uint8{},
		constants: []Value{},
		lines:     []int{},
	}
}

func (c *Chunk) writeOpCode(b uint8, line int) {

	c.code = append(c.code, b)
	c.lines = append(c.lines, line)
}

func (c *Chunk) addConstant(v Value) uint8 {

	// if constant is already in list, reuse it
	//ok, idx := c.inConstants(v)
	//if ok {
	//	return idx
	//}
	c.constants = append(c.constants, v)
	return uint8(len(c.constants) - 1)
}
