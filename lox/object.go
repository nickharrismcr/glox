package lox

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
	OBJECT_DICT
	OBJECT_CLASS
	OBJECT_INSTANCE
	OBJECT_BOUNDMETHOD
	OBJECT_MODULE
	OBJECT_FILE
)

type Object interface {
	isObject()
	getType() ObjectType
	String() string
}

type BuiltInFn func(argCount int, args_stackptr int, vm *VM) Value

// -------------------------------------------------------------------------------------------
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

func (FunctionObject) isObject() {}

func (FunctionObject) getType() ObjectType {

	return OBJECT_FUNCTION
}

func (f *FunctionObject) String() string {

	if f.name.get() == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.name)
}

// -------------------------------------------------------------------------------------------
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

func (ClosureObject) isObject() {}

func (ClosureObject) getType() ObjectType {

	return OBJECT_CLOSURE
}

func (f *ClosureObject) String() string {

	return f.function.String()
}

// -------------------------------------------------------------------------------------------
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

func (UpvalueObject) isObject() {}

func (UpvalueObject) getType() ObjectType {

	return OBJECT_UPVALUE
}

func (f *UpvalueObject) String() string {

	return "Upvalue"
}

// -------------------------------------------------------------------------------------------
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

type BuiltInObject struct {
	function BuiltInFn
}

func makeBuiltInObject(function BuiltInFn) *BuiltInObject {

	return &BuiltInObject{
		function: function,
	}
}

func (BuiltInObject) isObject() {}

func (BuiltInObject) getType() ObjectType {

	return OBJECT_NATIVE
}

func (f *BuiltInObject) String() string {

	return "<built-in>"
}

//-------------------------------------------------------------------------------------------

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
				return makeNilValue(), errors.New("Non string in join list.")
			}
		}
		if ln > 1 {
			for _, v := range o.items[1:ln] {
				if isString(v) {
					rs = rs + s + getStringValue(v)
				} else {
					return makeNilValue(), errors.New("Non string in join list.")
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

type ClassObject struct {
	name    StringObject
	methods map[string]Value
	super   *ClassObject
}

func makeClassObject(name string) *ClassObject {

	return &ClassObject{
		name:    makeStringObject(name),
		methods: map[string]Value{},
	}
}

func (ClassObject) isObject() {}

func (ClassObject) getType() ObjectType {

	return OBJECT_CLASS
}

func (f *ClassObject) String() string {

	return fmt.Sprintf("<class %s>", f.name.get())
}

func (f *ClassObject) IsSubclassOf(other *ClassObject) bool {
	for c := f; c != nil; c = c.super {
		if c == other {
			return true
		}
	}
	return false
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

func (InstanceObject) isObject() {}

func (InstanceObject) getType() ObjectType {

	return OBJECT_INSTANCE
}

func (f *InstanceObject) String() string {

	return fmt.Sprintf("<instance %s>", f.class.name.get())
}

// -------------------------------------------------------------------------------------------
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

func (BoundMethodObject) isObject() {}

func (BoundMethodObject) getType() ObjectType {

	return OBJECT_BOUNDMETHOD
}

func (f *BoundMethodObject) String() string {

	return f.method.String()
}

// -------------------------------------------------------------------------------------------
type ModuleObject struct {
	name    string
	globals map[string]Value
}

func makeModuleObject(name string, globals map[string]Value) *ModuleObject {

	return &ModuleObject{
		name:    name,
		globals: globals,
	}
}

func (ModuleObject) isObject() {}

func (ModuleObject) getType() ObjectType {

	return OBJECT_MODULE
}

func (f *ModuleObject) String() string {

	return fmt.Sprintf("<module %s>", f.name)
}

// -------------------------------------------------------------------------------------------
type FileObject struct {
	file   *os.File
	closed bool
	eof    bool
	reader *bufio.Reader
	writer *bufio.Writer
}

func makeFileObject(file *os.File) *FileObject {

	return &FileObject{
		file:   file,
		reader: bufio.NewReader(file),
		writer: bufio.NewWriter(file),
	}
}

func (FileObject) isObject() {}

func (FileObject) getType() ObjectType {

	return OBJECT_FILE
}

func (f *FileObject) String() string {

	return "<file>"
}

func (f *FileObject) close() {
	f.writer.Flush()
	f.file.Close()
	f.closed = true
}

func (f *FileObject) readLine() Value {

	if f.eof {
		return makeNilValue()
	}

	line, err := f.reader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil {
		if err.Error() == "EOF" {
			f.eof = true
			if len(line) > 0 {
				return makeObjectValue(makeStringObject(line), false)
			}
		}
	}
	return makeObjectValue(makeStringObject(line), false)
}

func (f *FileObject) write(str Value) {

	s := str.asString()
	s = strings.ReplaceAll(s, `\n`, "\n")
	f.writer.WriteString(s)

}

//-------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------
