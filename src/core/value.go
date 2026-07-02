package core

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"math"
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
	Data       uint64  // holds int (cast), float64 bits, or bool (0/1)
	Obj        Object
	Immut      bool
	InternedId int // for string objects, caches the interned id to avoid casting
}

func boolToUint64(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

func Immutable(v Value) Value {

	switch v.Type {
	case VAL_INT:
		return MakeIntValue(int(v.Data), true)
	case VAL_FLOAT:
		return MakeFloatValue(math.Float64frombits(v.Data), true)
	case VAL_BOOL:
		return MakeBooleanValue(v.Data != 0, true)
	case VAL_OBJ:
		return MakeObjectValue(v.Obj, true)
	case VAL_VEC2:
		vec2 := v.Obj.(*Vec2Object)
		return MakeVec2Value(vec2.X, vec2.Y, true)
	case VAL_VEC3:
		vec3 := v.Obj.(*Vec3Object)
		return MakeVec3Value(vec3.X, vec3.Y, vec3.Z, true)
	case VAL_VEC4:
		vec4 := v.Obj.(*Vec4Object)
		return MakeVec4Value(vec4.X, vec4.Y, vec4.Z, vec4.W, true)

	}
	return NIL_VALUE
}

func Mutable(v Value) Value {
	switch v.Type {
	case VAL_INT:
		return MakeIntValue(int(v.Data), false)
	case VAL_FLOAT:
		return MakeFloatValue(math.Float64frombits(v.Data), false)
	case VAL_BOOL:
		return MakeBooleanValue(v.Data != 0, false)
	case VAL_OBJ:
		return MakeObjectValue(v.Obj, false)
	case VAL_VEC2:
		vec2 := v.Obj.(*Vec2Object)
		return MakeVec2Value(vec2.X, vec2.Y, false)
	case VAL_VEC3:
		vec3 := v.Obj.(*Vec3Object)
		return MakeVec3Value(vec3.X, vec3.Y, vec3.Z, false)
	case VAL_VEC4:
		vec4 := v.Obj.(*Vec4Object)
		return MakeVec4Value(vec4.X, vec4.Y, vec4.Z, vec4.W, false)
	}
	return NIL_VALUE
}

func ValuesEqual(a, b Value, typesMustMatch bool) bool {

	if a.InternedId != 0 && b.InternedId != 0 {
		return a.InternedId == b.InternedId
	}

	switch a.Type {
	case VAL_BOOL:
		switch b.Type {
		case VAL_BOOL:
			return a.Data == b.Data
		default:
			return false
		}

	case VAL_INT:
		switch b.Type {
		case VAL_INT:
			return a.Data == b.Data
		case VAL_FLOAT:
			if typesMustMatch {
				return false
			}
			return float64(int(a.Data)) == math.Float64frombits(b.Data)
		default:
			return false
		}
	case VAL_FLOAT:
		switch b.Type {
		case VAL_INT:
			if typesMustMatch {
				return false
			}
			return math.Float64frombits(a.Data) == float64(int(b.Data))
		case VAL_FLOAT:
			return a.Data == b.Data
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
		return float64(int(v.Data))
	case VAL_FLOAT:
		return math.Float64frombits(v.Data)
	default:
		return 0.0
	}
}

func (v Value) AsInt() int {
	switch v.Type {
	case VAL_INT:
		return int(v.Data)
	case VAL_FLOAT:
		return int(math.Float64frombits(v.Data))
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
	return Value{Type: VAL_INT, Data: uint64(i), Immut: immut}
}

func MakeFloatValue(f float64, immut bool) Value {
	return Value{Type: VAL_FLOAT, Data: math.Float64bits(f), Immut: immut}
}

func MakeBooleanValue(b bool, immut bool) Value {
	return Value{Type: VAL_BOOL, Data: boolToUint64(b), Immut: immut}
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
		return fmt.Sprintf("%d", int(v.Data))
	case VAL_FLOAT:
		return fmt.Sprintf("%g", math.Float64frombits(v.Data))
	case VAL_BOOL:
		if v.Data != 0 {
			return "true"
		}
		return "false"
	case VAL_NIL:
		return "nil"
	case VAL_OBJ:
		return v.Obj.String()
	case VAL_VEC2:
		vec2 := v.Obj.(*Vec2Object)
		return fmt.Sprintf("vec2(%g, %g)", vec2.X, vec2.Y)
	case VAL_VEC3:
		vec3 := v.Obj.(*Vec3Object)
		return fmt.Sprintf("vec3(%g, %g, %g)", vec3.X, vec3.Y, vec3.Z)
	case VAL_VEC4:
		vec4 := v.Obj.(*Vec4Object)
		return fmt.Sprintf("vec4(%g, %g, %g, %g)", vec4.X, vec4.Y, vec4.Z, vec4.W)
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

func (v *Value) Serialise(buffer *bytes.Buffer) {

	switch v.Type {
	case VAL_FLOAT:
		buffer.Write([]byte{0x01})
		bin.Write(buffer, bin.LittleEndian, v.Data)
	case VAL_INT:
		buffer.Write([]byte{0x02})
		bin.Write(buffer, bin.LittleEndian, uint32(int(v.Data)))
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
		if v.Data != 0 {
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
