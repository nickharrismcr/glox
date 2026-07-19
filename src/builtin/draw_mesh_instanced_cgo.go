//go:build cgo || !windows

package builtin

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// rl.DrawMeshInstanced takes an int here (rmodels.go, the cgo backend used
// whenever a C compiler is available, which is every non-Windows build plus
// any Windows build with cgo enabled).
func drawMeshInstanced(mesh rl.Mesh, material rl.Material, transforms []rl.Matrix, count int) {
	rl.DrawMeshInstanced(mesh, material, transforms, count)
}
