package lox

type Environment struct {
	name string
	vars map[string]Value
}

func newEnvironment(name string) *Environment {
	Debugf("Creating new environment '%s' ", name)
	return &Environment{
		name: name,
		vars: map[string]Value{},
	}
}
func (env *Environment) setVar(name string, value Value) {
	if env == nil {
		panic("Cannot set variable in nil environment")
	}
	Debugf("Setting variable %s to %s in environment %s", name, value, env.name)
	env.vars[name] = value
}

func (env *Environment) getVar(name string) (Value, bool) {
	if env == nil {
		panic("Cannot get variable from nil environment")
	}
	value, ok := env.vars[name]

	return value, ok
}
