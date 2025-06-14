package core

type UpvalueObject struct {
	location *Value
	slot     int
	next     *UpvalueObject
	closed   Value
}

func MakeUpvalueObject(value *Value, slot int) *UpvalueObject {

	return &UpvalueObject{
		location: value,
		slot:     slot,
		next:     nil,
		closed:   MakeNilValue(),
	}
}

func (UpvalueObject) IsObject() {}

func (UpvalueObject) GetType() ObjectType {

	return OBJECT_UPVALUE
}

func (f *UpvalueObject) String() string {

	return "Upvalue"
}

// -------------------------------------------------------------------------------------------
