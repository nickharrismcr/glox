package core

import (
	"fmt"
)

type Vec2Object struct {
	X float64
	Y float64
}

func MakeVec2Object(x float64, y float64) *Vec2Object {

	return &Vec2Object{
		X: x,
		Y: y,
	}
}

func (*Vec2Object) IsObject() {}

func (v *Vec2Object) GetType() ObjectType {

	return OBJECT_VEC2
}

func (s *Vec2Object) String() string {

	return fmt.Sprintf("<Vec2 %f,%f>", s.X, s.Y)
}

func (s *Vec2Object) IsBuiltIn() bool {
	return false
}

func (s *Vec2Object) SetX(x float64) {
	s.X = x
}
func (s *Vec2Object) SetY(y float64) {
	s.Y = y
}
