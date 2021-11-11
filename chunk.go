package main

const (
	OP_RETURN uint8 = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
)

type Chunk struct {
	code      []uint8
	constants []Value
	lines     []int
}

func NewChunk() *Chunk {
	return &Chunk{
		code:      []uint8{},
		constants: []Value{},
		lines:     []int{},
	}
}

func (c *Chunk) WriteOpCode(b uint8, line int) {
	c.code = append(c.code, b)
	c.lines = append(c.lines, line)
}

func (c *Chunk) AddConstant(v Value) uint8 {
	c.constants = append(c.constants, v)
	return uint8(len(c.constants) - 1)
}
