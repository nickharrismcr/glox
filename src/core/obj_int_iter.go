package core

type IntIteratorObject struct {
	Start int
	End   int
	Step  int
	Index int
}

func MakeIntIteratorObject(start, end, step int) *IntIteratorObject {
	return &IntIteratorObject{
		Start: start,
		End:   end,
		Step:  step,
		Index: start,
	}
}

func (o *IntIteratorObject) GetIterator() (Value, bool) {
	return MakeObjectValue(o, true), true
}

func (o *IntIteratorObject) String() string {
	return "<range iterator>"
}

func (o *IntIteratorObject) IsObject() {}

func (o *IntIteratorObject) IsBuiltIn() bool {
	return false
}

func (o *IntIteratorObject) GetType() ObjectType {

	return OBJECT_ITERATOR
}

func (o *IntIteratorObject) Next() Value {

	if o.Index >= o.End {
		return NIL_VALUE
	}
	rv := MakeIntValue(o.Index, true)
	o.Index += o.Step
	return rv

}
