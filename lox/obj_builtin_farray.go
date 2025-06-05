package lox

import "fmt"

type FloatArray struct {
	width, height int
	data          []float64
}

type FloatArrayObject struct {
	BuiltInObject
	value *FloatArray
}

func makeFloatArrayObject(w int, h int) *FloatArrayObject {

	return &FloatArrayObject{
		BuiltInObject: BuiltInObject{},
		value:         &FloatArray{width: w, height: h, data: make([]float64, w*h)},
	}
}

func (o *FloatArrayObject) String() string {
	return fmt.Sprintf("<FloatArray %dx%d>", o.value.width, o.value.height)
}

func (o *FloatArrayObject) Type() ObjectType {
	return OBJECT_FLOAT_ARRAY
}

func (o *FloatArrayObject) GetMethod(name string) *BuiltInObject {
	switch name {
	case "get":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				xval := vm.stack[arg_stackptr]
				yval := vm.stack[arg_stackptr+1]
				x := xval.asInt()
				y := yval.asInt()
				return makeFloatValue(o.value.data[y*o.value.width+x], false)
			},
		}
	case "set":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				xval := vm.stack[arg_stackptr]
				yval := vm.stack[arg_stackptr+1]
				fval := vm.stack[arg_stackptr+2]
				x := xval.asInt()
				y := yval.asInt()
				f := fval.asFloat()
				o.value.data[y*o.value.width+x] = f
				return makeNilValue()
			},
		}
	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
