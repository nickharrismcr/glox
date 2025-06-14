package core

type ClosureObject struct {
	Function	*FunctionObject
	Upvalues	[]*UpvalueObject
	UpvalueCount	int
}

func MakeClosureObject(function *FunctionObject) *ClosureObject {

	rv := &ClosureObject{
		Function:	function,
		Upvalues:	[]*UpvalueObject{},
	}
	for i := 0; i < function.UpvalueCount; i++ {
		rv.Upvalues = append(rv.Upvalues, nil)
	}
	rv.UpvalueCount = function.UpvalueCount
	return rv
}

func (ClosureObject) IsObject()	{}

func (ClosureObject) GetType() ObjectType {

	return OBJECT_CLOSURE
}

func (f *ClosureObject) String() string {

	return f.Function.String()
}

// -------------------------------------------------------------------------------------------
