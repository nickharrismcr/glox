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
	BATCH_TRIANGLE3
	BATCH_CIRCLE3
)

// Number of triangle-fan segments used to draw each BATCH_CIRCLE3 entry.
// The fan is built natively in Draw() (see drawCircle3 below), so raising
// this only costs Go-side work, not additional Lox-level calls.
const circle3Segments = 10

// Unit-circle direction cosines/sines for the flat (XZ-plane) disc, shared
// by every BATCH_CIRCLE3 entry so Draw() never calls math.Cos/Sin per
// segment per entry per frame.
var circle3UnitX, circle3UnitZ [circle3Segments]float32

func init() {
	for i := 0; i < circle3Segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(circle3Segments)
		circle3UnitX[i] = float32(math.Cos(angle))
		circle3UnitZ[i] = float32(math.Sin(angle))
	}
}

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

// A flat, filled, horizontal (XZ-plane) circle -- e.g. a ground shadow.
// Center.Y is the height at which the disc lies; it always faces +Y.
type CircleBatchEntry struct {
	Center rl.Vector3
	Radius float32
	Color  rl.Color
}

type DrawBatch struct {
	BatchType      BatchPrimitive
	Entries        []BatchEntry
	TrianglePoints []TriangleBatchEntry // For BATCH_TRIANGLE3 type
	Circles        []CircleBatchEntry   // For BATCH_CIRCLE3 type
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
	batch := &DrawBatch{
		BatchType:      batchType,
		Entries:        make([]BatchEntry, 0), // Pre-allocate capacity
		TrianglePoints: make([]TriangleBatchEntry, 0),
		Circles:        make([]CircleBatchEntry, 0),
		Capacity:       0,
	}

	return &BatchObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         batch,
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

	case BATCH_TRIANGLE3:
		typeName = "TRIANGLE3"
		entryCount = len(o.Value.TrianglePoints)
	case BATCH_CIRCLE3:
		typeName = "CIRCLE3"
		entryCount = len(o.Value.Circles)
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

// Add a flat, filled circle (e.g. a ground shadow) in one call, instead of
// building a triangle fan by hand and calling AddTriangle3 per segment.
func (batch *DrawBatch) AddCircle3(center *core.Vec3Object, radius float64, color *core.Vec4Object) int {
	entry := CircleBatchEntry{
		Center: rl.Vector3{
			X: float32(center.X),
			Y: float32(center.Y),
			Z: float32(center.Z),
		},
		Radius: float32(radius),
		Color: rl.Color{
			R: uint8(color.X),
			G: uint8(color.Y),
			B: uint8(color.Z),
			A: uint8(color.W),
		},
	}
	batch.Circles = append(batch.Circles, entry)
	return len(batch.Circles) - 1
}

// Update center/radius/color of a circle in one call, taking raw floats
// for the center to avoid a throwaway vec3 allocation (mirrors
// SetTriangle3Full) -- useful when animating a persistent index instead of
// clearing and re-adding every frame.
func (batch *DrawBatch) SetCircle3Full(index int, x, y, z, radius float64, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.Circles) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Circles[index] = CircleBatchEntry{
		Center: rl.Vector3{X: float32(x), Y: float32(y), Z: float32(z)},
		Radius: float32(radius),
		Color: rl.Color{
			R: uint8(color.X),
			G: uint8(color.Y),
			B: uint8(color.Z),
			A: uint8(color.W),
		},
	}
	return nil
}

func (batch *DrawBatch) SetCircle3Color(index int, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.Circles) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.Circles[index].Color = rl.Color{
		R: uint8(color.X),
		G: uint8(color.Y),
		B: uint8(color.Z),
		A: uint8(color.W),
	}
	return nil
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

func (batch *DrawBatch) SetTriangle3(index int, p1 *core.Vec3Object, p2 *core.Vec3Object, p3 *core.Vec3Object) error {
	if index < 0 || index >= len(batch.TrianglePoints) {
		return fmt.Errorf("index out of range :%d ", index)
	}
	// Preserve the color when updating the triangle points
	oldEntry := batch.TrianglePoints[index]
	batch.TrianglePoints[index] = TriangleBatchEntry{
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
		Color: oldEntry.Color,
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
	batch.Circles = batch.Circles[:0]               // Clear circles too

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
	switch batch.BatchType {
	case BATCH_TRIANGLE3:
		return len(batch.TrianglePoints)
	case BATCH_CIRCLE3:
		return len(batch.Circles)
	}

	return len(batch.Entries)
}

// drawCircle3 renders one flat, filled circle as a native triangle fan.
// Winding matches the hexagon-shadow code this replaces: each segment is
// (center, next, current), giving a face normal that points +Y.
func drawCircle3(entry CircleBatchEntry) {
	cx, cy, cz := entry.Center.X, entry.Center.Y, entry.Center.Z
	r := entry.Radius
	for i := 0; i < circle3Segments; i++ {
		j := (i + 1) % circle3Segments
		p1 := rl.Vector3{X: cx + r*circle3UnitX[i], Y: cy, Z: cz + r*circle3UnitZ[i]}
		p2 := rl.Vector3{X: cx + r*circle3UnitX[j], Y: cy, Z: cz + r*circle3UnitZ[j]}
		rl.DrawTriangle3D(entry.Center, p2, p1, entry.Color)
	}
}

// Render all entries in the batch
func (batch *DrawBatch) Draw() {
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

	case BATCH_TRIANGLE3:
		// Draw 3-point triangles
		for _, entry := range batch.TrianglePoints {
			rl.DrawTriangle3D(entry.Point1, entry.Point2, entry.Point3, entry.Color)
		}

	case BATCH_CIRCLE3:
		for _, entry := range batch.Circles {
			drawCircle3(entry)
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
	case BATCH_TRIANGLE3:

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

	case BATCH_CIRCLE3:
		for _, entry := range batch.Circles {
			dx := entry.Center.X - cameraPos.X
			dy := entry.Center.Y - cameraPos.Y
			dz := entry.Center.Z - cameraPos.Z
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

				sizeAngleOffset := float32(math.Atan(float64(entry.Radius / distance)))
				adjustedMinDot := float32(math.Cos(float64(fovRadians/2.0 + sizeAngleOffset)))

				if dotProduct >= adjustedMinDot {
					drawCircle3(entry)
				}
			}
		}

	}
}

// Set the color of a triangle in a BATCH_TRIANGLE3 batch
func (batch *DrawBatch) SetTriangle3Color(index int, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.TrianglePoints) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.TrianglePoints[index].Color = rl.Color{
		R: uint8(color.X),
		G: uint8(color.Y),
		B: uint8(color.Z),
		A: uint8(color.W),
	}
	return nil
}

// Set points (as raw floats, avoiding vec3 allocation) and color of a triangle in one call
func (batch *DrawBatch) SetTriangle3Full(index int, x1, y1, z1, x2, y2, z2, x3, y3, z3 float64, color *core.Vec4Object) error {
	if index < 0 || index >= len(batch.TrianglePoints) {
		return fmt.Errorf("index out of range: %d", index)
	}
	batch.TrianglePoints[index] = TriangleBatchEntry{
		Point1: rl.Vector3{X: float32(x1), Y: float32(y1), Z: float32(z1)},
		Point2: rl.Vector3{X: float32(x2), Y: float32(y2), Z: float32(z2)},
		Point3: rl.Vector3{X: float32(x3), Y: float32(y3), Z: float32(z3)},
		Color: rl.Color{
			R: uint8(color.X),
			G: uint8(color.Y),
			B: uint8(color.Z),
			A: uint8(color.W),
		},
	}
	return nil
}

// Get the color of a triangle in a BATCH_TRIANGLE3 batch
func (batch *DrawBatch) GetTriangle3Color(index int) (*core.Vec4Object, error) {
	if index < 0 || index >= len(batch.TrianglePoints) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	color := &batch.TrianglePoints[index].Color
	return core.MakeVec4Object(float64(color.R), float64(color.G), float64(color.B), float64(color.A)), nil
}
