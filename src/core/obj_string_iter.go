package core

type StringIteratorObject struct {
	Data  StringObject
	Index int
}

func MakeStringIteratorObject(value StringObject) *StringIteratorObject {
	return &StringIteratorObject{
		Data:  value,
		Index: 0,
	}
}

func (o *StringIteratorObject) IsObject() {}

func (o *StringIteratorObject) IsBuiltIn() bool {
	return false
}
func (o *StringIteratorObject) String() string {
	return "<iterator>"
}

func (o *StringIteratorObject) GetType() ObjectType {

	return OBJECT_ITERATOR
}

func (o *StringIteratorObject) Next() Value {

	if o.Index >= o.Data.GetLength() {
		return MakeNilValue()
	}
	rv, _ := o.Data.Index(o.Index)
	o.Index++
	return rv
}
