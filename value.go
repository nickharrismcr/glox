package main

import "fmt"

type Value interface {
	isVal()
	String() string
}

func PrintValue(v Value) {
	fmt.Printf("%s\b", v.String())
}

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
