package core

import "testing"

// These exercise CopyValueForSpawn/CopyClosureForSpawn directly at the Go
// level rather than via .lox scripts: the scenarios here (cycles, upvalue
// snapshotting, pointer-identity checks) aren't things a Lox script can
// observe on its own until the thread module actually calls this code, and
// even then a script can't inspect "is this the same underlying pointer"
// the way these tests need to.

func TestCopyValueForSpawn_ListIsClonedNotAliased(t *testing.T) {
	original := MakeListObject([]Value{MakeIntValue(1, false)}, false)
	copyVal := CopyValueForSpawn(MakeObjectValue(original, false), map[Object]Object{})
	copyList := copyVal.AsList()

	if copyList == original {
		t.Fatal("expected a distinct *ListObject, got the same pointer")
	}
	original.Items[0] = MakeIntValue(99, false)
	if copyList.Items[0].AsInt() != 1 {
		t.Fatalf("mutating original leaked into copy: got %v", copyList.Items[0])
	}
}

func TestCopyValueForSpawn_SelfReferentialListPreservesCycle(t *testing.T) {
	l := MakeListObject([]Value{NIL_VALUE}, false)
	l.Items[0] = MakeObjectValue(l, false) // l references itself

	copyVal := CopyValueForSpawn(MakeObjectValue(l, false), map[Object]Object{})
	copyList := copyVal.AsList()

	if copyList == l {
		t.Fatal("expected a distinct copy of the outer list")
	}
	inner := copyList.Items[0].AsList()
	if inner != copyList {
		t.Fatal("expected the copy's self-reference to point at itself, not the original")
	}
}

func TestCopyValueForSpawn_DictIsClonedNotAliased(t *testing.T) {
	k := InternName("k")
	original := MakeDictObject(map[int]Value{k: MakeIntValue(1, false)})
	copyVal := CopyValueForSpawn(MakeObjectValue(original, false), map[Object]Object{})
	copyDict := copyVal.AsDict()

	if copyDict == original {
		t.Fatal("expected a distinct *DictObject, got the same pointer")
	}
	original.Items[k] = MakeIntValue(99, false)
	if copyDict.Items[k].AsInt() != 1 {
		t.Fatalf("mutating original leaked into copy: got %v", copyDict.Items[k])
	}
}

func TestCopyValueForSpawn_InstanceSharesClassClonesFields(t *testing.T) {
	class := MakeClassObject("Point")
	inst := MakeInstanceObject(class)
	fx := InternName("x")
	inst.Fields[fx] = MakeIntValue(1, false)

	copyVal := CopyValueForSpawn(MakeObjectValue(inst, false), map[Object]Object{})
	copyInst := copyVal.AsInstance()

	if copyInst == inst {
		t.Fatal("expected a distinct *InstanceObject, got the same pointer")
	}
	if copyInst.Class != class {
		t.Fatal("expected the class to be shared by pointer, not cloned")
	}
	inst.Fields[fx] = MakeIntValue(99, false)
	if copyInst.Fields[fx].AsInt() != 1 {
		t.Fatalf("mutating original's fields leaked into copy: got %v", copyInst.Fields[fx])
	}
}

func TestCopyValueForSpawn_ClassAndStringAreShared(t *testing.T) {
	class := MakeClassObject("Foo")
	copyVal := CopyValueForSpawn(MakeObjectValue(class, false), map[Object]Object{})
	if copyVal.AsClass() != class {
		t.Fatal("expected *ClassObject to be shared by pointer")
	}

	str := MakeStringObjectValue("hello", false)
	copyStr := CopyValueForSpawn(str, map[Object]Object{})
	if copyStr.Obj != str.Obj {
		t.Fatal("expected string Obj to be shared (interned, immutable)")
	}
}

func TestCopyValueForSpawn_VecIsClonedNotAliased(t *testing.T) {
	original := MakeVec2Value(1, 2, false)
	copyVal := CopyValueForSpawn(original, map[Object]Object{})

	if copyVal.Obj == original.Obj {
		t.Fatal("expected a distinct *Vec2Object, got the same pointer")
	}
	original.AsVec2().X = 99
	if copyVal.AsVec2().X != 1 {
		t.Fatalf("mutating original vec2 leaked into copy: got %v", copyVal.AsVec2().X)
	}
}

func TestCopyClosureForSpawn_UpvalueIsSnapshotAndClosed(t *testing.T) {
	fn := MakeFunctionObject("test", nil)
	fn.UpvalueCount = 1
	closure := MakeClosureObject(fn)

	captured := MakeIntValue(5, false)
	uv := MakeUpvalueObject(&captured, 0) // open: Location points at the local `captured`
	closure.Upvalues[0] = uv

	copyClosure := CopyClosureForSpawn(closure, nil)

	if copyClosure == closure {
		t.Fatal("expected a distinct *ClosureObject, got the same pointer")
	}
	if copyClosure.Function != fn {
		t.Fatal("expected Function to be shared by pointer (immutable compiled code)")
	}
	copyUv := copyClosure.Upvalues[0]
	if copyUv == uv {
		t.Fatal("expected a distinct *UpvalueObject, got the same pointer")
	}
	if *copyUv.Location != MakeIntValue(5, false) {
		t.Fatalf("expected snapshot value 5, got %v", *copyUv.Location)
	}
	// Mutate the original captured variable (as if the spawning VM kept
	// running) and confirm the copy's snapshot is unaffected.
	captured = MakeIntValue(42, false)
	if copyUv.Location.AsInt() != 5 {
		t.Fatalf("mutating the original captured variable leaked into the copy: got %v", copyUv.Location.AsInt())
	}
	// The copy must be born closed: Location should point at its own
	// Closed field, not at anything on the original's stack.
	if copyUv.Location != &copyUv.Closed {
		t.Fatal("expected copied upvalue to be born closed (Location == &Closed)")
	}
}
