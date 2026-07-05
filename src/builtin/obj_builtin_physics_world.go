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

func rotateVec(v, axis PVec3, angleDeg float64) PVec3 {
	length := math.Sqrt(axis.X*axis.X + axis.Y*axis.Y + axis.Z*axis.Z)
	if length < 1e-12 {
		return v
	}
	axis = PVec3{axis.X / length, axis.Y / length, axis.Z / length}

	rad := angleDeg * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	dot := v.X*axis.X + v.Y*axis.Y + v.Z*axis.Z
	cross := PVec3{
		axis.Y*v.Z - axis.Z*v.Y,
		axis.Z*v.X - axis.X*v.Z,
		axis.X*v.Y - axis.Y*v.X,
	}
	return PVec3{
		v.X*cosA + cross.X*sinA + axis.X*dot*(1-cosA),
		v.Y*cosA + cross.Y*sinA + axis.Y*dot*(1-cosA),
		v.Z*cosA + cross.Z*sinA + axis.Z*dot*(1-cosA),
	}
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ShapeType distinguishes the collision volume attached to a body.
type ShapeType uint8

const (
	ShapeSphere ShapeType = iota
	ShapeBox
)

// Shape describes a body's collision volume. For ShapeSphere, only
// Extent.X (radius) is meaningful. For ShapeBox, Extent holds
// half-extents per axis, and Axis/Angle orient it in world space (Angle
// == 0 means axis-aligned). ShapeBox is only ever created by
// AddStaticBox -- there is no dynamic box constructor -- so every box in
// the simulation is static.
type Shape struct {
	Type   ShapeType
	Extent PVec3
	Axis   PVec3
	Angle  float64 // degrees
}

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
	shapes           []Shape
	materialID       []int
	active           []bool
	static           []bool

	// staticIDs caches the ids of static bodies as they're created, so the
	// dynamic-vs-static collision pass (resolveStaticPairs) doesn't need to
	// rescan the whole static[] slice every step. Static bodies bypass the
	// grid entirely rather than being inserted into it -- a single grid
	// cell can't represent a large platform's true extent, and multi-cell
	// insertion would cost proportional to a static body's size every
	// frame despite it never moving. See NOTES_box_physics.md section 2.
	staticIDs []int

	materials []Material

	bounds   Bounds
	gravity  PVec3
	cellSize float64

	// grid buckets by cell; usedCells lists the keys populated during the
	// last rebuildGrid() so the next call can clear just those (re-slicing
	// each bucket to length 0, keeping its backing array) instead of
	// deleting every map entry and reallocating a fresh slice on next use.
	grid      map[cellKey][]int
	usedCells []cellKey

	collisions []CollisionPair

	// contactSets is a ping-pong pair of persistent maps: each Step(), the
	// buffer that was "prev" two frames ago becomes this frame's "curr"
	// (cleared in place via the builtin clear(), not reallocated). This
	// avoids the fresh make(map[pairKey]bool) that used to run every step.
	contactSets [2]map[pairKey]bool
	currIdx     int
}

func NewPhysicsWorld(min, max PVec3, cellSize float64, gravity PVec3) *PhysicsWorld {
	return &PhysicsWorld{
		bounds:   Bounds{min, max},
		gravity:  gravity,
		cellSize: cellSize,
		grid:     make(map[cellKey][]int),
		contactSets: [2]map[pairKey]bool{
			make(map[pairKey]bool),
			make(map[pairKey]bool),
		},
	}
}

func (w *PhysicsWorld) AddMaterial(restitution, friction, damping float64) int {
	w.materials = append(w.materials, Material{restitution, friction, damping})
	return len(w.materials) - 1
}

// appendBody is the common tail of Add/AddStaticBox: push one entry onto
// every parallel SoA slice, keeping them in lockstep. Static bodies are
// also recorded in staticIDs for the ungridded dynamic-vs-static pass in
// Step().
func (w *PhysicsWorld) appendBody(pos, vel PVec3, shape Shape, materialID int, static bool) int {
	w.posX = append(w.posX, pos.X)
	w.posY = append(w.posY, pos.Y)
	w.posZ = append(w.posZ, pos.Z)
	w.velX = append(w.velX, vel.X)
	w.velY = append(w.velY, vel.Y)
	w.velZ = append(w.velZ, vel.Z)
	w.shapes = append(w.shapes, shape)
	w.materialID = append(w.materialID, materialID)
	w.active = append(w.active, true)
	w.static = append(w.static, static)

	id := len(w.posX) - 1
	if static {
		w.staticIDs = append(w.staticIDs, id)
	}
	return id
}

func (w *PhysicsWorld) Add(pos, vel PVec3, radius float64, materialID int) (int, error) {
	if materialID < 0 || materialID >= len(w.materials) {
		return 0, fmt.Errorf("invalid material id: %d", materialID)
	}
	shape := Shape{Type: ShapeSphere, Extent: PVec3{radius, 0, 0}}
	return w.appendBody(pos, vel, shape, materialID, false), nil
}

// AddStaticBox creates a fixed, optionally rotated box (ramps, shelves,
// platforms). Orientation is set once here and never updated -- no
// velocity argument, signalling fixed-ness without inspecting a flag.
func (w *PhysicsWorld) AddStaticBox(pos, halfExtents, axis PVec3, angle float64, materialID int) (int, error) {
	if materialID < 0 || materialID >= len(w.materials) {
		return 0, fmt.Errorf("invalid material id: %d", materialID)
	}
	shape := Shape{Type: ShapeBox, Extent: halfExtents, Axis: axis, Angle: angle}
	return w.appendBody(pos, PVec3{}, shape, materialID, true), nil
}

// BoxTransform is what get_box_transform() exposes back to Lox, so a
// rendered box always matches exactly what physics collided against
// (see NOTES_box_physics.md section 4 -- single source of truth).
type BoxTransform struct {
	Pos         PVec3
	HalfExtents PVec3
	Axis        PVec3
	Angle       float64
}

func (w *PhysicsWorld) GetBoxTransform(id int) (BoxTransform, error) {
	if id < 0 || id >= len(w.posX) || !w.active[id] {
		return BoxTransform{}, fmt.Errorf("index out of range or inactive: %d", id)
	}
	if w.shapes[id].Type != ShapeBox {
		return BoxTransform{}, fmt.Errorf("body %d is not a box", id)
	}
	s := w.shapes[id]
	return BoxTransform{
		Pos:         PVec3{w.posX[id], w.posY[id], w.posZ[id]},
		HalfExtents: s.Extent,
		Axis:        s.Axis,
		Angle:       s.Angle,
	}, nil
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

// AddImpulse applies an instantaneous velocity change to a single body.
// This is the primitive Lox uses for explosion forces: the distance
// check, falloff curve, and "which bodies are nearby" loop all stay in
// Lox, which computes one impulse vector per affected body and calls
// this once per body. No mass division — matches the equal-mass
// assumption used in checkAndResolve.
func (w *PhysicsWorld) AddImpulse(id int, impulse PVec3) error {
	if id < 0 || id >= len(w.posX) || !w.active[id] {
		return fmt.Errorf("index out of range or inactive: %d", id)
	}
	w.velX[id] += impulse.X
	w.velY[id] += impulse.Y
	w.velZ[id] += impulse.Z
	return nil
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

	// Swap ping-pong buffers: the map that was "prev" becomes this frame's
	// "curr", cleared in place instead of allocated fresh.
	w.currIdx = 1 - w.currIdx
	clear(w.contactSets[w.currIdx])

	w.narrowPhase()
	w.resolveStaticPairs()
}

func (w *PhysicsWorld) integrate(dt float64) {
	for i := range w.posX {
		if !w.active[i] || w.static[i] {
			continue // static bodies never move
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
		if !w.active[i] || w.static[i] {
			continue // a fixed platform shouldn't bounce off the world bounds
		}
		mat := w.materials[w.materialID[i]]
		// Every dynamic body is a sphere -- ShapeBox is only ever created
		// by AddStaticBox, and statics are skipped above -- so Extent.X
		// (radius) alone is the correct per-axis bound.
		r := w.shapes[i].Extent.X

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
	// Clear only the cells touched last frame, re-slicing each bucket to
	// length 0 so its backing array is kept (and reused below) instead of
	// deleting the map entry and forcing a fresh slice allocation on the
	// next insert.
	for _, k := range w.usedCells {
		w.grid[k] = w.grid[k][:0]
	}
	w.usedCells = w.usedCells[:0]

	for i := range w.posX {
		// Static bodies are never inserted -- see staticIDs comment above.
		if !w.active[i] || w.static[i] {
			continue
		}
		k := w.cellOf(i)
		if len(w.grid[k]) == 0 {
			w.usedCells = append(w.usedCells, k)
		}
		w.grid[k] = append(w.grid[k], i)
	}
}

// narrowPhase visits, for each dynamic body i, the 27 cells around i's own
// cell. Those 27 (dx,dy,dz) offsets are all distinct absolute cell
// coordinates, so a given neighbor cell is visited at most once per i --
// no pair can be found twice within one i's scan. Combined with the
// `j <= i` guard (which only ever looks for the higher-indexed half of a
// pair), every unordered dynamic-dynamic pair {a,b} is discovered exactly
// once overall, so no separate `checked` dedup set is needed here.
//
// Static bodies never appear here at all (as i or via the grid, which
// never contains them) -- they're handled entirely by resolveStaticPairs.
func (w *PhysicsWorld) narrowPhase() {
	for i := range w.posX {
		if !w.active[i] || w.static[i] {
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
						w.checkAndResolve(i, j, pairKey{i, j})
					}
				}
			}
		}
	}
}

// resolveStaticPairs checks every active dynamic body against every static
// body directly, bypassing the grid entirely. Static body counts (walls,
// ramps, platforms) are expected to stay small, so this O(dynamic x
// static) pass is cheap and gives statics an exact membership test instead
// of an approximate, bounding-sphere-sized grid cell.
func (w *PhysicsWorld) resolveStaticPairs() {
	for i := range w.posX {
		if !w.active[i] || w.static[i] {
			continue
		}
		for _, sid := range w.staticIDs {
			if !w.active[sid] {
				continue
			}
			a, b := i, sid
			if a > b {
				a, b = b, a
			}
			w.checkAndResolve(a, b, pairKey{a, b})
		}
	}
}

func (w *PhysicsWorld) checkAndResolve(i, j int, pk pairKey) {
	if w.static[i] && w.static[j] {
		return // two static bodies can't meaningfully collide
	}
	normal, overlap, ok := w.collide(i, j)
	if !ok {
		return
	}
	w.resolve(i, j, pk, normal, overlap)
}

// collide dispatches on the pair's shape types and returns the contact
// normal (pointing from i toward j) and penetration depth. ok is false if
// the pair isn't touching. There is no box-box case: ShapeBox is only
// ever created by AddStaticBox (see Shape's doc comment), and static-
// static pairs already return early in checkAndResolve, so the only
// shape combinations that ever reach here are sphere-sphere and
// sphere-box.
func (w *PhysicsWorld) collide(i, j int) (normal PVec3, overlap float64, ok bool) {
	si, sj := w.shapes[i].Type, w.shapes[j].Type
	switch {
	case si == ShapeSphere && sj == ShapeSphere:
		return w.collideSphereSphere(i, j)
	case si == ShapeSphere: // sj == ShapeBox
		n, o, ok := w.collideSphereBox(i, j)
		if !ok {
			return PVec3{}, 0, false
		}
		return PVec3{-n.X, -n.Y, -n.Z}, o, true // box->sphere flipped to i->j
	default: // si == ShapeBox, sj == ShapeSphere
		return w.collideSphereBox(j, i) // box(i)->sphere(j) is already i->j
	}
}

func (w *PhysicsWorld) collideSphereSphere(i, j int) (PVec3, float64, bool) {
	dx := w.posX[j] - w.posX[i]
	dy := w.posY[j] - w.posY[i]
	dz := w.posZ[j] - w.posZ[i]
	distSq := dx*dx + dy*dy + dz*dz
	minDist := w.shapes[i].Extent.X + w.shapes[j].Extent.X

	if distSq >= minDist*minDist || distSq < 1e-12 {
		return PVec3{}, 0, false
	}

	dist := math.Sqrt(distSq)
	return PVec3{dx / dist, dy / dist, dz / dist}, minDist - dist, true
}

// collideSphereBox tests a sphere against a (possibly rotated) box: the
// sphere's center is rotated into the box's local frame (inverse
// rotation), clamped to the box's half-extents to find the closest
// surface point, then rotated back to world space. Angle == 0 degenerates
// to the axis-aligned case with no extra cost. Returns the normal
// box-surface -> sphere-center.
func (w *PhysicsWorld) collideSphereBox(sphereIdx, boxIdx int) (PVec3, float64, bool) {
	box := w.shapes[boxIdx]
	r := w.shapes[sphereIdx].Extent.X

	local := PVec3{
		w.posX[sphereIdx] - w.posX[boxIdx],
		w.posY[sphereIdx] - w.posY[boxIdx],
		w.posZ[sphereIdx] - w.posZ[boxIdx],
	}
	if box.Angle != 0 {
		local = rotateVec(local, box.Axis, -box.Angle)
	}

	clamped := PVec3{
		clampFloat(local.X, -box.Extent.X, box.Extent.X),
		clampFloat(local.Y, -box.Extent.Y, box.Extent.Y),
		clampFloat(local.Z, -box.Extent.Z, box.Extent.Z),
	}

	closest := clamped
	if box.Angle != 0 {
		closest = rotateVec(closest, box.Axis, box.Angle)
	}
	closestWorld := PVec3{
		closest.X + w.posX[boxIdx],
		closest.Y + w.posY[boxIdx],
		closest.Z + w.posZ[boxIdx],
	}

	dx := w.posX[sphereIdx] - closestWorld.X
	dy := w.posY[sphereIdx] - closestWorld.Y
	dz := w.posZ[sphereIdx] - closestWorld.Z
	distSq := dx*dx + dy*dy + dz*dz

	if distSq >= r*r {
		return PVec3{}, 0, false
	}

	if distSq < 1e-12 {
		// Sphere center coincides with the closest surface point (deeply
		// embedded/tunneled). No well-defined nearest face without extra
		// work this plan doesn't need -- push out along the box's local
		// +Y so resolution still makes progress instead of dividing by ~0.
		return rotateVec(PVec3{0, 1, 0}, box.Axis, box.Angle), r, true
	}

	dist := math.Sqrt(distSq)
	return PVec3{dx / dist, dy / dist, dz / dist}, r - dist, true
}

// resolve applies positional correction and an impulse along normal
// (pointing i->j). When one side is static, all correction and impulse go
// to the dynamic side -- the correct limit of the general two-body
// formula as one mass -> infinity, not a bolt-on special case.
func (w *PhysicsWorld) resolve(i, j int, pk pairKey, normal PVec3, overlap float64) {
	iStatic, jStatic := w.static[i], w.static[j]

	switch {
	case jStatic:
		w.posX[i] -= normal.X * overlap
		w.posY[i] -= normal.Y * overlap
		w.posZ[i] -= normal.Z * overlap
	case iStatic:
		w.posX[j] += normal.X * overlap
		w.posY[j] += normal.Y * overlap
		w.posZ[j] += normal.Z * overlap
	default:
		w.posX[i] -= normal.X * overlap * 0.5
		w.posY[i] -= normal.Y * overlap * 0.5
		w.posZ[i] -= normal.Z * overlap * 0.5
		w.posX[j] += normal.X * overlap * 0.5
		w.posY[j] += normal.Y * overlap * 0.5
		w.posZ[j] += normal.Z * overlap * 0.5
	}

	rvx := w.velX[j] - w.velX[i]
	rvy := w.velY[j] - w.velY[i]
	rvz := w.velZ[j] - w.velZ[i]
	velAlongNormal := rvx*normal.X + rvy*normal.Y + rvz*normal.Z

	restitution, _ := combineMaterials(w.materials[w.materialID[i]], w.materials[w.materialID[j]])

	curr := w.contactSets[w.currIdx]
	prev := w.contactSets[1-w.currIdx]

	if velAlongNormal < 0 {
		switch {
		case jStatic:
			impulse := -(1 + restitution) * velAlongNormal
			w.velX[i] -= impulse * normal.X
			w.velY[i] -= impulse * normal.Y
			w.velZ[i] -= impulse * normal.Z
		case iStatic:
			impulse := -(1 + restitution) * velAlongNormal
			w.velX[j] += impulse * normal.X
			w.velY[j] += impulse * normal.Y
			w.velZ[j] += impulse * normal.Z
		default:
			impulse := -(1 + restitution) * velAlongNormal * 0.5
			w.velX[i] -= impulse * normal.X
			w.velY[i] -= impulse * normal.Y
			w.velZ[i] -= impulse * normal.Z
			w.velX[j] += impulse * normal.X
			w.velY[j] += impulse * normal.Y
			w.velZ[j] += impulse * normal.Z
		}

		curr[pk] = true
		if !prev[pk] {
			w.collisions = append(w.collisions, CollisionPair{
				A: i, B: j,
				Normal:  normal,
				Impulse: math.Abs(velAlongNormal),
			})
		}
	} else {
		curr[pk] = true
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
