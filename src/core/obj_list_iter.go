package core

type ListIteratorObject struct {
	Data  *ListObject
	Index int
}

func MakeListIteratorObject(value *ListObject) *ListIteratorObject {
	return &ListIteratorObject{
		Data:  value,
		Index: 0,
	}
}

func (o *ListIteratorObject) String() string {
	return "<iterator>"
}

func (o *ListIteratorObject) IsObject() {}

func (o *ListIteratorObject) IsBuiltIn() bool {
	return false
}
func (o *ListIteratorObject) List() string {
	return "<iterator>"
}

func (o *ListIteratorObject) GetType() ObjectType {

	return OBJECT_ITERATOR
}

func (o *ListIteratorObject) Next() Value {

	if o.Index >= o.Data.GetLength() {
		return MakeNilValue()
	}
	rv, _ := o.Data.Index(o.Index)
	o.Index++
	return rv
}
