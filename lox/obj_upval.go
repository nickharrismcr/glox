package lox

type UpvalueObject struct {
	location *Value
	slot     int
	next     *UpvalueObject
	closed   Value
}

func makeUpvalueObject(value *Value, slot int) *UpvalueObject {

	return &UpvalueObject{
		location: value,
		slot:     slot,
		next:     nil,
		closed:   makeNilValue(),
	}
}

func (UpvalueObject) isObject() {}

func (UpvalueObject) getType() ObjectType {

	return OBJECT_UPVALUE
}

func (f *UpvalueObject) String() string {

	return "Upvalue"
}

// -------------------------------------------------------------------------------------------
