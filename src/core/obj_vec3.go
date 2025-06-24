package core

import (
	"fmt"
)

type Vec3Object struct {
	X float64
	Y float64
	Z float64
}

func MakeVec3Object(x float64, y float64, z float64) *Vec3Object {

	return &Vec3Object{
		X: x,
		Y: y,
		Z: z,
	}
}

func (*Vec3Object) IsObject() {}

func (v *Vec3Object) GetType() ObjectType {

	return OBJECT_VEC3
}

func (v *Vec3Object) GetNativeType() NativeType {
	return NATIVE_VEC3
}

func (s *Vec3Object) String() string {

	return fmt.Sprintf("<Vec3 %f,%f,%f>", s.X, s.Y, s.Z)
}

func (s *Vec3Object) IsBuiltIn() bool {
	return false
}

func (s *Vec3Object) SetX(x float64) {
	s.X = x
}
func (s *Vec3Object) SetY(y float64) {
	s.Y = y
}
func (s *Vec3Object) SetZ(z float64) {
	s.Z = z
}

func (s *Vec3Object) Add(other *Vec3Object) {

	s.X += other.X
	s.Y += other.Y
	s.Z += other.Z

}
