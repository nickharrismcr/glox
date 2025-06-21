package core

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"glox/src/util"
)

var NIL_VALUE = Value{Type: VAL_NIL}

type ValueType int

const (
	VAL_NIL ValueType = iota
	VAL_BOOL
	VAL_INT
	VAL_FLOAT
	VAL_OBJ
	VAL_VEC2
	VAL_VEC3
	VAL_VEC4
)

type Value struct {
	Type       ValueType
	Int        int
	Float      float64
	Bool       bool
	Obj        Object // your Object interface stays polymorphic
	Immut      bool
	InternedId int // for string objects, this is the interned id, saves an object cast
}

func Immutable(v Value) Value {

	switch v.Type {
	case VAL_INT:
		return MakeIntValue(v.Int, true)
	case VAL_FLOAT:
		return MakeFloatValue(v.Float, true)
	case VAL_BOOL:
		return MakeBooleanValue(v.Bool, true)
	case VAL_OBJ, VAL_VEC2, VAL_VEC3, VAL_VEC4:
		return MakeObjectValue(v.Obj, true)

	}
	return NIL_VALUE
}

func Mutable(v Value) Value {
	switch v.Type {
	case VAL_INT:
		return MakeIntValue(v.Int, false)
	case VAL_FLOAT:
		return MakeFloatValue(v.Float, false)
	case VAL_BOOL:
		return MakeBooleanValue(v.Bool, false)
	case VAL_OBJ, VAL_VEC2, VAL_VEC3, VAL_VEC4:
		return MakeObjectValue(v.Obj, false)
	}
	return NIL_VALUE
}

func ValuesEqual(a, b Value, typesMustMatch bool) bool {

	if a.InternedId != 0 && b.InternedId != 0 && a.InternedId != b.InternedId {
		// if the interned ids are different, we can immediately return false
		// as its either two different value types, or two different strings
		return false
	}

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
			avt := av.GetType()
			bvt := bv.GetType()
			if avt != bvt {
				return false
			}
			return av.String() == bv.String()
		default:
			return false
		}
	case VAL_VEC2:
		switch b.Type {
		case VAL_VEC2:
			av := a.Obj.(*Vec2Object)
			bv := b.Obj.(*Vec2Object)
			return av.X == bv.X && av.Y == bv.Y
		default:
			return false

		}
	case VAL_VEC3:
		switch b.Type {
		case VAL_VEC3:
			av := a.Obj.(*Vec3Object)
			bv := b.Obj.(*Vec3Object)
			return av.X == bv.X && av.Y == bv.Y && av.Z == bv.Z
		default:
			return false

		}
	case VAL_VEC4:
		switch b.Type {
		case VAL_VEC4:
			av := a.Obj.(*Vec4Object)
			bv := b.Obj.(*Vec4Object)
			return av.X == bv.X && av.Y == bv.Y && av.Z == bv.Z && av.W == bv.W
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

func IsString(v Value) bool {
	switch v.Type {
	case VAL_OBJ:
		return v.Obj.GetType() == OBJECT_STRING
	}
	return false
}

func GetStringValue(v Value) string {

	return v.AsString().Get()
}

func GetFunctionObjectValue(v Value) *FunctionObject {
	return v.AsFunction()
}

func GetClosureObjectValue(v Value) *ClosureObject {
	return v.AsClosure()
}

/* func GetClassObjectValue(v Value) *ClassObject {
	return v.asClass()
} */

func GetInstanceObjectValue(v Value) *InstanceObject {
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

func MakeStringObjectValue(s string, immut bool) Value {
	so := MakeStringObject(s)
	return Value{Type: VAL_OBJ, Obj: so, Immut: immut, InternedId: so.InternedId}
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

func (v Value) AsIterator() Iterator {

	return v.Obj.(Iterator)
}

func (v Value) AsListIterator() *ListIteratorObject {
	return v.Obj.(*ListIteratorObject)
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

func (v Value) IsInstanceObject() bool {

	return v.IsObj() && v.Obj.GetType() == OBJECT_INSTANCE
}

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

// built in vectors

func MakeVec2Value(x, y float64, immut bool) Value {
	return Value{Type: VAL_VEC2, Obj: MakeVec2Object(x, y), Immut: immut}
}
func MakeVec3Value(x, y, z float64, immut bool) Value {
	return Value{Type: VAL_VEC3, Obj: MakeVec3Object(x, y, z), Immut: immut}
}
func MakeVec4Value(x, y, z, w float64, immut bool) Value {
	return Value{Type: VAL_VEC4, Obj: MakeVec4Object(x, y, z, w), Immut: immut}
}

func (v Value) AsVec2() *Vec2Object {
	return v.Obj.(*Vec2Object)
}
func (v Value) AsVec3() *Vec3Object {
	return v.Obj.(*Vec3Object)
}
func (v Value) AsVec4() *Vec4Object {
	return v.Obj.(*Vec4Object)
}
func (v Value) IsVec2() bool {
	return v.Type == VAL_VEC2
}
func (v Value) IsVec3() bool {
	return v.Type == VAL_VEC3
}
func (v Value) IsVec4() bool {
	return v.Type == VAL_VEC4
}
