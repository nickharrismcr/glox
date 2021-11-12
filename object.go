package main

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

type StringObject struct {
	chars *string
}

func MakeStringObject(s string) StringObject {
	return StringObject{
		chars: &s,
	}
}

func (_ StringObject) isObject() {}

func (_ StringObject) getType() ObjectType {
	return OBJECT_STRING
}

func (s StringObject) get() string {
	return *s.chars
}

func (s StringObject) String() string {
	return *s.chars
}
