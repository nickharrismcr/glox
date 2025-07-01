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
	var batchType BatchPrimitive

	if batchTypeVal.IsInt() {
		batchType = BatchPrimitive(batchTypeVal.AsInt())
	} else {
		vm.RunTimeError("batch() argument must be a batch type constant")
		return core.NIL_VALUE
	}

	batchObj := MakeBatchObject(batchType)
	RegisterAllBatchMethods(batchObj)
	return core.MakeObjectValue(batchObj, true)
}

type BatchPrimitive int

const (
	BATCH_CUBE BatchPrimitive = iota
	BATCH_SPHERE
	BATCH_PLANE
)

// Internal data structures
type BatchEntry struct {
	Position rl.Vector3
	Size     rl.Vector3
	Color    rl.Color
	Rotation rl.Vector3
}

type DrawBatch struct {
	BatchType BatchPrimitive
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
func MakeBatchObject(batchType BatchPrimitive) *BatchObject {
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
	var typeName string
	switch o.Value.BatchType {
	case BATCH_CUBE:
		typeName = "CUBE"
	case BATCH_SPHERE:
		typeName = "SPHERE"
	case BATCH_PLANE:
		typeName = "PLANE"
	default:
		typeName = "UNKNOWN"
	}
	return fmt.Sprintf("<Batch %s [%d entries]>", typeName, len(o.Value.Entries))
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
		Position: rl.Vector3{
			X: float32(pos.X),
			Y: float32(pos.Y),
			Z: float32(pos.Z),
		},
		Size: rl.Vector3{
			X: float32(size.X),
			Y: float32(size.Y),
			Z: float32(size.Z),
		},
		Color: rl.Color{
			R: uint8(color.X),
			G: uint8(color.Y),
			B: uint8(color.Z),
			A: uint8(color.W),
		},
		Rotation: rl.Vector3{X: 0, Y: 0, Z: 0}, // Default no rotation
	}
	batch.Entries = append(batch.Entries, entry)
	return len(batch.Entries) - 1
}

func (batch *DrawBatch) SetPosition(index int, pos *core.Vec3Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Position = rl.Vector3{
		X: float32(pos.X),
		Y: float32(pos.Y),
		Z: float32(pos.Z),
	}
	return nil
}

func (batch *DrawBatch) SetColor(index int, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Color = rl.Color{
		R: uint8(color.X),
		G: uint8(color.Y),
		B: uint8(color.Z),
		A: uint8(color.W),
	}
	return nil
}

func (batch *DrawBatch) SetSize(index int, size *core.Vec3Object) error {
	if index < 0 || index >= len(batch.Entries) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Entries[index].Size = rl.Vector3{
		X: float32(size.X),
		Y: float32(size.Y),
		Z: float32(size.Z),
	}
	return nil
}

func (batch *DrawBatch) GetPosition(index int) (*core.Vec3Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	pos := &batch.Entries[index].Position
	return core.MakeVec3Object(float64(pos.X), float64(pos.Y), float64(pos.Z)), nil
}

func (batch *DrawBatch) GetColor(index int) (*core.Vec4Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	color := &batch.Entries[index].Color
	return core.MakeVec4Object(float64(color.R), float64(color.G), float64(color.B), float64(color.A)), nil
}

func (batch *DrawBatch) GetSize(index int) (*core.Vec3Object, error) {
	if index < 0 || index >= len(batch.Entries) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	size := &batch.Entries[index].Size
	return core.MakeVec3Object(float64(size.X), float64(size.Y), float64(size.Z)), nil
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
	case BATCH_CUBE:
		for _, entry := range batch.Entries {
			rl.DrawCube(entry.Position, entry.Size.X, entry.Size.Y, entry.Size.Z, entry.Color)
		}
	case BATCH_SPHERE:
		for _, entry := range batch.Entries {
			// Use X component of size as radius
			rl.DrawSphere(entry.Position, entry.Size.X, entry.Color)
		}
	case BATCH_PLANE:
		for _, entry := range batch.Entries {
			size := rl.Vector2{
				X: entry.Size.X,
				Y: entry.Size.Z, // Use X and Z for plane dimensions
			}
			rl.DrawPlane(entry.Position, size, entry.Color)
		}
	}
}
