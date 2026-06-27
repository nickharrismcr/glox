package core

type Environment struct {
	Name    string
	Vars    map[int]Value // InternedId → Value, for module export and import-all iteration
	Globals []Value       // slot-indexed, for fast OP_GET_GLOBAL
	Defined []bool        // slot-indexed defined flags
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
	env.Vars[stringId] = value
}

// GetVar reads from the InternedId-keyed map (used for module property access).
func (env *Environment) GetVar(stringId int) (Value, bool) {
	if env == nil {
		panic("Cannot get variable from nil environment")
	}
	value, ok := env.Vars[stringId]
	return value, ok
}
