package core

import (
	"fmt"
)

type Vec4Object struct {
	X float64
	Y float64
	Z float64
	W float64
}

func MakeVec4Object(x float64, y float64, z float64, w float64) *Vec4Object {

	return &Vec4Object{
		X: x,
		Y: y,
		Z: z,
		W: w,
	}
}

func (v *Vec4Object) IsObject() {}

func (v *Vec4Object) GetType() ObjectType {

	return OBJECT_VEC4
}

func (v *Vec4Object) GetNativeType() NativeType {
	return NATIVE_VEC4
}

func (s *Vec4Object) String() string {

	return fmt.Sprintf("<Vec4 %f,%f,%f,%f>", s.X, s.Y, s.Z, s.W)
}

func (s *Vec4Object) IsBuiltIn() bool {
	return false
}

func (s *Vec4Object) SetX(x float64) {
	s.X = x
}
func (s *Vec4Object) SetY(y float64) {
	s.Y = y
}
func (s *Vec4Object) SetZ(z float64) {
	s.Z = z
}
func (s *Vec4Object) SetW(w float64) {
	s.W = w
}
func (s *Vec4Object) Add(other *Vec4Object) {

	s.X += other.X
	s.Y += other.Y
	s.Z += other.Z
	s.W += other.W

}
