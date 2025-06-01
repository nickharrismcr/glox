package lox

import (
	"errors"
	"fmt"
	"strings"
)

type DictObject struct {
	items map[string]Value
}

func makeDictObject(items map[string]Value) *DictObject {

	return &DictObject{
		items: items,
	}
}

func (DictObject) isObject() {}

func (DictObject) getType() ObjectType {
	return OBJECT_DICT
}

func (o *DictObject) String() string {
	s := "Dict({ "
	for k, v := range o.items {
		s = s + fmt.Sprintf("%s:%s,", k, v.String())
	}
	return s[:len(s)-1] + " })"
}

func (o *DictObject) set(key string, value Value) {

	o.items[key] = value
}

func (o *DictObject) get(key string) (Value, error) {

	rv, ok := o.items[key]
	if !ok {
		return makeNilValue(), errors.New("key not found")
	}
	return rv, nil
}

func (o *DictObject) keys() Value {

	keys := []Value{}
	for k := range o.items {
		key := strings.Replace(k, "\"", "", -1)
		so := makeStringObject(key)
		v := makeObjectValue(so, false)
		keys = append(keys, v)
	}
	return makeObjectValue(makeListObject(keys, false), false)
}

//-------------------------------------------------------------------------------------------
