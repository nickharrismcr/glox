package core

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
	case "width":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				return MakeIntValue(o.Value.Width, true)
			},
		}
	case "height":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				return MakeIntValue(o.Value.Height, true)
			},
		}
	case "get":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				x := xval.AsInt()
				y := yval.AsInt()
				return MakeFloatValue(o.Value.Get(x, y), false)
			},
		}
	case "set":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				xval := vm.Stack(arg_stackptr)
				yval := vm.Stack(arg_stackptr + 1)
				fval := vm.Stack(arg_stackptr + 2)
				x := xval.AsInt()
				y := yval.AsInt()
				f := fval.AsFloat()
				o.Value.Set(x, y, f)
				return MakeNilValue()
			},
		}
	case "clear":
		return &BuiltInObject{
			Function: func(argCount int, arg_stackptr int, vm VMContext) Value {
				fval := vm.Stack(arg_stackptr)
				f := fval.AsFloat()
				o.Value.Clear(f)
				return MakeNilValue()
			},
		}
	default:
		return nil
	}
}

//-------------------------------------------------------------------------------------------
