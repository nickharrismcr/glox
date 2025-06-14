package core

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"glox/src/util"
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
		return MakeIntValue(v.Int, true)
	case VAL_FLOAT:
		return MakeFloatValue(v.Float, true)
	case VAL_BOOL:
		return MakeBooleanValue(v.Bool, true)
	case VAL_OBJ:
		return MakeObjectValue(v.Obj, true)
	}
	return MakeNilValue()
}

func mutable(v Value) Value {
	switch v.Type {
	case VAL_INT:
		return MakeIntValue(v.Int, false)
	case VAL_FLOAT:
		return MakeFloatValue(v.Float, false)
	case VAL_BOOL:
		return MakeBooleanValue(v.Bool, false)
	case VAL_OBJ:
		return MakeObjectValue(v.Obj, false)
	}
	return MakeNilValue()
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
			if av.GetType() != bv.GetType() {
				return false
			}
			return av.String() == bv.String()
		default:
			return false
		}
	}
	return false
}

func (v Value) IsInt() bool    { return v.Type == VAL_INT }
func (v Value) IsFloat() bool  { return v.Type == VAL_FLOAT }
func (v Value) IsNumber() bool { return v.Type == VAL_INT || v.Type == VAL_FLOAT }
func (v Value) IsBool() bool   { return v.Type == VAL_BOOL }

// func (v Value) isNil() bool    { return v.Type == VAL_NIL }
func (v Value) IsObj() bool { return v.Type == VAL_OBJ }

func (v Value) AsFloat() float64 {
	switch v.Type {
	case VAL_INT:
		return float64(v.Int)
	case VAL_FLOAT:
		return v.Float
	default:
		return 0.0
	}
}

func (v Value) AsInt() int {
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
		return v.Obj.GetType() == OBJECT_STRING
	}
	return false
}

func getStringValue(v Value) string {

	return v.AsString().Get()
}

func getFunctionObjectValue(v Value) *FunctionObject {
	return v.AsFunction()
}

func getClosureObjectValue(v Value) *ClosureObject {
	return v.AsClosure()
}

/* func getClassObjectValue(v Value) *ClassObject {
	return v.asClass()
} */

func getInstanceObjectValue(v Value) *InstanceObject {
	return v.AsInstance()
}

// ================================================================================================
func MakeIntValue(i int, immut bool) Value {
	return Value{Type: VAL_INT, Int: i, Immut: immut}
}

func MakeFloatValue(f float64, immut bool) Value {
	return Value{Type: VAL_FLOAT, Float: f, Immut: immut}
}

func MakeBooleanValue(b bool, immut bool) Value {
	return Value{Type: VAL_BOOL, Bool: b, Immut: immut}
}

func MakeNilValue() Value {
	return Value{Type: VAL_NIL}
}

func MakeObjectValue(obj Object, immut bool) Value {
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

func (v Value) IsStringObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_STRING
}

func (v Value) AsString() StringObject {

	return v.Obj.(StringObject)
}

func (v Value) AsList() *ListObject {

	return v.Obj.(*ListObject)
}

func (v Value) AsDict() *DictObject {

	return v.Obj.(*DictObject)
}

func (v Value) AsFunction() *FunctionObject {

	return v.Obj.(*FunctionObject)
}

func (v Value) AsBuiltIn() *BuiltInObject {

	return v.Obj.(*BuiltInObject)
}

func (v Value) AsClosure() *ClosureObject {

	return v.Obj.(*ClosureObject)
}

func (v Value) AsClass() *ClassObject {

	return v.Obj.(*ClassObject)
}

func (v Value) AsInstance() *InstanceObject {

	return v.Obj.(*InstanceObject)
}

func (v Value) AsModule() *ModuleObject {

	return v.Obj.(*ModuleObject)
}

func (v Value) AsBoundMethod() *BoundMethodObject {

	return v.Obj.(*BoundMethodObject)
}

func (v Value) AsFloatArray() *FloatArrayObject {

	return v.Obj.(*FloatArrayObject)
}

func (v Value) IsListObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_LIST
}

// func (v Value) isDictObject() bool {

// 	return v.IsObj() && v.Obj.GetType() == OBJECT_DICT
// }

/*
	 func (v Value) isFunctionObject() bool {

		return v.IsObj() && v.Obj.GetType() == OBJECT_FUNCTION
	}
*/
func (v Value) IsBuiltInObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_NATIVE
}

func (v Value) IsClosureObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_CLOSURE
}

func (v Value) IsClassObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_CLASS
}

/* func (v Value) isInstanceObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_INSTANCE
} */

func (v Value) IsBoundMethodObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_BOUNDMETHOD
}

func (v Value) IsFloatArrayObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_FLOAT_ARRAY
}

func (v *Value) Serialise(buffer *bytes.Buffer) {

	switch v.Type {
	case VAL_FLOAT:
		buffer.Write([]byte{0x01})
		bin.Write(buffer, bin.LittleEndian, v.Float)
	case VAL_INT:
		buffer.Write([]byte{0x02})
		bin.Write(buffer, bin.LittleEndian, uint32(v.Int))
	case VAL_OBJ:
		switch v.Obj.GetType() {
		case OBJECT_STRING:
			buffer.Write([]byte{0x03})
			s := v.AsString().Get()
			bin.Write(buffer, bin.LittleEndian, uint32(len(s)))
			buffer.Write([]byte(s))

		case OBJECT_FUNCTION:
			fo := v.AsFunction()
			buffer.Write([]byte{0x04})
			util.WriteString(buffer, fo.Name.Get())
			bin.Write(buffer, bin.LittleEndian, uint32(fo.Arity))
			bin.Write(buffer, bin.LittleEndian, uint32(fo.UpvalueCount))
			fo.Chunk.Serialise(buffer)
		default:
			panic("serialise object value not handled")
		}
	case VAL_BOOL:
		buffer.Write([]byte{0x05})
		b := byte(0)
		if v.Bool {
			b = byte(1)
		}
		buffer.Write([]byte{b})
	case VAL_NIL:
		buffer.Write([]byte{0x06})
	default:
		panic("serialise value not handled")
	}
}

//================================================================================================
//================================================================================================
