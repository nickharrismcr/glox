package core

type BuiltInObject struct {
	function BuiltInFn
}

func MakeBuiltInObject(function BuiltInFn) *BuiltInObject {

	return &BuiltInObject{
		function: function,
	}
}

func (BuiltInObject) IsObject() {}

func (BuiltInObject) GetType() ObjectType {

	return OBJECT_NATIVE
}

func (f *BuiltInObject) String() string {

	return "<built-in>"
}

//-------------------------------------------------------------------------------------------
