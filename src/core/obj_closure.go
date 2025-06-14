package core

type ClosureObject struct {
	function     *FunctionObject
	upvalues     []*UpvalueObject
	upvalueCount int
}

func MakeClosureObject(function *FunctionObject) *ClosureObject {

	rv := &ClosureObject{
		function: function,
		upvalues: []*UpvalueObject{},
	}
	for i := 0; i < function.UpvalueCount; i++ {
		rv.upvalues = append(rv.upvalues, nil)
	}
	rv.upvalueCount = function.UpvalueCount
	return rv
}

func (ClosureObject) IsObject() {}

func (ClosureObject) GetType() ObjectType {

	return OBJECT_CLOSURE
}

func (f *ClosureObject) String() string {

	return f.function.String()
}

// -------------------------------------------------------------------------------------------
