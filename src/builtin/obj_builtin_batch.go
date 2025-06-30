package builtin

import (
	"fmt"
	"glox/src/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Constructor function (follows standard pattern)
func BatchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 1 {
		vm.RunTimeError("batch() expects 1 argument")
		return core.NIL_VALUE
	}

	batchTypeVal := vm.Stack(arg_stackptr)
	var batchType string

	if batchTypeVal.IsStringObject() {
		// Accept string literals (legacy support)
		batchType = batchTypeVal.AsString().Get()
	} else if batchTypeVal.IsInt() {
		// Accept integer constants (new preferred method)
		switch batchTypeVal.AsInt() {
		case 0:
			batchType = "cube"
		case 1:
			batchType = "sphere"
		case 2:
			batchType = "plane"
		default:
			vm.RunTimeError("Invalid batch type constant")
			return core.NIL_VALUE
		}
	} else {
		vm.RunTimeError("batch() argument must be a string or batch type constant")
		return core.NIL_VALUE
	}

	batchObj := MakeBatchObject(batchType)
	RegisterAllBatchMethods(batchObj)
	return core.MakeObjectValue(batchObj, true)
}

// Internal data structures
type BatchEntry struct {
	Position core.Vec3Object
	Size     core.Vec3Object
	Color    core.Vec4Object
	Rotation core.Vec3Object
}

type DrawBatch struct {
	BatchType string
	Entries   []BatchEntry
	Capacity  int
}

// Main object (follows standard pattern)
type BatchObject struct {
	core.BuiltInObject
	Value   *DrawBatch
	Methods map[int]*core.BuiltInObject
}

// Constructor
func MakeBatchObject(batchType string) *BatchObject {
	return &BatchObject{
		BuiltInObject: core.BuiltInObject{},
		Value: &DrawBatch{
			BatchType: batchType,
			Entries:   make([]BatchEntry, 0, 1000), // Pre-allocate capacity
			Capacity:  1000,
		},
	}
}

// Standard interface implementations
func (o *BatchObject) String() string {
	return fmt.Sprintf("<Batch %s [%d entries]>", o.Value.BatchType, len(o.Value.Entries))
}

func (o *BatchObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *BatchObject) GetNativeType() core.NativeType {
	return core.NATIVE_BATCH
}

func (o *BatchObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *BatchObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *BatchObject) IsBuiltIn() bool {
	return true
}

// Utility functions
func IsBatchObject(v core.Value) bool {
	_, ok := v.Obj.(*BatchObject)
	return ok
}

func AsBatch(v core.Value) *BatchObject {
	return v.Obj.(*BatchObject)
}

// Core batch operations (internal methods)
func (batch *DrawBatch) Add(pos *core.Vec3Object, size *core.Vec3Object, color *core.Vec4Object) int {
	entry := BatchEntry{
		Position: *pos,
		Size:     *size,
		Color:    *color,
		Rotation: *core.MakeVec3Object(0, 0, 0), // Default no rotation
	}
	batch.Entries = append(batch.Entries, entry)
	return len(batch.Entries) - 1
}

func (batch *DrawBatch) SetPosition(index int, pos *core.Vec3Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Position = *pos
	return nil
}

func (batch *DrawBatch) SetColor(index int, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Color = *color
	return nil
}

func (batch *DrawBatch) SetSize(index int, size *core.Vec3Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Size = *size
	return nil
}

func (batch *DrawBatch) GetPosition(index int) (*core.Vec3Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	return &batch.Entries[index].Position, nil
}

func (batch *DrawBatch) GetColor(index int) (*core.Vec4Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	return &batch.Entries[index].Color, nil
}

func (batch *DrawBatch) GetSize(index int) (*core.Vec3Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	return &batch.Entries[index].Size, nil
}

func (batch *DrawBatch) IsValidIndex(index int) bool {
	return index >= 0 && index < len(batch.Entries)
}

func (batch *DrawBatch) Clear() {
	batch.Entries = batch.Entries[:0] // Keep capacity, reset length
}

func (batch *DrawBatch) Reserve(capacity int) {
	if capacity > len(batch.Entries) {
		newEntries := make([]BatchEntry, len(batch.Entries), capacity)
		copy(newEntries, batch.Entries)
		batch.Entries = newEntries
		batch.Capacity = capacity
	}
}

func (batch *DrawBatch) Count() int {
	return len(batch.Entries)
}

// Render all entries in the batch
func (batch *DrawBatch) Draw() {
	if len(batch.Entries) == 0 {
		return
	}

	// Batch render based on type
	switch batch.BatchType {
	case "cube":
		for _, entry := range batch.Entries {
			pos := rl.Vector3{
				X: float32(entry.Position.X),
				Y: float32(entry.Position.Y),
				Z: float32(entry.Position.Z),
			}
			color := rl.Color{
				R: uint8(entry.Color.X),
				G: uint8(entry.Color.Y),
				B: uint8(entry.Color.Z),
				A: uint8(entry.Color.W),
			}

			rl.DrawCube(pos, float32(entry.Size.X), float32(entry.Size.Y), float32(entry.Size.Z), color)
		}
	case "sphere":
		for _, entry := range batch.Entries {
			pos := rl.Vector3{
				X: float32(entry.Position.X),
				Y: float32(entry.Position.Y),
				Z: float32(entry.Position.Z),
			}
			color := rl.Color{
				R: uint8(entry.Color.X),
				G: uint8(entry.Color.Y),
				B: uint8(entry.Color.Z),
				A: uint8(entry.Color.W),
			}

			// Use X component of size as radius
			rl.DrawSphere(pos, float32(entry.Size.X), color)
		}
	case "plane":
		for _, entry := range batch.Entries {
			pos := rl.Vector3{
				X: float32(entry.Position.X),
				Y: float32(entry.Position.Y),
				Z: float32(entry.Position.Z),
			}
			size := rl.Vector2{
				X: float32(entry.Size.X),
				Y: float32(entry.Size.Z), // Use X and Z for plane dimensions
			}
			color := rl.Color{
				R: uint8(entry.Color.X),
				G: uint8(entry.Color.Y),
				B: uint8(entry.Color.Z),
				A: uint8(entry.Color.W),
			}

			rl.DrawPlane(pos, size, color)
		}
	}
}
