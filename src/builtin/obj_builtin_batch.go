package builtin

import (
	"fmt"
	"glox/src/core"
	"math"

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
	BATCH_TRIANGLE
	BATCH_TRIANGLE3
)

// Internal data structures
type BatchEntry struct {
	Position rl.Vector3
	Size     rl.Vector3
	Color    rl.Color
	Rotation rl.Vector3
}

// For triangles defined by three 3D points
type TriangleBatchEntry struct {
	Point1 rl.Vector3
	Point2 rl.Vector3
	Point3 rl.Vector3
	Color  rl.Color
}

type DrawBatch struct {
	BatchType      BatchPrimitive
	Entries        []BatchEntry
	TrianglePoints []TriangleBatchEntry // For BATCH_TRIANGLE3 type
	Capacity       int
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
			BatchType:      batchType,
			Entries:        make([]BatchEntry, 0, 1000), // Pre-allocate capacity
			TrianglePoints: make([]TriangleBatchEntry, 0, 1000),
			Capacity:       1000,
		},
	}
}

// Standard interface implementations
func (o *BatchObject) String() string {
	var typeName string
	var entryCount int

	switch o.Value.BatchType {
	case BATCH_CUBE:
		typeName = "CUBE"
		entryCount = len(o.Value.Entries)
	case BATCH_SPHERE:
		typeName = "SPHERE"
		entryCount = len(o.Value.Entries)
	case BATCH_PLANE:
		typeName = "PLANE"
		entryCount = len(o.Value.Entries)
	case BATCH_TRIANGLE:
		typeName = "TRIANGLE"
		entryCount = len(o.Value.Entries)
	case BATCH_TRIANGLE3:
		typeName = "TRIANGLE3"
		entryCount = len(o.Value.TrianglePoints)
	default:
		typeName = "UNKNOWN"
		entryCount = len(o.Value.Entries)
	}
	return fmt.Sprintf("<Batch %s [%d entries]>", typeName, entryCount)
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

// Add triangle with three specific 3D points
func (batch *DrawBatch) AddTriangle3(p1 *core.Vec3Object, p2 *core.Vec3Object, p3 *core.Vec3Object, color *core.Vec4Object) int {
	entry := TriangleBatchEntry{
		Point1: rl.Vector3{
			X: float32(p1.X),
			Y: float32(p1.Y),
			Z: float32(p1.Z),
		},
		Point2: rl.Vector3{
			X: float32(p2.X),
			Y: float32(p2.Y),
			Z: float32(p2.Z),
		},
		Point3: rl.Vector3{
			X: float32(p3.X),
			Y: float32(p3.Y),
			Z: float32(p3.Z),
		},
		Color: rl.Color{
			R: uint8(color.X),
			G: uint8(color.Y),
			B: uint8(color.Z),
			A: uint8(color.W),
		},
	}
	batch.TrianglePoints = append(batch.TrianglePoints, entry)
	return len(batch.TrianglePoints) - 1
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
	batch.Entries = batch.Entries[:0]               // Keep capacity, reset length
	batch.TrianglePoints = batch.TrianglePoints[:0] // Clear triangle points too
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
	if batch.BatchType == BATCH_TRIANGLE3 {
		return len(batch.TrianglePoints)
	}
	return len(batch.Entries)
}

// Render all entries in the batch
func (batch *DrawBatch) Draw() {
	if batch.BatchType == BATCH_TRIANGLE3 {
		// Draw 3-point triangles
		for _, entry := range batch.TrianglePoints {
			rl.DrawTriangle3D(entry.Point1, entry.Point2, entry.Point3, entry.Color)
		}
		return
	}

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

// Simple distance-based culling for better performance
func (batch *DrawBatch) DrawWithCulling(cameraPos rl.Vector3, maxDistance float32) {
	maxDistanceSq := maxDistance * maxDistance

	if batch.BatchType == BATCH_TRIANGLE3 {
		// Draw 3-point triangles with distance culling
		for _, entry := range batch.TrianglePoints {
			// Calculate center point of triangle for distance check
			centerX := (entry.Point1.X + entry.Point2.X + entry.Point3.X) / 3.0
			centerY := (entry.Point1.Y + entry.Point2.Y + entry.Point3.Y) / 3.0
			centerZ := (entry.Point1.Z + entry.Point2.Z + entry.Point3.Z) / 3.0

			dx := centerX - cameraPos.X
			dy := centerY - cameraPos.Y
			dz := centerZ - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq <= maxDistanceSq {
				rl.DrawTriangle3D(entry.Point1, entry.Point2, entry.Point3, entry.Color)
			}
		}
		return
	}

	if len(batch.Entries) == 0 {
		return
	}

	// Batch render based on type with distance culling
	switch batch.BatchType {
	case BATCH_CUBE:
		for _, entry := range batch.Entries {
			// Calculate distance squared (avoid expensive sqrt)
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq <= maxDistanceSq {
				rl.DrawCube(entry.Position, entry.Size.X, entry.Size.Y, entry.Size.Z, entry.Color)
			}
		}
	case BATCH_SPHERE:
		for _, entry := range batch.Entries {
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq <= maxDistanceSq {
				rl.DrawSphere(entry.Position, entry.Size.X, entry.Color)
			}
		}
	case BATCH_PLANE:
		for _, entry := range batch.Entries {
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq <= maxDistanceSq {
				size := rl.Vector2{
					X: entry.Size.X,
					Y: entry.Size.Z,
				}
				rl.DrawPlane(entry.Position, size, entry.Color)
			}
		}
	}
}

// Improved culling with camera direction (eliminates objects behind camera)
func (batch *DrawBatch) DrawWithDirectionalCulling(cameraPos rl.Vector3, cameraForward rl.Vector3, maxDistance float32, fovAngleDegrees float32) {
	maxDistanceSq := maxDistance * maxDistance
	// Convert FOV to radians and get cosine for dot product comparison
	// Add some padding to the FOV to prevent edge flickering
	paddedFOV := fovAngleDegrees + 10.0 // Add 10 degrees padding
	fovRadians := paddedFOV * 3.14159 / 180.0

	if batch.BatchType == BATCH_TRIANGLE3 {
		// Draw 3-point triangles with directional culling
		for _, entry := range batch.TrianglePoints {
			// Calculate center point of triangle for culling calculations
			centerX := (entry.Point1.X + entry.Point2.X + entry.Point3.X) / 3.0
			centerY := (entry.Point1.Y + entry.Point2.Y + entry.Point3.Y) / 3.0
			centerZ := (entry.Point1.Z + entry.Point2.Z + entry.Point3.Z) / 3.0

			// Calculate distance squared
			dx := centerX - cameraPos.X
			dy := centerY - cameraPos.Y
			dz := centerZ - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			// Early distance check
			if distanceSq > maxDistanceSq {
				continue
			}

			// Calculate distance and direction to triangle center
			distance := float32(math.Sqrt(float64(distanceSq)))
			if distance > 0.001 { // Avoid division by zero
				objDirX := dx / distance
				objDirY := dy / distance
				objDirZ := dz / distance

				// Dot product with camera forward vector
				dotProduct := objDirX*cameraForward.X + objDirY*cameraForward.Y + objDirZ*cameraForward.Z

				// Estimate triangle size for visibility calculations
				// Calculate the maximum distance between triangle points
				d12 := float32(math.Sqrt(float64((entry.Point2.X-entry.Point1.X)*(entry.Point2.X-entry.Point1.X) +
					(entry.Point2.Y-entry.Point1.Y)*(entry.Point2.Y-entry.Point1.Y) +
					(entry.Point2.Z-entry.Point1.Z)*(entry.Point2.Z-entry.Point1.Z))))
				d13 := float32(math.Sqrt(float64((entry.Point3.X-entry.Point1.X)*(entry.Point3.X-entry.Point1.X) +
					(entry.Point3.Y-entry.Point1.Y)*(entry.Point3.Y-entry.Point1.Y) +
					(entry.Point3.Z-entry.Point1.Z)*(entry.Point3.Z-entry.Point1.Z))))
				d23 := float32(math.Sqrt(float64((entry.Point3.X-entry.Point2.X)*(entry.Point3.X-entry.Point2.X) +
					(entry.Point3.Y-entry.Point2.Y)*(entry.Point3.Y-entry.Point2.Y) +
					(entry.Point3.Z-entry.Point2.Z)*(entry.Point3.Z-entry.Point2.Z))))

				maxEdge := d12
				if d13 > maxEdge {
					maxEdge = d13
				}
				if d23 > maxEdge {
					maxEdge = d23
				}

				objectRadius := maxEdge / 2.0                                           // Use half the maximum edge as "radius"
				sizeAngleOffset := float32(math.Atan(float64(objectRadius / distance))) // Angular size
				adjustedMinDot := float32(math.Cos(float64(fovRadians/2.0 + sizeAngleOffset)))

				// Check if triangle is within FOV cone
				if dotProduct >= adjustedMinDot {
					rl.DrawTriangle3D(entry.Point1, entry.Point2, entry.Point3, entry.Color)
				}
			}
		}
		return
	}

	if len(batch.Entries) == 0 {
		return
	}

	// Batch render based on type with directional culling
	switch batch.BatchType {
	case BATCH_CUBE:
		for _, entry := range batch.Entries {
			// Calculate distance squared
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			// Early distance check
			if distanceSq > maxDistanceSq {
				continue
			}

			// Calculate distance and direction to object
			distance := float32(math.Sqrt(float64(distanceSq)))
			if distance > 0.001 { // Avoid division by zero
				objDirX := dx / distance
				objDirY := dy / distance
				objDirZ := dz / distance

				// Dot product with camera forward vector
				dotProduct := objDirX*cameraForward.X + objDirY*cameraForward.Y + objDirZ*cameraForward.Z

				// Account for object size - larger objects should be visible from wider angles
				objectRadius := (entry.Size.X + entry.Size.Y + entry.Size.Z) / 3.0      // Average size as radius
				sizeAngleOffset := float32(math.Atan(float64(objectRadius / distance))) // Angular size
				adjustedMinDot := float32(math.Cos(float64(fovRadians/2.0 + sizeAngleOffset)))

				// Check if object is within FOV cone (use the more permissive threshold)
				if dotProduct >= adjustedMinDot {
					rl.DrawCube(entry.Position, entry.Size.X, entry.Size.Y, entry.Size.Z, entry.Color)
				}
			}
		}
	case BATCH_SPHERE:
		for _, entry := range batch.Entries {
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq > maxDistanceSq {
				continue
			}

			distance := float32(math.Sqrt(float64(distanceSq)))
			if distance > 0.001 {
				objDirX := dx / distance
				objDirY := dy / distance
				objDirZ := dz / distance

				dotProduct := objDirX*cameraForward.X + objDirY*cameraForward.Y + objDirZ*cameraForward.Z

				// Account for sphere radius
				objectRadius := entry.Size.X // Sphere radius
				sizeAngleOffset := float32(math.Atan(float64(objectRadius / distance)))
				adjustedMinDot := float32(math.Cos(float64(fovRadians/2.0 + sizeAngleOffset)))

				if dotProduct >= adjustedMinDot {
					rl.DrawSphere(entry.Position, entry.Size.X, entry.Color)
				}
			}
		}
	case BATCH_PLANE:
		for _, entry := range batch.Entries {
			dx := entry.Position.X - cameraPos.X
			dy := entry.Position.Y - cameraPos.Y
			dz := entry.Position.Z - cameraPos.Z
			distanceSq := dx*dx + dy*dy + dz*dz

			if distanceSq > maxDistanceSq {
				continue
			}

			distance := float32(math.Sqrt(float64(distanceSq)))
			if distance > 0.001 {
				objDirX := dx / distance
				objDirY := dy / distance
				objDirZ := dz / distance

				dotProduct := objDirX*cameraForward.X + objDirY*cameraForward.Y + objDirZ*cameraForward.Z

				// Account for plane size (use max of X and Z dimensions)
				objectRadius := entry.Size.X
				if entry.Size.Z > objectRadius {
					objectRadius = entry.Size.Z
				}
				sizeAngleOffset := float32(math.Atan(float64(objectRadius / distance)))
				adjustedMinDot := float32(math.Cos(float64(fovRadians/2.0 + sizeAngleOffset)))

				if dotProduct >= adjustedMinDot {
					size := rl.Vector2{
						X: entry.Size.X,
						Y: entry.Size.Z,
					}
					rl.DrawPlane(entry.Position, size, entry.Color)
				}
			}
		}
	}
}
