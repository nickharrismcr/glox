package lox

import (
	"fmt"
)

type Value interface {
	isVal()
	String() string
	Immutable() bool
}

func immutable(v Value) Value {

	switch v := v.(type) {
	case IntValue:
		return makeIntValue(v.get(), true)
	case FloatValue:
		return makeFloatValue(v.get(), true)
	case BooleanValue:
		return makeBooleanValue(v.get(), true)
	case ObjectValue:
		return makeObjectValue(v.get(), true)
	}
	return makeNilValue()
}

func mutable(v Value) Value {

	switch v := v.(type) {
	case IntValue:
		return makeIntValue(v.get(), false)
	case FloatValue:
		return makeFloatValue(v.get(), false)
	case BooleanValue:
		return makeBooleanValue(v.get(), false)
	case ObjectValue:
		return makeObjectValue(v.get(), false)
	}
	return makeNilValue()
}

func valuesEqual(a, b Value, typesMustMatch bool) bool {

	switch a.(type) {
	case BooleanValue:
		switch b.(type) {
		case BooleanValue:
			return a.(BooleanValue).get() == b.(BooleanValue).get()
		default:
			return false
		}

	case IntValue:
		switch b.(type) {
		case IntValue:
			return a.(IntValue).get() == b.(IntValue).get()
		case FloatValue:
			if typesMustMatch {
				return false
			}
			return float64(a.(IntValue).get()) == b.(FloatValue).get()
		default:
			return false
		}
	case FloatValue:
		switch b.(type) {
		case IntValue:
			if typesMustMatch {
				return false
			}
			return a.(FloatValue).get() == float64(b.(IntValue).get())
		case FloatValue:
			return a.(FloatValue).get() == b.(FloatValue).get()
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

func isInt(v Value) bool {
	if _, ok := v.(IntValue); ok {
		return true
	}
	return false
}

func isNumber(v Value) bool {
	switch v.(type) {
	case IntValue:
		return true
	case FloatValue:
		return true
	}
	return false
}

func asFloat(v Value) float64 {
	switch nv := v.(type) {
	case IntValue:
		return float64(nv.get())
	case FloatValue:
		return nv.get()
	}
	return 0.0
}

func asInt(v Value) int {
	switch nv := v.(type) {
	case IntValue:
		return nv.get()
	case FloatValue:
		return int(nv.get())
	}
	return 0
}

func isObject(v Value) bool {
	switch v.(type) {
	case ObjectValue:
		return true
	}
	return false
}

func getStringValue(v Value) string {

	return v.(ObjectValue).asString()
}

func getFunctionObjectValue(v Value) *FunctionObject {
	return v.(ObjectValue).asFunction()
}

func getClosureObjectValue(v Value) *ClosureObject {
	return v.(ObjectValue).asClosure()
}

func getClassObjectValue(v Value) *ClassObject {
	return v.(ObjectValue).asClass()
}

func getInstanceObjectValue(v Value) *InstanceObject {
	return v.(ObjectValue).asInstance()
}

//================================================================================================
type IntValue struct {
	value     int
	immutable bool
}

func (IntValue) isVal() {}

func makeIntValue(v int, immutable bool) IntValue {

	return IntValue{
		value:     v,
		immutable: immutable,
	}
}

func (nv IntValue) Immutable() bool {

	return nv.immutable
}

func (nv IntValue) get() int {

	return nv.value
}

func (nv IntValue) String() string {

	return fmt.Sprintf("%d", nv.value)
}

//================================================================================================
type FloatValue struct {
	value     float64
	immutable bool
}

func (FloatValue) isVal() {}

func makeFloatValue(v float64, immutable bool) FloatValue {

	return FloatValue{
		value:     v,
		immutable: immutable,
	}
}

func (nv FloatValue) Immutable() bool {

	return nv.immutable
}

func (nv FloatValue) get() float64 {

	return nv.value
}

func (nv FloatValue) String() string {

	return fmt.Sprintf("%f", nv.value)
}

//================================================================================================
type BooleanValue struct {
	value     bool
	immutable bool
}

func (BooleanValue) isVal() {}

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

func (NilValue) isVal() {}

func makeNilValue() NilValue {

	return NilValue{
		value: false,
	}
}

func (nv NilValue) Immutable() bool {

	return false
}

func (nv NilValue) String() string {

	return "nil"
}

//================================================================================================
type ObjectValue struct {
	value     Object
	immutable bool
}

func (ObjectValue) isVal() {}

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

func (ov ObjectValue) asString() string {

	return ov.value.(StringObject).get()
}

func (ov ObjectValue) asList() *ListObject {

	return ov.value.(*ListObject)
}

func (ov ObjectValue) asFunction() *FunctionObject {

	return ov.value.(*FunctionObject)
}

func (ov ObjectValue) asNative() *NativeObject {

	return ov.value.(*NativeObject)
}

func (ov ObjectValue) asClosure() *ClosureObject {

	return ov.value.(*ClosureObject)
}

func (ov ObjectValue) asClass() *ClassObject {

	return ov.value.(*ClassObject)
}

func (ov ObjectValue) asInstance() *InstanceObject {

	return ov.value.(*InstanceObject)
}

func (ov ObjectValue) asBoundMethod() *BoundMethodObject {

	return ov.value.(*BoundMethodObject)
}

func (ov ObjectValue) isListObject() bool {

	return ov.value.getType() == OBJECT_LIST
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

func (ov ObjectValue) isClassObject() bool {

	return ov.value.getType() == OBJECT_CLASS
}

func (ov ObjectValue) isInstanceObject() bool {

	return ov.value.getType() == OBJECT_INSTANCE
}

func (ov ObjectValue) isBoundMethodObject() bool {

	return ov.value.getType() == OBJECT_BOUNDMETHOD
}

//================================================================================================
//================================================================================================
