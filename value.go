package main

import "fmt"

type Value interface {
	isVal()
	String() string
	Immutable() bool
	SetImmutable(bool)
}

func PrintValue(v Value) {
	fmt.Printf("%s\n", v.String())
}

func valuesEqual(a, b Value) bool {
	switch a.(type) {
	case *BooleanValue:
		switch b.(type) {
		case *BooleanValue:
			return a.(*BooleanValue).Get() == b.(*BooleanValue).Get()
		default:
			return false
		}
	case *NumberValue:
		switch b.(type) {
		case *NumberValue:
			return a.(*NumberValue).Get() == b.(*NumberValue).Get()
		default:
			return false
		}

	case *NilValue:
		switch b.(type) {
		case *NilValue:
			return true
		default:
			return false
		}
	case *ObjectValue:
		switch b.(type) {
		case *ObjectValue:
			av := a.(*ObjectValue).value
			bv := b.(*ObjectValue).value
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

func MakeNumberValue(v float64) *NumberValue {
	return &NumberValue{
		value: v,
	}
}

func (nv *NumberValue) Immutable() bool {
	return nv.immutable
}

func (nv *NumberValue) SetImmutable(b bool) {
	nv.immutable = b
}

func (nv *NumberValue) Get() float64 {
	return nv.value
}

func (nv *NumberValue) String() string {
	return fmt.Sprintf("%f", nv.value)
}

//================================================================================================
type BooleanValue struct {
	value     bool
	immutable bool
}

func (_ BooleanValue) isVal() {}

func MakeBooleanValue(v bool) *BooleanValue {
	return &BooleanValue{
		value: v,
	}
}

func (nv *BooleanValue) Immutable() bool {
	return nv.immutable
}

func (nv *BooleanValue) SetImmutable(b bool) {
	nv.immutable = b
}

func (nv *BooleanValue) Get() bool {
	return nv.value
}

func (nv *BooleanValue) String() string {
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

func MakeNilValue() *NilValue {
	return &NilValue{
		value: false,
	}
}

func (nv *NilValue) Immutable() bool {
	return true
}

func (nv *NilValue) SetImmutable(b bool) {

}

func (nv *NilValue) Get() bool {
	return nv.value
}

func (nv *NilValue) String() string {
	return "nil"
}

//================================================================================================
type ObjectValue struct {
	value     Object
	immutable bool
}

func (_ ObjectValue) isVal() {}

func MakeObjectValue(obj Object) *ObjectValue {
	return &ObjectValue{
		value: obj,
	}
}

func (nv *ObjectValue) Immutable() bool {
	return nv.immutable
}

func (nv *ObjectValue) SetImmutable(b bool) {
	nv.immutable = b
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
