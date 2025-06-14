package lox

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

func (BoundMethodObject) isObject() {}

func (BoundMethodObject) getType() ObjectType {

	return OBJECT_BOUNDMETHOD
}

func (f *BoundMethodObject) String() string {

	return f.method.String()
}

// -------------------------------------------------------------------------------------------
