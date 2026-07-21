package core

import (
	"fmt"
	"sync"
)

type Environment struct {
	Name        string
	Vars        map[int]Value // InternedId → Value, for module export and import-all iteration
	Globals     []Value       // slot-indexed, for fast OP_GET_GLOBAL
	Defined     []bool        // slot-indexed defined flags
	GlobalNames []string      // slot → name, shared by every function in the compilation unit (for error messages)

	// varsMu guards Vars only. A built-in module's Environment is shared
	// by reference across the parent VM and every thread-module worker
	// spawned from it, and module.attr = x (OP_SET_PROPERTY) writes to
	// Vars directly -- without this, two threads both writing a module
	// attribute would hit Go's fatal, unrecoverable concurrent-map-write
	// detector. Globals/Defined are NOT guarded here: they're part of the
	// documented "globals aren't isolated across threads" limitation (see
	// docs/thread-module-plan.md), not fixed by this lock.
	varsMu sync.RWMutex
}

// NameForSlot returns the global variable name for a slot, for error messages.
// The slot→name table lives on the top-level chunk, but all functions in a
// compilation unit share this Environment, so inner functions resolve names
// here rather than from their own (empty) chunk.GlobalNames.
func (env *Environment) NameForSlot(slot int) string {
	if slot >= 0 && slot < len(env.GlobalNames) {
		return env.GlobalNames[slot]
	}
	return fmt.Sprintf("#%d", slot)
}

func NewEnvironment(name string) *Environment {
	return &Environment{
		Name: name,
		Vars: map[int]Value{},
	}
}

func (env *Environment) InitGlobals(count int) {
	env.Globals = make([]Value, count)
	env.Defined = make([]bool, count)
}

// GrowGlobals extends the slot-indexed slices to hold at least count entries,
// preserving existing values. Used by the REPL so globals defined on earlier
// lines survive into later ones (unlike InitGlobals, which reallocates).
func (env *Environment) GrowGlobals(count int) {
	if count <= len(env.Globals) {
		return
	}
	g := make([]Value, count)
	copy(g, env.Globals)
	d := make([]bool, count)
	copy(d, env.Defined)
	env.Globals = g
	env.Defined = d
}

// SetGlobal writes to the fast slot-indexed array.
func (env *Environment) SetGlobal(slot int, value Value) {
	env.Globals[slot] = value
	env.Defined[slot] = true
}

// SetVar writes to the InternedId-keyed map (for module exports and import-all iteration).
func (env *Environment) SetVar(stringId int, value Value) {
	if env == nil {
		panic("Cannot set variable in nil environment")
	}
	env.varsMu.Lock()
	defer env.varsMu.Unlock()
	env.Vars[stringId] = value
}

// GetVar reads from the InternedId-keyed map (used for module property access).
func (env *Environment) GetVar(stringId int) (Value, bool) {
	if env == nil {
		panic("Cannot get variable from nil environment")
	}
	env.varsMu.RLock()
	defer env.varsMu.RUnlock()
	value, ok := env.Vars[stringId]
	return value, ok
}

// VarsSnapshot returns a copy of Vars, safe to range over without racing a
// concurrent SetVar (e.g. from a thread-module worker writing the same
// shared module Environment). Used by introspection (inspect module) rather
// than ranging over Vars directly.
func (env *Environment) VarsSnapshot() map[int]Value {
	if env == nil {
		return nil
	}
	env.varsMu.RLock()
	defer env.varsMu.RUnlock()
	snapshot := make(map[int]Value, len(env.Vars))
	for k, v := range env.Vars {
		snapshot[k] = v
	}
	return snapshot
}
