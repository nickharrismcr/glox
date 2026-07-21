package vm

import (
	"testing"
	"time"

	"glox/src/core"
)

// TestSpawnThreadRecoversPanic exercises a genuine Go panic inside a
// spawned thread's goroutine -- hard to trigger from valid Lox source
// (Lox-level recursion is already capped by FRAMES_MAX before it could
// overflow the Go stack), so this constructs a malformed closure directly:
// an empty Chunk.Code makes run()'s first instruction fetch an
// out-of-range slice index, panicking. Asserts runThreadWorker's recover()
// catches it (Handle.Err is set, surfaced the same way an unhandled Lox
// exception would be) and that this test process itself survives.
func TestSpawnThreadRecoversPanic(t *testing.T) {
	vm := NewVM("test", true)

	fn := core.MakeFunctionObject("broken", nil)
	fn.Chunk = core.NewChunk("broken") // empty Code -- first fetch panics
	closure := core.MakeClosureObject(fn)

	handle, err := vm.SpawnThread(core.MakeObjectValue(closure, false), nil)
	if err != nil {
		t.Fatalf("SpawnThread returned an error: %v", err)
	}

	select {
	case <-handle.Done:
	case <-time.After(5 * time.Second):
		t.Fatal("spawned thread never finished -- panic recovery likely hung instead of returning")
	}

	if handle.Err == nil {
		t.Fatal("expected Handle.Err to be set after a worker panic, got nil")
	}
}
