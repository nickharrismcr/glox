package core

import "fmt"

type FloatArray struct {
	width, height int
	data          []float64
}

func (f *FloatArray) get(x, y int) float64 {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		panic(fmt.Sprintf("Index out of bounds: (%d, %d) for array size %dx%d", x, y, f.width, f.height))
	}
	return f.data[y*f.width+x]
}

func (f *FloatArray) set(x, y int, value float64) {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		panic(fmt.Sprintf("Index out of bounds: (%d, %d) for array size %dx%d", x, y, f.width, f.height))
	}
	f.data[y*f.width+x] = value
}

func (f *FloatArray) clear(value float64) {
	for i := range f.data {
		f.data[i] = value
	}
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

func (o *FloatArrayObject) GetType() ObjectType {
	return OBJECT_FLOAT_ARRAY
}

func (o *FloatArrayObject) GetMethod(name string) *BuiltInObject {
	switch name {
	case "width":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				return makeIntValue(o.value.width, true)
			},
		}
	case "height":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				return makeIntValue(o.value.height, true)
			},
		}
	case "get":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				x := xval.asInt()
				y := yval.asInt()
				return makeFloatValue(o.value.get(x, y), false)
			},
		}
	case "set":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				fval := vm.Stack(arg_stackptr + 2)
				x := xval.asInt()
				y := yval.asInt()
				f := fval.asFloat()
				o.value.set(x, y, f)
				return makeNilValue()
			},
		}
	case "clear":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				fval := vm.Stack(arg_stackptr)
				f := fval.asFloat()
				o.value.clear(f)
				return makeNilValue()
			},
		}
	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
