package lox

import "fmt"

type FloatArray struct {
	Width, Height int
	Data          []float64
}

func (f *FloatArray) Get(x, y int) float64 {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		panic(fmt.Sprintf("Index out of bounds: (%d, %d) for array size %dx%d", x, y, f.Width, f.Height))
	}
	return f.Data[y*f.Width+x]
}

func (f *FloatArray) Set(x, y int, value float64) {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		panic(fmt.Sprintf("Index out of bounds: (%d, %d) for array size %dx%d", x, y, f.Width, f.Height))
	}
	f.Data[y*f.Width+x] = value
}

func (f *FloatArray) Clear(value float64) {
	for i := range f.Data {
		f.Data[i] = value
	}
}

type FloatArrayObject struct {
	BuiltInObject
	Value *FloatArray
}

func MakeFloatArrayObject(w int, h int) *FloatArrayObject {

	return &FloatArrayObject{
		BuiltInObject: BuiltInObject{},
		Value:         &FloatArray{Width: w, Height: h, Data: make([]float64, w*h)},
	}
}

func (o *FloatArrayObject) String() string {
	return fmt.Sprintf("<FloatArray %dx%d>", o.Value.Width, o.Value.Height)
}

func (o *FloatArrayObject) GetType() ObjectType {
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
				return makeFloatValue(o.Value.Get(x, y), false)
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
				o.Value.Set(x, y, f)
				return makeNilValue()
			},
		}
	case "clear":
		return &BuiltInObject{
			function: func(argCount int, arg_stackptr int, vm *VM) Value {
				fval := vm.stack[arg_stackptr]
				f := fval.asFloat()
				o.Value.Clear(f)
				return makeNilValue()
			},
		}
	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
