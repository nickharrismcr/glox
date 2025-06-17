package compiler

import (
	"glox/src/core"
)

func PeepHoleOptimise(chunk *core.Chunk) {
	optimized := []uint8{}
	i := 0
	for i < len(chunk.Code) {

		// // Print the current peephole window for debugging
		// windowEnd := i + 6
		// if windowEnd > len(chunk.Code) {
		// 	windowEnd = len(chunk.Code)
		// }
		// fmt.Printf("Peephole [%d:%d]: ", i, windowEnd)
		// for j := i; j < windowEnd; j++ {
		// 	fmt.Printf("%v ", chunk.Code[j])
		// }
		// fmt.Println()

		// Look for pattern: GET_LOCAL x, CONSTANT y, ADD, SET_LOCAL x, POP
		if i+7 < len(chunk.Code) &&
			chunk.Code[i] == core.OP_GET_LOCAL &&
			chunk.Code[i+2] == core.OP_CONSTANT &&
			chunk.Code[i+4] == core.OP_ADD &&
			chunk.Code[i+5] == core.OP_SET_LOCAL &&
			chunk.Code[i+7] == core.OP_POP &&
			// Ensure that the local index matches for both GET_LOCAL and SET_LOCAL
			chunk.Code[i+1] == chunk.Code[i+6] {

			local := chunk.Code[i+1]
			constant := chunk.Code[i+3]
			// Replace with OP_ADD_CONST_LOCAL local, constant
			optimized = append(optimized, core.OP_ADD_CONST_LOCAL)
			optimized = append(optimized, local)
			optimized = append(optimized, constant)
			optimized = append(optimized, core.OP_NOOP)
			optimized = append(optimized, core.OP_NOOP)
			optimized = append(optimized, core.OP_NOOP)
			optimized = append(optimized, core.OP_NOOP)
			optimized = append(optimized, core.OP_NOOP)
			i += 8 // Skip over the optimized pattern
			continue
		}
		// Otherwise, copy instruction as-is
		optimized = append(optimized, chunk.Code[i])
		i++
	}
	chunk.Code = optimized
}
