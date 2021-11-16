package main

import "fmt"

type Value interface {
	isVal()
	String() string
	Immutable() bool
}

func printValue(v Value) {
	fmt.Printf("%s\n", v.String())
}

func immutable(v Value) Value {
	switch v.(type) {
	case NumberValue:
		return makeNumberValue(v.(NumberValue).Get(), true)
	case BooleanValue:
		return makeBooleanValue(v.(BooleanValue).Get(), true)
	case ObjectValue:
		return makeObjectValue(v.(ObjectValue).Get(), true)
	}
	return makeNilValue()
}

func valuesEqual(a, b Value) bool {
	switch a.(type) {
	case BooleanValue:
		switch b.(type) {
		case BooleanValue:
			return a.(BooleanValue).Get() == b.(BooleanValue).Get()
		default:
			return false
		}
	case NumberValue:
		switch b.(type) {
		case NumberValue:
			return a.(NumberValue).Get() == b.(NumberValue).Get()
		default:
			return false
		}

	case NilValue:
		switch b.(type) {
		case NilValue:
			return true
		default:
			return false
		}
	case ObjectValue:
		switch b.(type) {
		case ObjectValue:
			av := a.(ObjectValue).value
			bv := b.(ObjectValue).value
			if av.getType() != bv.getType() {
				return false
			}
			return av.String() == bv.String()
		default:
			return false
		}
	}
	return false
}

//================================================================================================
type NumberValue struct {
	value     float64
	immutable bool
}

func (_ NumberValue) isVal() {}

func makeNumberValue(v float64, immutable bool) NumberValue {
	return NumberValue{
		value:     v,
		immutable: immutable,
	}
}

func (nv NumberValue) Immutable() bool {
	return nv.immutable
}

func (nv NumberValue) Get() float64 {
	return nv.value
}

func (nv NumberValue) String() string {
	return fmt.Sprintf("%f", nv.value)
}

//================================================================================================
type BooleanValue struct {
	value     bool
	immutable bool
}

func (_ BooleanValue) isVal() {}

func makeBooleanValue(v bool, immutable bool) BooleanValue {
	return BooleanValue{
		value:     v,
		immutable: immutable,
	}
}

func (nv BooleanValue) Immutable() bool {
	return nv.immutable
}

func (nv BooleanValue) Get() bool {
	return nv.value
}

func (nv BooleanValue) String() string {
	if nv.value {
		return "true"
	}
	return "false"
}

//================================================================================================
type NilValue struct {
	value bool
}

func (_ NilValue) isVal() {}

func makeNilValue() NilValue {
	return NilValue{
		value: false,
	}
}

func (nv NilValue) Immutable() bool {
	return true
}

func (nv NilValue) Get() bool {
	return nv.value
}

func (nv NilValue) String() string {
	return "nil"
}

//================================================================================================
type ObjectValue struct {
	value     Object
	immutable bool
}

func (_ ObjectValue) isVal() {}

func makeObjectValue(obj Object, immutable bool) ObjectValue {
	return ObjectValue{
		value:     obj,
		immutable: immutable,
	}
}

func (nv ObjectValue) Immutable() bool {
	return nv.immutable
}

func (ov ObjectValue) Get() Object {
	return ov.value
}

func (ov ObjectValue) String() string {
	return ov.value.String()
}

func (ov ObjectValue) isStringObject() bool {
	return ov.value.getType() == OBJECT_STRING
}

//================================================================================================
//================================================================================================
