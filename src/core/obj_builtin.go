package core

type BuiltInObject struct {
	Function BuiltInFn
}

func MakeBuiltInObject(function BuiltInFn) *BuiltInObject {

	return &BuiltInObject{
		Function: function,
	}
}

func (BuiltInObject) IsObject() {}

func (BuiltInObject) GetType() ObjectType {

	return OBJECT_NATIVE
}

func (f *BuiltInObject) String() string {

	return "<built-in>"
}
func

// -------------------------------------------------------------------------------------------
(t *BuiltInObject) IsBuiltIn() bool {
	return true
}
