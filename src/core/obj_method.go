package core

type BoundMethodObject struct {
	receiver Value
	method   *ClosureObject
}

func makeBoundMethodObject(receiver Value, method *ClosureObject) *BoundMethodObject {

	return &BoundMethodObject{
		receiver: receiver,
		method:   method,
	}
}

func (BoundMethodObject) IsObject() {}

func (BoundMethodObject) GetType() ObjectType {

	return OBJECT_BOUNDMETHOD
}

func (f *BoundMethodObject) String() string {

	return f.method.String()
}

// -------------------------------------------------------------------------------------------
