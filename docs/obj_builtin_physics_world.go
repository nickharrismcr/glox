package builtin

import (
	"fmt"
	"glox/src/core"
	"math"
)

// ---------- Constructor (follows standard pattern, cf. BatchBuiltIn) ----------
// physics_world(min vec3, max vec3, cell_size number, gravity vec3)
func PhysicsWorldBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
	if argCount != 4 {
		vm.RunTimeError("physics_world() expects 4 arguments (min, max, cell_size, gravity)")
		return core.NIL_VALUE
	}

	minVal := vm.Stack(arg_stackptr)
	maxVal := vm.Stack(arg_stackptr + 1)
	cellSizeVal := vm.Stack(arg_stackptr + 2)
	gravityVal := vm.Stack(arg_stackptr + 3)

	if minVal.Type != core.VAL_VEC3 {
		vm.RunTimeError("physics_world() first argument must be a vec3 (bounds min)")
		return core.NIL_VALUE
	}
	if maxVal.Type != core.VAL_VEC3 {
		vm.RunTimeError("physics_world() second argument must be a vec3 (bounds max)")
		return core.NIL_VALUE
	}
	if !cellSizeVal.IsNumber() {
		vm.RunTimeError("physics_world() third argument must be a number (cell_size)")
		return core.NIL_VALUE
	}
	if gravityVal.Type != core.VAL_VEC3 {
		vm.RunTimeError("physics_world() fourth argument must be a vec3 (gravity)")
		return core.NIL_VALUE
	}

	min := minVal.Obj.(*core.Vec3Object)
	max := maxVal.Obj.(*core.Vec3Object)
	gravity := gravityVal.Obj.(*core.Vec3Object)

	worldObj := MakePhysicsWorldObject(
		PVec3{min.X, min.Y, min.Z},
		PVec3{max.X, max.Y, max.Z},
		cellSizeVal.AsFloat(),
		PVec3{gravity.X, gravity.Y, gravity.Z},
	)
	RegisterAllPhysicsWorldMethods(worldObj)
	return core.MakeObjectValue(worldObj, true)
}

// ---------- Internal simulation types ----------

// PVec3 avoids clashing with core.Vec3Object / raylib's Vector3 — this
// is purely an internal math type for the SoA simulation, not exposed
// to Lox directly (positions cross the boundary as core.Vec3Object).
type PVec3 struct{ X, Y, Z float64 }

type Material struct {
	Restitution float64
	Friction    float64
	Damping     float64
}

func combineMaterials(a, b Material) (restitution, friction float64) {
	restitution = math.Sqrt(a.Restitution * b.Restitution)
	friction = math.Sqrt(a.Friction * b.Friction)
	return
}

type Bounds struct{ Min, Max PVec3 }

type CollisionPair struct {
	A, B    int
	Normal  PVec3
	Impulse float64
}

type cellKey struct{ X, Y, Z int32 }
type pairKey struct{ A, B int }

type PhysicsWorld struct {
	posX, posY, posZ []float64
	velX, velY, velZ []float64
	radius           []float64
	materialID       []int
	active           []bool

	materials []Material

	bounds   Bounds
	gravity  PVec3
	cellSize float64
	grid     map[cellKey][]int

	collisions   []CollisionPair
	prevContacts map[pairKey]bool
	currContacts map[pairKey]bool
}

func NewPhysicsWorld(min, max PVec3, cellSize float64, gravity PVec3) *PhysicsWorld {
	return &PhysicsWorld{
		bounds:       Bounds{min, max},
		gravity:      gravity,
		cellSize:     cellSize,
		grid:         make(map[cellKey][]int),
		prevContacts: make(map[pairKey]bool),
		currContacts: make(map[pairKey]bool),
	}
}

func (w *PhysicsWorld) AddMaterial(restitution, friction, damping float64) int {
	w.materials = append(w.materials, Material{restitution, friction, damping})
	return len(w.materials) - 1
}

func (w *PhysicsWorld) Add(pos, vel PVec3, radius float64, materialID int) (int, error) {
	if materialID < 0 || materialID >= len(w.materials) {
		return 0, fmt.Errorf("invalid material id: %d", materialID)
	}
	w.posX = append(w.posX, pos.X)
	w.posY = append(w.posY, pos.Y)
	w.posZ = append(w.posZ, pos.Z)
	w.velX = append(w.velX, vel.X)
	w.velY = append(w.velY, vel.Y)
	w.velZ = append(w.velZ, vel.Z)
	w.radius = append(w.radius, radius)
	w.materialID = append(w.materialID, materialID)
	w.active = append(w.active, true)
	return len(w.posX) - 1, nil
}

func (w *PhysicsWorld) Remove(id int) error {
	if id < 0 || id >= len(w.active) {
		return fmt.Errorf("index out of range: %d", id)
	}
	w.active[id] = false
	return nil
}

func (w *PhysicsWorld) GetPosition(id int) (PVec3, error) {
	if id < 0 || id >= len(w.posX) || !w.active[id] {
		return PVec3{}, fmt.Errorf("index out of range or inactive: %d", id)
	}
	return PVec3{w.posX[id], w.posY[id], w.posZ[id]}, nil
}

func (w *PhysicsWorld) Count() int {
	n := 0
	for _, a := range w.active {
		if a {
			n++
		}
	}
	return n
}

func (w *PhysicsWorld) Collisions() []CollisionPair {
	return w.collisions
}

// ---------- Step: integration + broad phase + narrow phase ----------

func (w *PhysicsWorld) Step(dt float64) {
	w.integrate(dt)
	w.boundaryCollisions()
	w.rebuildGrid()

	w.collisions = w.collisions[:0]
	w.currContacts = make(map[pairKey]bool, len(w.currContacts))

	w.narrowPhase()

	w.prevContacts = w.currContacts
}

func (w *PhysicsWorld) integrate(dt float64) {
	for i := range w.posX {
		if !w.active[i] {
			continue
		}
		mat := w.materials[w.materialID[i]]

		w.velX[i] += w.gravity.X * dt
		w.velY[i] += w.gravity.Y * dt
		w.velZ[i] += w.gravity.Z * dt

		w.velX[i] *= mat.Damping
		w.velY[i] *= mat.Damping
		w.velZ[i] *= mat.Damping

		w.posX[i] += w.velX[i] * dt
		w.posY[i] += w.velY[i] * dt
		w.posZ[i] += w.velZ[i] * dt
	}
}

func (w *PhysicsWorld) boundaryCollisions() {
	for i := range w.posX {
		if !w.active[i] {
			continue
		}
		mat := w.materials[w.materialID[i]]
		r := w.radius[i]

		clampAxis(&w.posX[i], &w.velX[i], w.bounds.Min.X+r, w.bounds.Max.X-r, mat.Restitution)
		clampAxis(&w.posY[i], &w.velY[i], w.bounds.Min.Y+r, w.bounds.Max.Y-r, mat.Restitution)
		clampAxis(&w.posZ[i], &w.velZ[i], w.bounds.Min.Z+r, w.bounds.Max.Z-r, mat.Restitution)
	}
}

func clampAxis(pos, vel *float64, lo, hi, restitution float64) {
	if *pos < lo {
		*pos = lo
		*vel = -*vel * restitution
	} else if *pos > hi {
		*pos = hi
		*vel = -*vel * restitution
	}
}

func (w *PhysicsWorld) cellOf(i int) cellKey {
	return cellKey{
		X: int32(math.Floor(w.posX[i] / w.cellSize)),
		Y: int32(math.Floor(w.posY[i] / w.cellSize)),
		Z: int32(math.Floor(w.posZ[i] / w.cellSize)),
	}
}

func (w *PhysicsWorld) rebuildGrid() {
	for k := range w.grid {
		delete(w.grid, k)
	}
	for i := range w.posX {
		if !w.active[i] {
			continue
		}
		k := w.cellOf(i)
		w.grid[k] = append(w.grid[k], i)
	}
}

func (w *PhysicsWorld) narrowPhase() {
	checked := make(map[pairKey]bool)

	for i := range w.posX {
		if !w.active[i] {
			continue
		}
		base := w.cellOf(i)

		for dz := int32(-1); dz <= 1; dz++ {
			for dy := int32(-1); dy <= 1; dy++ {
				for dx := int32(-1); dx <= 1; dx++ {
					k := cellKey{base.X + dx, base.Y + dy, base.Z + dz}
					for _, j := range w.grid[k] {
						if j <= i || !w.active[j] {
							continue
						}
						pk := pairKey{i, j}
						if checked[pk] {
							continue
						}
						checked[pk] = true
						w.checkAndResolve(i, j, pk)
					}
				}
			}
		}
	}
}

func (w *PhysicsWorld) checkAndResolve(i, j int, pk pairKey) {
	dx := w.posX[j] - w.posX[i]
	dy := w.posY[j] - w.posY[i]
	dz := w.posZ[j] - w.posZ[i]
	distSq := dx*dx + dy*dy + dz*dz
	minDist := w.radius[i] + w.radius[j]

	if distSq >= minDist*minDist || distSq < 1e-12 {
		return
	}

	dist := math.Sqrt(distSq)
	nx, ny, nz := dx/dist, dy/dist, dz/dist

	overlap := minDist - dist
	w.posX[i] -= nx * overlap * 0.5
	w.posY[i] -= ny * overlap * 0.5
	w.posZ[i] -= nz * overlap * 0.5
	w.posX[j] += nx * overlap * 0.5
	w.posY[j] += ny * overlap * 0.5
	w.posZ[j] += nz * overlap * 0.5

	rvx := w.velX[j] - w.velX[i]
	rvy := w.velY[j] - w.velY[i]
	rvz := w.velZ[j] - w.velZ[i]
	velAlongNormal := rvx*nx + rvy*ny + rvz*nz

	restitution, _ := combineMaterials(w.materials[w.materialID[i]], w.materials[w.materialID[j]])

	if velAlongNormal < 0 {
		impulse := -(1 + restitution) * velAlongNormal * 0.5
		w.velX[i] -= impulse * nx
		w.velY[i] -= impulse * ny
		w.velZ[i] -= impulse * nz
		w.velX[j] += impulse * nx
		w.velY[j] += impulse * ny
		w.velZ[j] += impulse * nz

		w.currContacts[pk] = true
		if !w.prevContacts[pk] {
			w.collisions = append(w.collisions, CollisionPair{
				A: i, B: j,
				Normal:  PVec3{nx, ny, nz},
				Impulse: math.Abs(velAlongNormal),
			})
		}
	} else {
		w.currContacts[pk] = true
	}
}

// ---------- Native object wrapper (exact BatchObject convention) ----------

type PhysicsWorldObject struct {
	core.BuiltInObject
	Value   *PhysicsWorld
	Methods map[int]*core.BuiltInObject
}

func MakePhysicsWorldObject(min, max PVec3, cellSize float64, gravity PVec3) *PhysicsWorldObject {
	return &PhysicsWorldObject{
		BuiltInObject: core.BuiltInObject{},
		Value:         NewPhysicsWorld(min, max, cellSize, gravity),
	}
}

func (o *PhysicsWorldObject) String() string {
	return fmt.Sprintf("<PhysicsWorld [%d bodies]>", o.Value.Count())
}

func (o *PhysicsWorldObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *PhysicsWorldObject) GetNativeType() core.NativeType {
	return core.NATIVE_PHYSICS_WORLD // add this const alongside NATIVE_BATCH etc. in core/object.go
}

func (o *PhysicsWorldObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *PhysicsWorldObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *PhysicsWorldObject) IsBuiltIn() bool {
	return true
}
