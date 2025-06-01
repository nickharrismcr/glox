package lox

import (
	"errors"
	"fmt"
	"strings"
)

type StringObject struct {
	chars *string
}

func makeStringObject(s string) StringObject {

	return StringObject{
		chars: &s,
	}
}

func (StringObject) isObject() {}

func (StringObject) getType() ObjectType {

	return OBJECT_STRING
}

func (s StringObject) get() string {

	return *s.chars
}

func (s StringObject) replace(from Value, to Value) Value {

	old := from.asString()
	new := to.asString()
	rv := strings.Replace(*s.chars, old, new, -1)
	return makeObjectValue(makeStringObject(rv), false)
}

func (s StringObject) String() string {

	return fmt.Sprintf("\"%s\"", *s.chars)
}

func (s StringObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.get()) + ix
	}

	if ix < 0 || ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	so := makeStringObject(string(s.get()[ix]))
	return makeObjectValue(so, false), nil
}

func (s StringObject) slice(from_ix, to_ix int) (Value, error) {

	if to_ix < 0 {
		to_ix = len(s.get()) + 1 + to_ix
	}
	if from_ix < 0 {
		from_ix = len(s.get()) + 1 + from_ix
	}

	if to_ix < 0 || to_ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	if from_ix < 0 || from_ix > len(s.get()) {
		return makeNilValue(), errors.New("list subscript out of range")
	}

	so := makeStringObject(s.get()[from_ix:to_ix])
	return makeObjectValue(so, false), nil

}

//-------------------------------------------------------------------------------------------
