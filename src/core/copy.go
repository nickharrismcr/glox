package core

// copy.go implements a structural deep-copy of Value used by the thread
// module (see docs/thread-module-plan.md) to isolate a closure's captured
// state -- and its arguments -- when handing it to a new goroutine-backed
// VM. It's structurally parallel to pickle.go's encodeObjectValue, but
// builds fresh Go objects directly instead of writing bytes, since both
// VMs share one address space and no serialisation is needed.
//
// Unlike pickle.go, cycles are not an error here: a memo map[Object]Object
// (original -> its copy) is threaded through the whole walk, so a
// self-referential structure copies into an equally self-referential copy,
// the same way Python's copy.deepcopy handles cycles.
//
// What gets copied vs. shared is deliberate, not exhaustive: only mutable
// data reachable from a spawned closure's own captured upvalues needs
// isolating. Compiled code (*FunctionObject/*Chunk) and *ClassObject are
// shared by pointer -- they're immutable after compilation, and sharing a
// class is what lets an instance value cross into a new thread without any
// pickle-style by-name class resolution (both VMs already have the same
// live class object). See the "scope limitation" in
// docs/thread-module-plan.md: globals/class statics/module Vars reachable
// through a *ClassObject's methods are NOT isolated by this file -- only a
// closure's own upvalues are.

// CopyValueForSpawn returns a deep copy of v suitable for handing to a
// different VM instance. memo must be non-nil and is typically shared
// across every value copied for one spawn (the closure's upvalues, plus
// any extra arguments), so a structure referenced from more than one of
// them is only copied once and both copies end up pointing at the same
// new object, mirroring the original aliasing.
func CopyValueForSpawn(v Value, memo map[Object]Object) Value {
	switch v.Type {
	case VAL_OBJ:
		return copyObjectValueForSpawn(v, memo)
	case VAL_VEC2:
		vec := v.AsVec2()
		return MakeVec2Value(vec.X, vec.Y, v.Immut)
	case VAL_VEC3:
		vec := v.AsVec3()
		return MakeVec3Value(vec.X, vec.Y, vec.Z, v.Immut)
	case VAL_VEC4:
		vec := v.AsVec4()
		return MakeVec4Value(vec.X, vec.Y, vec.Z, vec.W, v.Immut)
	default:
		// nil/bool/int/float: already value types, no aliasing possible.
		return v
	}
}

func copyObjectValueForSpawn(v Value, memo map[Object]Object) Value {
	switch v.Obj.GetType() {
	case OBJECT_STRING:
		// Interned and immutable -- safe to share as-is, no copy needed.
		return v

	case OBJECT_LIST:
		list := v.AsList()
		if copy, ok := memo[list]; ok {
			return MakeObjectValue(copy, false)
		}
		newList := MakeListObject(make([]Value, len(list.Items)), list.Tuple)
		memo[list] = newList
		for i, item := range list.Items {
			newList.Items[i] = CopyValueForSpawn(item, memo)
		}
		return MakeObjectValue(newList, false)

	case OBJECT_DICT:
		dict := v.AsDict()
		if copy, ok := memo[dict]; ok {
			return MakeObjectValue(copy, false)
		}
		newDict := MakeDictObject(make(map[int]Value, len(dict.Items)))
		memo[dict] = newDict
		for k, val := range dict.Items {
			newDict.Items[k] = CopyValueForSpawn(val, memo)
		}
		return MakeObjectValue(newDict, false)

	case OBJECT_INSTANCE:
		inst := v.AsInstance()
		if copy, ok := memo[inst]; ok {
			return MakeObjectValue(copy, false)
		}
		// Class shared by pointer -- see the package doc comment. No
		// pickle-style by-name resolution needed: both VMs already have
		// the same live *ClassObject.
		newInst := MakeInstanceObject(inst.Class)
		memo[inst] = newInst
		for k, val := range inst.Fields {
			newInst.Fields[k] = CopyValueForSpawn(val, memo)
		}
		return MakeObjectValue(newInst, false)

	case OBJECT_CLOSURE:
		closure := v.AsClosure()
		if copy, ok := memo[closure]; ok {
			return MakeObjectValue(copy, false)
		}
		return MakeObjectValue(CopyClosureForSpawn(closure, memo), false)

	case OBJECT_BOUNDMETHOD:
		bm := v.AsBoundMethod()
		if copy, ok := memo[bm]; ok {
			return MakeObjectValue(copy, false)
		}
		newBM := &BoundMethodObject{
			Receiver: CopyValueForSpawn(bm.Receiver, memo),
			Method:   bm.Method, // shared: closures over class methods aren't isolated (see scope limitation)
		}
		memo[bm] = newBM
		return MakeObjectValue(newBM, false)

	case OBJECT_CLASS:
		// Shared by pointer -- deliberate, see package doc comment.
		return v

	default:
		// Modules, files, native/graphics handles, etc.: shared by
		// pointer, same caveat class as raylib, just narrower in scope.
		return v
	}
}

// CopyClosureForSpawn returns a deep copy of closure suitable for running
// in a different VM instance: Function (compiled bytecode) is shared by
// pointer -- immutable after compilation, safe to alias -- but every
// upvalue is snapshotted via copyUpvalueForSpawn, so the copy shares no
// mutable captured state with the original.
//
// memo may be nil, in which case a fresh one is used for just this
// closure's own upvalues; pass a shared memo (e.g. one also used to copy
// the call's arguments) if aliasing across the closure and its arguments
// needs to be preserved.
func CopyClosureForSpawn(closure *ClosureObject, memo map[Object]Object) *ClosureObject {
	if memo == nil {
		memo = map[Object]Object{}
	}
	if copy, ok := memo[closure]; ok {
		return copy.(*ClosureObject)
	}
	newClosure := &ClosureObject{
		Function:     closure.Function,
		Upvalues:     make([]*UpvalueObject, len(closure.Upvalues)),
		UpvalueCount: closure.UpvalueCount,
	}
	memo[closure] = newClosure
	for i, uv := range closure.Upvalues {
		newClosure.Upvalues[i] = copyUpvalueForSpawn(uv, memo)
	}
	return newClosure
}

// copyUpvalueForSpawn snapshots the current value an upvalue refers to.
// *uv.Location is always the live value regardless of open/closed state
// (Location points at the enclosing frame's stack slot while open, or at
// &Closed once closed) -- see obj_upval.go. The copy is born already
// closed: there's no live stack in the new VM for it to reference, since
// it's a one-time snapshot, not a continuing binding, and it's never
// linked into any VM's openUpValues list.
func copyUpvalueForSpawn(uv *UpvalueObject, memo map[Object]Object) *UpvalueObject {
	newUv := &UpvalueObject{}
	newUv.Closed = CopyValueForSpawn(*uv.Location, memo)
	newUv.Location = &newUv.Closed
	return newUv
}
