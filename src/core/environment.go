package core

type Environment struct {
	name string
	vars map[string]Value
}

func NewEnvironment(name string) *Environment {

	return &Environment{
		name: name,
		vars: map[string]Value{},
	}
}
func (env *Environment) SetVar(name string, value Value) {
	if env == nil {
		panic("Cannot set variable in nil environment")
	}

	env.vars[name] = value
}

func (env *Environment) GetVar(name string) (Value, bool) {
	if env == nil {
		panic("Cannot get variable from nil environment")
	}
	value, ok := env.vars[name]

	return value, ok
}
