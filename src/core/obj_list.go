package core

import (
	"errors"
	"fmt"
	"strings"
)

type ListObject struct {
	Items []Value
	Tuple bool
}

func MakeListObject(items []Value, isTuple bool) *ListObject {

	return &ListObject{
		Items: items,
		Tuple: isTuple,
	}
}

func (ListObject) IsObject() {}

func (ListObject) GetType() ObjectType {

	return OBJECT_LIST
}

func (o *ListObject) GetMethod(name string) *BuiltInObject {

	switch name {
	case "append":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 {
					vm.RunTimeError("append takes one argument.")
					return NIL_VALUE
				}
				val := vm.Peek(0)
				o.Append(val)
				return NIL_VALUE
			},
		}
	case "remove":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 {
					vm.RunTimeError("remove takes one argument.")
					return NIL_VALUE
				}
				val := vm.Peek(0)
				idx := val.Int
				o.Remove(idx)
				return NIL_VALUE
			},
		}

	default:
		return nil
	}
}

func (o *ListObject) Get() []Value {

	return o.Items
}

func (o *ListObject) GetIterator() (Value, bool) {
	return MakeObjectValue(MakeListIteratorObject(o), false), true
}

func (o *ListObject) GetLength() int {
	return len(o.Items)
}

func (o *ListObject) Append(v Value) {
	o.Items = append(o.Items, v)
}

func (o *ListObject) Remove(ix int) {
	if ix < 0 || ix >= len(o.Items) {
		return
	}
	o.Items = append(o.Items[:ix], o.Items[ix+1:]...)
}

func (o *ListObject) Join(s string) (Value, error) {
	rs := ""
	ln := len(o.Items)
	if ln > 0 {
		for _, v := range o.Items[0:1] {
			if IsString(v) {
				rs = GetStringValue(v)
			} else {
				return NIL_VALUE, errors.New("mon string in join list")
			}
		}
		if ln > 1 {
			for _, v := range o.Items[1:ln] {
				if IsString(v) {
					rs = rs + s + GetStringValue(v)
				} else {
					return NIL_VALUE, errors.New("non string in join list")
				}
			}
		}
	}
	return MakeStringObjectValue(rs, false), nil
}

func (o *ListObject) String() string {

	list := []string{}

	for _, v := range o.Items {
		list = append(list, v.String())
	}
	if o.Tuple {
		return fmt.Sprintf("( %s )", strings.Join(list, " , "))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(list, " , "))
}

func (o *ListObject) Add(other *ListObject) *ListObject {

	l := []Value{}
	l = append(l, o.Items...)
	l = append(l, other.Items...)
	return MakeListObject(l, false)
}

func (o *ListObject) Contains(v Value) Value {

	for _, a := range o.Items {
		if ValuesEqual(a, v, true) {
			return MakeBooleanValue(true, true)
		}
	}
	return MakeBooleanValue(false, true)
}

func (o *ListObject) Index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(o.Get()) + ix
	}

	if ix < 0 || ix >= len(o.Get()) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	return o.Get()[ix], nil
}

func (o *ListObject) Slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(o.Items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(o.Items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(o.Items) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(o.Items) {
		return NIL_VALUE, errors.New("list subscript out of range")
	}

	if from_ix > to_ix {
		return NIL_VALUE, errors.New("invalid slice indices")
	}

	lo := MakeListObject(o.Items[from_ix:to_ix], false)
	return MakeObjectValue(lo, false), nil
}

func (o *ListObject) AssignToIndex(ix int, val Value) error {

	if ix < 0 {
		ix = len(o.Get()) + ix
	}

	if ix < 0 || ix > len(o.Get()) {
		return errors.New("list subscript out of range")
	}

	o.Items[ix] = val
	return nil
}

func (o *ListObject) AssignToSlice(from_ix, to_ix int, val Value) error {

	if to_ix < 0 {
		to_ix = len(o.Items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(o.Items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(o.Items) {
		return errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(o.Items) {
		return errors.New("list subscript out of range")
	}

	if from_ix > to_ix {
		return errors.New("invalid slice indices")
	}

	if val.Type == VAL_OBJ {

		if val.IsListObject() {
			lv := val.AsList()
			tmp := []Value{}
			tmp = append(tmp, o.Items[0:from_ix]...)
			tmp = append(tmp, lv.Items...)
			tmp = append(tmp, o.Items[to_ix:]...)
			o.Items = tmp
			return nil
		}
	}

	return errors.New("can only assign list to list slice")
}

// -------------------------------------------------------------------------------------------
func (t *ListObject) IsBuiltIn() bool {
	return false
}
