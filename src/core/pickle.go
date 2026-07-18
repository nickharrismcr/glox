package core

import (
	bin "encoding/binary"
	"bytes"
	"errors"
	"fmt"
	"math"
)

// pickle.go implements a standalone, plain-data serialiser for Value used by
// the "pickle" builtin module (pickle.dumps/pickle.loads). It is
// deliberately independent of the .lxc bytecode-cache serialiser in
// bc_cache.go/value.go's Value.Serialise: that pair only ever encodes
// acyclic compile-time constants and panics on anything else, whereas this
// one encodes live runtime values (which can be cyclic) and must never
// panic on bad input or unsupported types -- it always returns an error
// instead, so a single bad value can't crash the interpreter.

const (
	pickleTagNil byte = iota + 1
	pickleTagBool
	pickleTagInt
	pickleTagFloat
	pickleTagString
	pickleTagList
	pickleTagDict
	pickleTagVec2
	pickleTagVec3
	pickleTagVec4
)

// EncodeValue serialises v to a byte slice. Lists/dicts are walked
// recursively; a "currently visiting" set of object identities detects
// cycles (e.g. a list appended to itself) and reports an error rather than
// recursing forever.
func EncodeValue(v Value) ([]byte, error) {
	var buf bytes.Buffer
	visiting := make(map[Object]bool)
	if err := encodeValue(&buf, v, visiting); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, v Value, visiting map[Object]bool) error {
	switch v.Type {
	case VAL_NIL:
		buf.WriteByte(pickleTagNil)
		return nil
	case VAL_BOOL:
		buf.WriteByte(pickleTagBool)
		if v.Data != 0 {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return nil
	case VAL_INT:
		buf.WriteByte(pickleTagInt)
		bin.Write(buf, bin.LittleEndian, int64(v.AsInt()))
		return nil
	case VAL_FLOAT:
		buf.WriteByte(pickleTagFloat)
		bin.Write(buf, bin.LittleEndian, v.AsFloat())
		return nil
	case VAL_VEC2:
		vec := v.AsVec2()
		buf.WriteByte(pickleTagVec2)
		bin.Write(buf, bin.LittleEndian, vec.X)
		bin.Write(buf, bin.LittleEndian, vec.Y)
		return nil
	case VAL_VEC3:
		vec := v.AsVec3()
		buf.WriteByte(pickleTagVec3)
		bin.Write(buf, bin.LittleEndian, vec.X)
		bin.Write(buf, bin.LittleEndian, vec.Y)
		bin.Write(buf, bin.LittleEndian, vec.Z)
		return nil
	case VAL_VEC4:
		vec := v.AsVec4()
		buf.WriteByte(pickleTagVec4)
		bin.Write(buf, bin.LittleEndian, vec.X)
		bin.Write(buf, bin.LittleEndian, vec.Y)
		bin.Write(buf, bin.LittleEndian, vec.Z)
		bin.Write(buf, bin.LittleEndian, vec.W)
		return nil
	case VAL_OBJ:
		return encodeObjectValue(buf, v, visiting)
	default:
		return fmt.Errorf("cannot pickle value of type %d", v.Type)
	}
}

func encodeObjectValue(buf *bytes.Buffer, v Value, visiting map[Object]bool) error {
	switch v.Obj.GetType() {
	case OBJECT_STRING:
		buf.WriteByte(pickleTagString)
		s := v.AsString().Get()
		bin.Write(buf, bin.LittleEndian, uint32(len(s)))
		buf.WriteString(s)
		return nil
	case OBJECT_LIST:
		list := v.AsList()
		if visiting[list] {
			return errors.New("cannot pickle cyclic structure")
		}
		visiting[list] = true
		defer delete(visiting, list)

		buf.WriteByte(pickleTagList)
		tupleFlag := byte(0)
		if list.Tuple {
			tupleFlag = 1
		}
		buf.WriteByte(tupleFlag)
		bin.Write(buf, bin.LittleEndian, uint32(len(list.Items)))
		for _, item := range list.Items {
			if err := encodeValue(buf, item, visiting); err != nil {
				return err
			}
		}
		return nil
	case OBJECT_DICT:
		dict := v.AsDict()
		if visiting[dict] {
			return errors.New("cannot pickle cyclic structure")
		}
		visiting[dict] = true
		defer delete(visiting, dict)

		buf.WriteByte(pickleTagDict)
		bin.Write(buf, bin.LittleEndian, uint32(len(dict.Items)))
		for k, val := range dict.Items {
			name := NameFromID(k)
			bin.Write(buf, bin.LittleEndian, uint32(len(name)))
			buf.WriteString(name)
			if err := encodeValue(buf, val, visiting); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("cannot pickle value of type %s", objectTypeName(v.Obj.GetType()))
	}
}

func objectTypeName(t ObjectType) string {
	switch t {
	case OBJECT_FUNCTION:
		return "function"
	case OBJECT_CLOSURE:
		return "closure"
	case OBJECT_UPVALUE:
		return "upvalue"
	case OBJECT_NATIVE:
		return "native"
	case OBJECT_CLASS:
		return "class"
	case OBJECT_INSTANCE:
		return "instance"
	case OBJECT_BOUNDMETHOD:
		return "bound method"
	case OBJECT_MODULE:
		return "module"
	case OBJECT_FILE:
		return "file"
	case OBJECT_ITERATOR:
		return "iterator"
	default:
		return "object"
	}
}

// pickleReader is a bounds-checked cursor over a byte slice. Every read
// returns an error on truncation rather than panicking, since the input to
// DecodeValue is untrusted (arbitrary bytes from a script or another
// process).
type pickleReader struct {
	data []byte
	pos  int
}

var errTruncated = errors.New("truncated pickle data")

func (r *pickleReader) readByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, errTruncated
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *pickleReader) readBytes(n int) ([]byte, error) {
	if n < 0 || r.pos+n > len(r.data) {
		return nil, errTruncated
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *pickleReader) readUint32() (uint32, error) {
	b, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return bin.LittleEndian.Uint32(b), nil
}

func (r *pickleReader) readInt64() (int64, error) {
	b, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return int64(bin.LittleEndian.Uint64(b)), nil
}

func (r *pickleReader) readFloat64() (float64, error) {
	b, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bin.LittleEndian.Uint64(b)), nil
}

func (r *pickleReader) readString() (string, error) {
	n, err := r.readUint32()
	if err != nil {
		return "", err
	}
	b, err := r.readBytes(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodeValue deserialises a byte slice produced by EncodeValue. It returns
// an error rather than panicking on truncated or malformed input.
func DecodeValue(data []byte) (Value, error) {
	r := &pickleReader{data: data}
	return decodeValue(r)
}

func decodeValue(r *pickleReader) (Value, error) {
	tag, err := r.readByte()
	if err != nil {
		return NIL_VALUE, err
	}
	switch tag {
	case pickleTagNil:
		return NIL_VALUE, nil
	case pickleTagBool:
		b, err := r.readByte()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeBooleanValue(b != 0, false), nil
	case pickleTagInt:
		n, err := r.readInt64()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeIntValue(int(n), false), nil
	case pickleTagFloat:
		f, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeFloatValue(f, false), nil
	case pickleTagString:
		s, err := r.readString()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeStringObjectValue(s, false), nil
	case pickleTagList:
		tupleFlag, err := r.readByte()
		if err != nil {
			return NIL_VALUE, err
		}
		count, err := r.readUint32()
		if err != nil {
			return NIL_VALUE, err
		}
		items := make([]Value, 0, count)
		for i := uint32(0); i < count; i++ {
			item, err := decodeValue(r)
			if err != nil {
				return NIL_VALUE, err
			}
			items = append(items, item)
		}
		return MakeObjectValue(MakeListObject(items, tupleFlag != 0), false), nil
	case pickleTagDict:
		count, err := r.readUint32()
		if err != nil {
			return NIL_VALUE, err
		}
		items := make(map[int]Value, count)
		for i := uint32(0); i < count; i++ {
			key, err := r.readString()
			if err != nil {
				return NIL_VALUE, err
			}
			val, err := decodeValue(r)
			if err != nil {
				return NIL_VALUE, err
			}
			items[InternName(key)] = val
		}
		return MakeObjectValue(MakeDictObject(items), false), nil
	case pickleTagVec2:
		x, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		y, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeVec2Value(x, y, false), nil
	case pickleTagVec3:
		x, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		y, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		z, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeVec3Value(x, y, z, false), nil
	case pickleTagVec4:
		x, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		y, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		z, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		w, err := r.readFloat64()
		if err != nil {
			return NIL_VALUE, err
		}
		return MakeVec4Value(x, y, z, w, false), nil
	default:
		return NIL_VALUE, fmt.Errorf("unknown pickle tag %d", tag)
	}
}
