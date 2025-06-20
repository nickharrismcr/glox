package core

type UpvalueObject struct {
	Location *Value
	Slot     int
	Next     *UpvalueObject
	Closed   Value
}

func MakeUpvalueObject(value *Value, slot int) *UpvalueObject {

	return &UpvalueObject{
		Location: value,
		Slot:     slot,
		Next:     nil,
		Closed:   NIL_VALUE,
	}
}

func (UpvalueObject) IsObject() {}

func (UpvalueObject) GetType() ObjectType {

	return OBJECT_UPVALUE
}

func (f *UpvalueObject) String() string {

	return "Upvalue"
}
func

// -------------------------------------------------------------------------------------------
(t *UpvalueObject) IsBuiltIn() bool {
	return false
}
