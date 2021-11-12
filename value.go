package main

import "fmt"

type Value interface {
	isVal()
	String() string
}

func PrintValue(v Value) {
	fmt.Printf("%s\n", v.String())
}

//================================================================================================
type NumberValue struct {
	value float64
}

func (_ NumberValue) isVal() {}

func MakeNumberValue(v float64) NumberValue {
	return NumberValue{
		value: v,
	}
}

func (nv NumberValue) Get() float64 {
	return nv.value
}

func (nv NumberValue) String() string {
	return fmt.Sprintf("%f", nv.value)
}

//================================================================================================
type BooleanValue struct {
	value bool
}

func (_ BooleanValue) isVal() {}

func MakeBooleanValue(v bool) BooleanValue {
	return BooleanValue{
		value: v,
	}
}

func (nv BooleanValue) Get() bool {
	return nv.value
}

func (nv BooleanValue) String() string {
	if nv.value {
		return "true"
	}
	return "false"
}

//================================================================================================
type NilValue struct {
	value bool
}

func (_ NilValue) isVal() {}

func MakeNilValue() NilValue {
	return NilValue{
		value: false,
	}
}

func (nv NilValue) Get() bool {
	return nv.value
}

func (nv NilValue) String() string {
	return "nil"
}

//================================================================================================
//================================================================================================
//================================================================================================
