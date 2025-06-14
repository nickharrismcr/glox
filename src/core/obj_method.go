package core

type BoundMethodObject struct {
	Receiver Value
	Method   *ClosureObject
}

func MakeBoundMethodObject(receiver Value, method *ClosureObject) *BoundMethodObject {

	return &BoundMethodObject{
		Receiver: receiver,
		Method:   method,
	}
}

func (BoundMethodObject) IsObject() {}

func (BoundMethodObject) GetType() ObjectType {

	return OBJECT_BOUNDMETHOD
}

func (f *BoundMethodObject) String() string {

	return f.Method.String()
}

// -------------------------------------------------------------------------------------------
