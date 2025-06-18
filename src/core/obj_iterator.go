package core

type IteratorObject struct {
	Data  Iterable
	Index int
}

func MakeIteratorObject(value Iterable) *IteratorObject {
	return &IteratorObject{
		Data:  value,
		Index: 0,
	}
}

func (IteratorObject) IsObject() {}

func (IteratorObject) IsBuiltIn() bool {
	return false
}
func (o *IteratorObject) String() string {
	return "<iterator>"
}

func (IteratorObject) GetType() ObjectType {

	return OBJECT_ITERATOR
}

func (o *IteratorObject) Next() Value {

	if o.Index >= o.Data.GetLength() {
		return MakeNilValue()
	}
	rv, _ := o.Data.Index(o.Index)
	o.Index++
	return rv
}
