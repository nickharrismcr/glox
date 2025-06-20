package core

type Environment struct {
	Name string
	Vars map[int]Value
}

func NewEnvironment(name string) *Environment {

	return &Environment{
		Name: name,
		Vars: map[int]Value{},
	}
}
func (env *Environment) SetVar(stringId int, value Value) {
	if env == nil {
		panic("Cannot set variable in nil environment")
	}

	env.Vars[stringId] = value
}

func (env *Environment) GetVar(stringId int) (Value, bool) {
	if env == nil {
		panic("Cannot get variable from nil environment")
	}
	value, ok := env.Vars[stringId]

	return value, ok
}
