//go:build !cgo && windows

package builtin

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// rl.DrawMeshInstanced takes an int32 here (raylib_purego.go, the pure-Go backend used on Windows builds without a C compiler on PATH).
func drawMeshInstanced(mesh rl.Mesh, material rl.Material, transforms []rl.Matrix, count int) {
	rl.DrawMeshInstanced(mesh, material, transforms, int32(count))
}
