package builtin

import (
	"fmt"
	"glox/src/core"
)

func FloatArrayBuiltin(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	widthval := vm.Stack(arg_stackptr)
	heightval := vm.Stack(arg_stackptr + 1)
	if argCount != 2 {
		vm.RunTimeError("Invalid argument count to float_array.")
		return core.MakeNilValue()
	}
	if !widthval.IsInt() || !heightval.IsInt() {
		vm.RunTimeError("float_array arguments must be integers")
		return core.MakeNilValue()
	}
	width := widthval.Int
	height := heightval.Int
	floatArrObj := MakeFloatArrayObject(width, height)
	RegisterAllFloatArrayMethods(floatArrObj)
	return core.MakeObjectValue(floatArrObj, false)
}

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
	core.BuiltInObject
	Value   *FloatArray
	Methods map[string]*core.BuiltInObject
}

func MakeFloatArrayObject(w int, h int) *FloatArrayObject {

	return &FloatArrayObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         &FloatArray{Width: w, Height: h, Data: make([]float64, w*h)},
	}
}

func AsFloatArray(v core.Value) *FloatArrayObject {

	return v.Obj.(*FloatArrayObject)
}

func (o *FloatArrayObject) String() string {
	return fmt.Sprintf("<FloatArray %dx%d>", o.Value.Width, o.Value.Height)
}

func (o *FloatArrayObject) GetType() core.ObjectType {
	return core.OBJECT_FLOAT_ARRAY
}

func (o *FloatArrayObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[string]*core.BuiltInObject)
	}
	o.Methods[name] = method
}

func (t *FloatArrayObject) IsBuiltIn() bool {
	return true
}
