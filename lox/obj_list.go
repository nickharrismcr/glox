package lox

import (
	"errors"
	"fmt"
	"strings"
)

type ListObject struct {
	items []Value
	tuple bool
}

func makeListObject(items []Value, isTuple bool) *ListObject {

	return &ListObject{
		items: items,
		tuple: isTuple,
	}
}

func (ListObject) isObject() {}

func (ListObject) getType() ObjectType {

	return OBJECT_LIST
}

func (o *ListObject) get() []Value {

	return o.items
}

func (o *ListObject) append(v Value) {
	o.items = append(o.items, v)
}

func (o *ListObject) join(s string) (Value, error) {
	rs := ""
	ln := len(o.items)
	if ln > 0 {
		for _, v := range o.items[0:1] {
			if isString(v) {
				rs = getStringValue(v)
			} else {
				return makeNilValue(), errors.New("mon string in join list")
			}
		}
		if ln > 1 {
			for _, v := range o.items[1:ln] {
				if isString(v) {
					rs = rs + s + getStringValue(v)
				} else {
					return makeNilValue(), errors.New("non string in join list")
				}
			}
		}
	}
	return makeObjectValue(makeStringObject(rs), false), nil
}

func (o *ListObject) String() string {

	list := []string{}

	for _, v := range o.items {
		list = append(list, v.String())
	}
	if o.tuple {
		return fmt.Sprintf("( %s )", strings.Join(list, " , "))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(list, " , "))
}

func (o *ListObject) add(other *ListObject) *ListObject {

	l := []Value{}
	l = append(l, o.items...)
	l = append(l, other.items...)
	return makeListObject(l, false)
}

func (o *ListObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(o.get()) + ix
	}

	if ix < 0 || ix >= len(o.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	return o.get()[ix], nil
}

func (o *ListObject) slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(o.items) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(o.items) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(o.items) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(o.items) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix > to_ix {
		return makeNilValue(), errors.New("invalid slice indices")
	}

	lo := makeListObject(o.items[from_ix:to_ix], false)
	return makeObjectValue(lo, false), nil
}

func (o *ListObject) assignToIndex(ix int, val Value) error {

	if ix < 0 {
		ix = len(o.get()) + ix
	}

	if ix < 0 || ix > len(o.get()) {
		return errors.New("list subscript out of range")
	}

	o.items[ix] = val
	return nil
}

func (o *ListObject) assignToSlice(from_ix, to_ix int, val Value) error {

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

		if val.isListObject() {
			lv := val.asList()
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
