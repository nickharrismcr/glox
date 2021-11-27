package lox

import (
	"errors"
	"fmt"
	"strings"
)

type ObjectType int

const (
	OBJECT_STRING ObjectType = iota
	OBJECT_FUNCTION
	OBJECT_CLOSURE
	OBJECT_UPVALUE
	OBJECT_NATIVE
	OBJECT_LIST
	OBJECT_CLASS
	OBJECT_INSTANCE
	OBJECT_BOUNDMETHOD
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

type NativeFn func(argCount int, args_stackptr int, vm *VM) Value

//-------------------------------------------------------------------------------------------
type FunctionObject struct {
	arity        int
	chunk        *Chunk
	name         StringObject
	upvalueCount int
}

func makeFunctionObject() *FunctionObject {

	return &FunctionObject{
		arity: 0,
		name:  makeStringObject(""),
		chunk: newChunk(),
	}
}

func (_ FunctionObject) isObject() {}

func (_ FunctionObject) getType() ObjectType {

	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {

	if f.name.get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.name)
}

//-------------------------------------------------------------------------------------------
type ClosureObject struct {
	function     *FunctionObject
	upvalues     []*UpvalueObject
	upvalueCount int
}

func makeClosureObject(function *FunctionObject) *ClosureObject {

	rv := &ClosureObject{
		function: function,
		upvalues: []*UpvalueObject{},
	}
	for i := 0; i < function.upvalueCount; i++ {
		rv.upvalues = append(rv.upvalues, nil)
	}
	rv.upvalueCount = function.upvalueCount
	return rv
}

func (_ ClosureObject) isObject() {}

func (_ ClosureObject) getType() ObjectType {

	return OBJECT_CLOSURE
}

func (f *ClosureObject) String() string {

	return f.function.String()
}

//-------------------------------------------------------------------------------------------
type UpvalueObject struct {
	location *Value
	slot     int
	next     *UpvalueObject
	closed   Value
}

func makeUpvalueObject(value *Value, slot int) *UpvalueObject {

	return &UpvalueObject{
		location: value,
		slot:     slot,
		next:     nil,
		closed:   makeNilValue(),
	}
}

func (_ UpvalueObject) isObject() {}

func (_ UpvalueObject) getType() ObjectType {

	return OBJECT_UPVALUE
}

func (f *UpvalueObject) String() string {

	return "Upvalue"
}

//-------------------------------------------------------------------------------------------
type StringObject struct {
	chars *string
}

func makeStringObject(s string) StringObject {

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

	return fmt.Sprintf("\"%s\"", *s.chars)
}

func (s StringObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(s.get()) + ix
	}

	if ix < 0 || ix > len(s.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
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
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix < 0 || from_ix > len(s.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	so := makeStringObject(s.get()[from_ix:to_ix])
	return makeObjectValue(so, false), nil

}

//-------------------------------------------------------------------------------------------

type NativeObject struct {
	function NativeFn
}

func makeNativeObject(function NativeFn) *NativeObject {

	return &NativeObject{
		function: function,
	}
}

func (_ NativeObject) isObject() {}

func (_ NativeObject) getType() ObjectType {

	return OBJECT_NATIVE
}

func (f *NativeObject) String() string {

	return "<built-in>"
}

//-------------------------------------------------------------------------------------------

type ListObject struct {
	items []Value
}

func makeListObject(items []Value) *ListObject {

	return &ListObject{
		items: items,
	}
}

func (_ ListObject) isObject() {}

func (_ ListObject) getType() ObjectType {

	return OBJECT_LIST
}

func (o *ListObject) get() []Value {

	return o.items
}

func (o *ListObject) append(v Value) {
	o.items = append(o.items, v)
}

func (o *ListObject) String() string {

	list := []string{}

	for _, v := range o.items {
		list = append(list, v.String())
	}
	return fmt.Sprintf("[ %s ]", strings.Join(list, " , "))
}

func (o *ListObject) add(other *ListObject) *ListObject {

	l := []Value{}
	l = append(l, o.items...)
	l = append(l, other.items...)
	return makeListObject(l)
}

func (o *ListObject) index(ix int) (Value, error) {

	if ix < 0 {
		ix = len(o.get()) + ix
	}

	if ix < 0 || ix > len(o.get()) {
		return NilValue{}, errors.New("List subscript out of range.")
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
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix < 0 || from_ix > len(o.items) {
		return NilValue{}, errors.New("List subscript out of range.")
	}

	if from_ix > to_ix {
		return NilValue{}, errors.New("Invalid slice indices.")
	}

	lo := makeListObject(o.items[from_ix:to_ix])
	return makeObjectValue(lo, false), nil
}

func (o *ListObject) assignToIndex(ix int, val Value) error {

	if ix < 0 {
		ix = len(o.get()) + ix
	}

	if ix < 0 || ix > len(o.get()) {
		return errors.New("List subscript out of range.")
	}

	o.items[ix] = val
	return nil
}

//-------------------------------------------------------------------------------------------

type ClassObject struct {
	name    StringObject
	methods map[string]Value
}

func makeClassObject(name string) *ClassObject {

	return &ClassObject{
		name:    makeStringObject(name),
		methods: map[string]Value{},
	}
}

func (_ ClassObject) isObject() {}

func (_ ClassObject) getType() ObjectType {

	return OBJECT_CLASS
}

func (f *ClassObject) String() string {

	return fmt.Sprintf("<class %s>", f.name.get())
}

//-------------------------------------------------------------------------------------------

type InstanceObject struct {
	class  *ClassObject
	fields map[string]Value
}

func makeInstanceObject(class *ClassObject) *InstanceObject {

	return &InstanceObject{
		class:  class,
		fields: map[string]Value{},
	}
}

func (_ InstanceObject) isObject() {}

func (_ InstanceObject) getType() ObjectType {

	return OBJECT_INSTANCE
}

func (f *InstanceObject) String() string {

	return fmt.Sprintf("<instance %s>", f.class.name.get())
}

//-------------------------------------------------------------------------------------------
type BoundMethodObject struct {
	receiver Value
	method   *ClosureObject
}

func makeBoundMethodObject(receiver Value, method *ClosureObject) *BoundMethodObject {

	return &BoundMethodObject{
		receiver: receiver,
		method:   method,
	}
}

func (_ BoundMethodObject) isObject() {}

func (_ BoundMethodObject) getType() ObjectType {

	return OBJECT_BOUNDMETHOD
}

func (f *BoundMethodObject) String() string {

	return f.method.String()
}

//-------------------------------------------------------------------------------------------
