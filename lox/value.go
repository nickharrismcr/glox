package lox

import (
	"fmt"
)

type ValueType int

const (
	VAL_NIL ValueType = iota
	VAL_BOOL
	VAL_INT
	VAL_FLOAT
	VAL_OBJ
)

type Value struct {
	Type  ValueType
	Int   int
	Float float64
	Bool  bool
	Obj   Object // your Object interface stays polymorphic
	Immut bool
}

func immutable(v Value) Value {

	switch v.Type {
	case VAL_INT:
		return makeIntValue(v.Int, true)
	case VAL_FLOAT:
		return makeFloatValue(v.Float, true)
	case VAL_BOOL:
		return makeBooleanValue(v.Bool, true)
	case VAL_OBJ:
		return makeObjectValue(v.Obj, true)
	}
	return makeNilValue()
}

func mutable(v Value) Value {
	switch v.Type {
	case VAL_INT:
		return makeIntValue(v.Int, false)
	case VAL_FLOAT:
		return makeFloatValue(v.Float, false)
	case VAL_BOOL:
		return makeBooleanValue(v.Bool, false)
	case VAL_OBJ:
		return makeObjectValue(v.Obj, false)
	}
	return makeNilValue()
}

func valuesEqual(a, b Value, typesMustMatch bool) bool {

	switch a.Type {
	case VAL_BOOL:
		switch b.Type {
		case VAL_BOOL:
			return a.Bool == b.Bool
		default:
			return false
		}

	case VAL_INT:
		switch b.Type {
		case VAL_INT:
			return a.Int == b.Int
		case VAL_FLOAT:
			if typesMustMatch {
				return false
			}
			return float64(a.Int) == b.Float
		default:
			return false
		}
	case VAL_FLOAT:
		switch b.Type {
		case VAL_INT:
			if typesMustMatch {
				return false
			}
			return a.Float == float64(b.Int)
		case VAL_FLOAT:
			return a.Float == b.Float
		default:
			return false
		}

	case VAL_NIL:
		switch b.Type {
		case VAL_NIL:
			return true
		default:
			return false
		}
	case VAL_OBJ:
		switch b.Type {
		case VAL_OBJ:
			av := a.Obj
			bv := b.Obj
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

func (v Value) isInt() bool { return v.Type == VAL_INT }

// func (v Value) isFloat() bool  { return v.Type == VAL_FLOAT }
func (v Value) isNumber() bool { return v.Type == VAL_INT || v.Type == VAL_FLOAT }

// func (v Value) isNil() bool    { return v.Type == VAL_NIL }
func (v Value) isObj() bool { return v.Type == VAL_OBJ }

func (v Value) asFloat() float64 {
	switch v.Type {
	case VAL_INT:
		return float64(v.Int)
	case VAL_FLOAT:
		return v.Float
	default:
		return 0.0
	}
}

func (v Value) asInt() int {
	switch v.Type {
	case VAL_INT:
		return v.Int
	case VAL_FLOAT:
		return int(v.Float)
	default:
		return 0
	}
}

func isString(v Value) bool {
	switch v.Type {
	case VAL_OBJ:
		return v.Obj.getType() == OBJECT_STRING
	}
	return false
}

func getStringValue(v Value) string {

	return v.asString().get()
}

func getFunctionObjectValue(v Value) *FunctionObject {
	return v.asFunction()
}

func getClosureObjectValue(v Value) *ClosureObject {
	return v.asClosure()
}

func getClassObjectValue(v Value) *ClassObject {
	return v.asClass()
}

func getInstanceObjectValue(v Value) *InstanceObject {
	return v.asInstance()
}

// ================================================================================================
func makeIntValue(i int, immut bool) Value {
	return Value{Type: VAL_INT, Int: i, Immut: immut}
}

func makeFloatValue(f float64, immut bool) Value {
	return Value{Type: VAL_FLOAT, Float: f, Immut: immut}
}

func makeBooleanValue(b bool, immut bool) Value {
	return Value{Type: VAL_BOOL, Bool: b, Immut: immut}
}

func makeNilValue() Value {
	return Value{Type: VAL_NIL}
}

func makeObjectValue(obj Object, immut bool) Value {
	return Value{Type: VAL_OBJ, Obj: obj, Immut: immut}
}

func (v Value) String() string {
	switch v.Type {
	case VAL_INT:
		return fmt.Sprintf("%d", v.Int)
	case VAL_FLOAT:
		return fmt.Sprintf("%g", v.Float)
	case VAL_BOOL:
		if v.Bool {
			return "true"
		}
		return "false"
	case VAL_NIL:
		return "nil"
	case VAL_OBJ:
		return v.Obj.String()
	default:
		return "<unknown>"
	}
}

func (v Value) Immutable() bool {
	return v.Immut
}

//================================================================================================

func (v Value) isStringObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_STRING
}

func (v Value) asString() StringObject {

	return v.Obj.(StringObject)
}

func (v Value) asList() *ListObject {

	return v.Obj.(*ListObject)
}

func (v Value) asDict() *DictObject {

	return v.Obj.(*DictObject)
}

func (v Value) asFunction() *FunctionObject {

	return v.Obj.(*FunctionObject)
}

func (v Value) asBuiltIn() *BuiltInObject {

	return v.Obj.(*BuiltInObject)
}

func (v Value) asClosure() *ClosureObject {

	return v.Obj.(*ClosureObject)
}

func (v Value) asClass() *ClassObject {

	return v.Obj.(*ClassObject)
}

func (v Value) asInstance() *InstanceObject {

	return v.Obj.(*InstanceObject)
}

func (v Value) asModule() *ModuleObject {

	return v.Obj.(*ModuleObject)
}

func (v Value) asBoundMethod() *BoundMethodObject {

	return v.Obj.(*BoundMethodObject)
}

func (v Value) asFloatArray() *FloatArrayObject {

	return v.Obj.(*FloatArrayObject)
}

func (v Value) isListObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_LIST
}

func (v Value) isDictObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_DICT
}

func (v Value) isFunctionObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_FUNCTION
}

func (v Value) isBuiltInObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_NATIVE
}

func (v Value) isClosureObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_CLOSURE
}

func (v Value) isClassObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_CLASS
}

func (v Value) isInstanceObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_INSTANCE
}

func (v Value) isBoundMethodObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_BOUNDMETHOD
}

func (v Value) isFloatArrayObject() bool {

	return v.isObj() && v.Obj.getType() == OBJECT_FLOAT_ARRAY
}

//================================================================================================
//================================================================================================
