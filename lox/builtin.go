package lox

import (
	"math"
	"time"
)

func (vm *VM) defineBuiltIns() {

	vm.defineBuiltIn("args", argsBuiltIn)
	vm.defineBuiltIn("clock", clockBuiltIn)
	vm.defineBuiltIn("len", lenBuiltIn)
	vm.defineBuiltIn("sin", sinBuiltIn)
	vm.defineBuiltIn("cos", cosBuiltIn)
	vm.defineBuiltIn("append", appendBuiltIn)
	vm.defineBuiltIn("float", floatBuiltIn)
	vm.defineBuiltIn("int", intBuiltIn)
	vm.defineBuiltIn("join", joinBuiltIn)
	vm.defineBuiltIn("keys", keysBuiltIn)
	vm.defineBuiltIn("lox_mandel", mandelBuiltIn)
}

func keysBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to keys.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]

	switch val.Type {
	case VAL_OBJ:
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

	switch val.Type {
	case VAL_OBJ:
		if val.isListObject() {
			err = "Argument 2 to join must be a string."
			l := val.asList()
			val2 := vm.stack[arg_stackptr+1]
			switch val2.Type {
			case VAL_OBJ:
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

	switch arg.Type {
	case VAL_FLOAT:
		return arg
	case VAL_INT:
		return makeFloatValue(float64(arg.Int), false)
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

	switch arg.Type {
	case VAL_INT:
		return arg
	case VAL_FLOAT:
		return makeIntValue(int(arg.Float), false)
	}
	vm.runTimeError("Argument must be number.")
	return makeNilValue()
}

func clockBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	elapsed := time.Since(vm.starttime)
	return makeFloatValue(float64(elapsed.Seconds()), false)
}

// len( string )
func lenBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to len.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	if val.Type != VAL_OBJ {
		vm.runTimeError("Invalid argument type to len.")
		return makeNilValue()
	}
	switch val.Obj.getType() {
	case OBJECT_STRING:
		s := val.Obj.(StringObject).get()
		return makeIntValue(len(s), false)
	case OBJECT_LIST:
		l := val.Obj.(*ListObject).get()
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

	if vnum.Type != VAL_FLOAT {
		vm.runTimeError("Invalid argument type to sin.")
		return makeNilValue()
	}
	n := vnum.Float
	return makeFloatValue(math.Sin(n), false)
}

// cos(number)
func cosBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 1 {
		vm.runTimeError("Invalid argument count to cos.")
		return makeNilValue()
	}
	vnum := vm.stack[arg_stackptr]

	if vnum.Type != VAL_FLOAT {

		vm.runTimeError("Invalid argument type to cos.")
		return makeNilValue()
	}
	n := vnum.Float
	return makeFloatValue(math.Cos(n), false)
}

// append(obj,value)
func appendBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 2 {
		vm.runTimeError("Invalid argument count to append.")
		return makeNilValue()
	}
	val := vm.stack[arg_stackptr]
	if val.Type != VAL_OBJ {
		vm.runTimeError("Argument 1 to append must be list.")
		return makeNilValue()
	}
	val2 := vm.stack[arg_stackptr+1]
	switch val.Obj.getType() {

	case OBJECT_LIST:
		l := val.Obj.(*ListObject)
		l.append(val2)
		return val
	}
	vm.runTimeError("Argument 1 to append must be list.")
	return makeNilValue()
}

func mandelBuiltIn(argCount int, arg_stackptr int, vm *VM) Value {

	if argCount != 5 {
		vm.runTimeError("Invalid argument count to lox_mandel.")
		return makeNilValue()
	}
	ii := vm.stack[arg_stackptr]
	jj := vm.stack[arg_stackptr+1]
	h := vm.stack[arg_stackptr+2]
	w := vm.stack[arg_stackptr+3]
	max := vm.stack[arg_stackptr+4]

	if ii.Type != VAL_INT || jj.Type != VAL_INT || h.Type != VAL_INT || w.Type != VAL_INT || max.Type != VAL_INT {
		vm.runTimeError("Invalid arguments to lox_mandel")
		return makeNilValue()
	}

	i := ii.Int
	j := jj.Int
	height := h.Int
	width := w.Int
	maxIteration := max.Int

	x0 := 4.0*(float64(i)-float64(height)/2)/float64(height) - 1.0
	y0 := 4.0 * (float64(j) - float64(width)/2) / float64(width)
	x, y := 0.0, 0.0
	iteration := 0

	for (x*x+y*y <= 4) && (iteration < maxIteration) {
		xtemp := x*x - y*y + x0
		y = 2*x*y + y0
		x = xtemp
		iteration++
	}

	if iteration == maxIteration {
		return makeIntValue(0, false)
	}
	return makeIntValue(iteration, false)

}
