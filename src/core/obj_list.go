package core

import (
	"errors"
	"fmt"
	"strings"
)

type ListObject struct {
	items []Value
	Tuple bool
}

func MakeListObject(items []Value, isTuple bool) *ListObject {

	return &ListObject{
		items: items,
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
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 {
					vm.RunTimeError("append takes one argument.")
					return MakeNilValue()
				}
				val := vm.Peek(0)
				o.Append(val)
				return MakeNilValue()
			},
		}
	case "remove":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				if argCount != 1 {
					vm.RunTimeError("remove takes one argument.")
					return MakeNilValue()
				}
				val := vm.Peek(0)
				idx := val.Int
				o.Remove(idx)
				return MakeNilValue()
			},
		}

	default:
		return nil
	}
}

func (o *ListObject) Get() []Value {

	return o.items
}

func (o *ListObject) Append(v Value) {
	o.items = append(o.items, v)
}

func (o *ListObject) Remove(ix int) {
	if ix < 0 || ix >= len(o.items) {
		return
	}
	o.items = append(o.items[:ix], o.items[ix+1:]...)
}

func (o *ListObject) Join(s string) (Value, error) {
	rs := ""
	ln := len(o.items)
	if ln > 0 {
		for _, v := range o.items[0:1] {
			if IsString(v) {
				rs = GetStringValue(v)
			} else {
				return MakeNilValue(), errors.New("mon string in join list")
			}
		}
		if ln > 1 {
			for _, v := range o.items[1:ln] {
				if IsString(v) {
					rs = rs + s + GetStringValue(v)
				} else {
					return MakeNilValue(), errors.New("non string in join list")
				}
			}
		}
	}
	return MakeObjectValue(MakeStringObject(rs), false), nil
}

func (o *ListObject) String() string {

	list := []string{}

	for _, v := range o.items {
		list = append(list, v.String())
	}
	if o.Tuple {
		return fmt.Sprintf("( %s )", strings.Join(list, " , "))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(list, " , "))
}

func (o *ListObject) Add(other *ListObject) *ListObject {

	l := []Value{}
	l = append(l, o.items...)
	l = append(l, other.items...)
	return MakeListObject(l, false)
}

func (o *ListObject) Contains(v Value) Value {

	for _, a := range o.items {
		if valuesEqual(a, v, true) {
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
		return MakeNilValue(), errors.New("list subscript out of range")
	}

	return o.Get()[ix], nil
}

func (o *ListObject) Slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(o.items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(o.items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(o.items) {
		return MakeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(o.items) {
		return MakeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix > to_ix {
		return MakeNilValue(), errors.New("invalid slice indices")
	}

	lo := MakeListObject(o.items[from_ix:to_ix], false)
	return MakeObjectValue(lo, false), nil
}

func (o *ListObject) AssignToIndex(ix int, val Value) error {

	if ix < 0 {
		ix = len(o.Get()) + ix
	}

	if ix < 0 || ix > len(o.Get()) {
		return errors.New("list subscript out of range")
	}

	o.items[ix] = val
	return nil
}

func (o *ListObject) AssignToSlice(from_ix, to_ix int, val Value) error {

	if to_ix < 0 {
		to_ix = len(o.items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(o.items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(o.items) {
		return errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(o.items) {
		return errors.New("list subscript out of range")
	}

	if from_ix > to_ix {
		return errors.New("invalid slice indices")
	}

	if val.Type == VAL_OBJ {

		if val.IsListObject() {
			lv := val.AsList()
			tmp := []Value{}
			tmp = append(tmp, o.items[0:from_ix]...)
			tmp = append(tmp, lv.items...)
			tmp = append(tmp, o.items[to_ix:]...)
			o.items = tmp
			return nil
		}
	}

	return errors.New("can only assign list to list slice")
}

//-------------------------------------------------------------------------------------------
