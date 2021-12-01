package lox

import (
	"fmt"
	"math"
	"time"
)

func (vm *VM) defineBuiltInFunctions() {

	vm.defineBuiltIn("args", argsBuiltIn)
	vm.defineBuiltIn("clock", clockBuiltIn)
	vm.defineBuiltIn("str", strBuiltIn)
	vm.defineBuiltIn("len", lenBuiltIn)
	vm.defineBuiltIn("sin", sinBuiltIn)
	vm.defineBuiltIn("cos", cosBuiltIn)
	vm.defineBuiltIn("append", appendBuiltIn)
	vm.defineBuiltIn("float", floatBuiltIn)
	vm.defineBuiltIn("int", intBuiltIn)
	vm.defineBuiltIn("join", joinBuiltIn)
	vm.defineBuiltIn("keys", keysBuiltIn)
}

func keysBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to keys.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]

	switch val := val.(type) {
	case ObjectValue:
		if val.isDictObject() {
			d := val.asDict()
			return d.keys()
		}
	}
	vm.runTimeError("Argument to keys must be a dictionary.")
	return makeNilValue()
}

func joinBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	var err string = ""

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to join.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	err = "Argument 2 to join must be a string."

	switch val := val.(type) {
	case ObjectValue:
		if val.isListObject() {
			err = "Argument 2 to join must be a string."
			l := val.asList()
			val2 := vm.stack[arg_stackptr+1]
			switch val2 := val2.(type) {
			case ObjectValue:
				if val2.isStringObject() {
					rv, errj := l.join(val2.asString())
					if errj == nil {
						return rv
					} else {
						err = errj.Error()
					}
				}
			}
		}
	}
	vm.runTimeError(err)
	return makeNilValue()
}

func argsBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	argvList := []Value{}
	for _, a := range vm.args {
		argvList = append(argvList, makeObjectValue(makeStringObject(a), true))
	}
	list := makeListObject(argvList)
	return makeObjectValue(list, false)
}

func floatBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.(type) {
	case FloatValue:
		return arg
	case IntValue:
		return makeFloatValue(float64(asInt(arg)), false)
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func intBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg.(type) {
	case IntValue:
		return arg
	case FloatValue:
		return makeIntValue(int(asFloat(arg)), false)
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func clockBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeFloatValue(float64(elapsed.Seconds()), false)
}

// str(x)
func strBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Single argument expected.")
		return makeNilValue()
	}
	arg := vm.stack[arg_stackptr]

	switch arg := arg.(type) {

	case FloatValue:
		s := arg.String()
		so := makeStringObject(s)
		return makeObjectValue(so, false)

	case IntValue:
		s := arg.String()
		so := makeStringObject(s)
		return makeObjectValue(so, false)

	case NilValue:
		so := makeStringObject("nil")
		return makeObjectValue(so, false)

	case ObjectValue:
		o := arg
		switch o.value.getType() {
		case OBJECT_STRING:
			return arg
		case OBJECT_FUNCTION:
			return makeObjectValue(makeStringObject("<func>"), false)
		case OBJECT_LIST:
			return makeObjectValue(makeStringObject(o.String()), false)
		case OBJECT_DICT:
			return makeObjectValue(makeStringObject(o.String()), false)
		case OBJECT_CLASS:
			return makeObjectValue(makeStringObject(o.String()), false)
		case OBJECT_INSTANCE:
			return makeObjectValue(makeStringObject(o.String()), false)
		}

	case BooleanValue:
		s := fmt.Sprintf("%t", arg.get())
		so := makeStringObject(s)
		return makeObjectValue(so, false)
	}

	return makeObjectValue(makeStringObject(""), false)
}

// len( string )
func lenBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to len.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	vobj, ok := val.(ObjectValue)
	if !ok {
		vm.runTimeError("Invalid argument type to len.")
		return makeNilValue()
	}
	switch vobj.get().getType() {
	case OBJECT_STRING:
		s := vobj.get().(StringObject).get()
		return makeIntValue(len(s), false)
	case OBJECT_LIST:
		l := vobj.get().(*ListObject).get()
		return makeIntValue(len(l), false)
	}
	vm.runTimeError("Invalid argument type to len.")
	return makeNilValue()
}

// sin(number)
func sinBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to sin.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	vn, ok := vnum.(FloatValue)
	if !ok {
		vm.runTimeError("Invalid argument type to sin.")
		return makeNilValue()
	}
	n := vn.get()
	return makeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to cos.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	vn, ok := vnum.(FloatValue)
	if !ok {
		vm.runTimeError("Invalid argument type to cos.")
		return makeNilValue()
	}
	n := vn.get()
	return makeFloatValue(math.Cos(n), false)
}

// append(obj,value)
func appendBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to append.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	vobj, ok := val.(ObjectValue)
	if !ok {
		vm.runTimeError("Argument 1 to append must be list.")
		return makeNilValue()
	}
	val2 := vm.stack[arg_stackptr+1]
	switch vobj.get().getType() {

	case OBJECT_LIST:
		l := vobj.get().(*ListObject)
		l.append(val2)
		return val
	}
	vm.runTimeError("Argument 1 to append must be list.")
	return makeNilValue()
}
