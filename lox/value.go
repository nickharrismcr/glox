package lox

import (
	"fmt"
)

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
		return makeNumberValue(v.(NumberValue).get(), true)
	case BooleanValue:
		return makeBooleanValue(v.(BooleanValue).get(), true)
	case ObjectValue:
		return makeObjectValue(v.(ObjectValue).get(), true)
	}
	return makeNilValue()
}

func mutable(v Value) Value {

	switch v.(type) {
	case NumberValue:
		return makeNumberValue(v.(NumberValue).get(), false)
	case BooleanValue:
		return makeBooleanValue(v.(BooleanValue).get(), false)
	case ObjectValue:
		return makeObjectValue(v.(ObjectValue).get(), false)
	}
	return makeNilValue()
}

func valuesEqual(a, b Value) bool {

	switch a.(type) {
	case BooleanValue:
		switch b.(type) {
		case BooleanValue:
			return a.(BooleanValue).get() == b.(BooleanValue).get()
		default:
			return false
		}
	case NumberValue:
		switch b.(type) {
		case NumberValue:
			return a.(NumberValue).get() == b.(NumberValue).get()
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

func getStringValue(v Value) string {

	return v.(ObjectValue).stringObjectValue()
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

func (nv NumberValue) get() float64 {

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

func (nv BooleanValue) get() bool {

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

	return false
}

func (nv NilValue) get() bool {

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

func (ov ObjectValue) get() Object {

	return ov.value
}

func (ov ObjectValue) String() string {

	return ov.value.String()
}

func (ov ObjectValue) isStringObject() bool {

	return ov.value.getType() == OBJECT_STRING
}

func (ov ObjectValue) stringObjectValue() string {

	return ov.value.(StringObject).get()
}

func (ov ObjectValue) isFunctionObject() bool {

	return ov.value.getType() == OBJECT_FUNCTION
}

func (ov ObjectValue) isNativeFunction() bool {

	return ov.value.getType() == OBJECT_NATIVE
}

func (ov ObjectValue) isClosureObject() bool {

	return ov.value.getType() == OBJECT_CLOSURE
}

//================================================================================================
//================================================================================================
