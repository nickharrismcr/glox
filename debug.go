package main

import "fmt"

var debugTraceExecution = false
var debugPrintCode = false

var token_names = map[TokenType]string{
	TOKEN_LEFT_PAREN:    "TOKEN_LEFT_PAREN ",
	TOKEN_RIGHT_PAREN:   "TOKEN_RIGHT_PAREN",
	TOKEN_LEFT_BRACE:    "TOKEN_LEFT_BRACE",
	TOKEN_RIGHT_BRACE:   "TOKEN_RIGHT_BRACE",
	TOKEN_COMMA:         "TOKEN_COMMA",
	TOKEN_DOT:           "TOKEN_DOT",
	TOKEN_MINUS:         "TOKEN_MINUS",
	TOKEN_PLUS:          "TOKEN_PLUS",
	TOKEN_SEMICOLON:     "TOKEN_SEMICOLON",
	TOKEN_SLASH:         "TOKEN_SLASH",
	TOKEN_STAR:          "TOKEN_STAR",
	TOKEN_BANG:          "TOKEN_BANG",
	TOKEN_BANG_EQUAL:    "TOKEN_BANG_EQUAL",
	TOKEN_EQUAL:         "TOKEN_EQUAL",
	TOKEN_EQUAL_EQUAL:   "TOKEN_EQUAL_EQUAL",
	TOKEN_GREATER:       "TOKEN_GREATER",
	TOKEN_GREATER_EQUAL: "TOKEN_GREATER_EQUAL",
	TOKEN_LESS:          "TOKEN_LESS",
	TOKEN_LESS_EQUAL:    "TOKEN_LESS_EQUAL",
	TOKEN_IDENTIFIER:    "TOKEN_IDENTIFIER",
	TOKEN_STRING:        "TOKEN_STRING",
	TOKEN_NUMBER:        "TOKEN_NUMBER",
	TOKEN_AND:           "TOKEN_AND",
	TOKEN_CLASS:         "TOKEN_CLASS",
	TOKEN_ELSE:          "TOKEN_ELSE",
	TOKEN_FALSE:         "TOKEN_FALSE",
	TOKEN_FOR:           "TOKEN_FOR",
	TOKEN_FUNC:          "TOKEN_FUNC",
	TOKEN_IF:            "TOKEN_IF",
	TOKEN_NIL:           "TOKEN_NIL",
	TOKEN_OR:            "TOKEN_OR",
	TOKEN_PRINT:         "TOKEN_PRINT",
	TOKEN_RETURN:        "TOKEN_RETURN",
	TOKEN_SUPER:         "TOKEN_SUPER",
	TOKEN_THIS:          "TOKEN_THIS",
	TOKEN_TRUE:          "TOKEN_TRUE",
	TOKEN_VAR:           "TOKEN_VAR",
	TOKEN_WHILE:         "TOKEN_WHILE",
	TOKEN_ERROR:         "TOKEN_ERROR",
	TOKEN_EOF:           "TOKEN_EOF",
}

func (c *Chunk) disassemble(name string) {
	fmt.Printf("=== %s ===\n", name)
	offset := 0
	for {
		instr := c.code[offset]
		offset = c.disassembleInstruction(instr, offset)
		if offset >= len(c.code) {
			break
		}
	}
}

func (c *Chunk) disassembleInstruction(i uint8, offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && c.lines[offset] == c.lines[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%04d ", c.lines[offset])
	}

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
	case OP_POP:
		return c.simpleInstruction("OP_POP", offset)
	default:
		fmt.Printf("Unknown opcode %d", i)
		return offset + 1
	}
}

func (_ *Chunk) simpleInstruction(name string, offset int) int {
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

func (vm *VM) stackTrace() {
	fmt.Printf("                                                         ")
	for i := 0; i < vm.stackTop; i++ {
		v := vm.stack[i]
		fmt.Printf("[ %s ]", v.String())
	}
	fmt.Printf("\n")
}
