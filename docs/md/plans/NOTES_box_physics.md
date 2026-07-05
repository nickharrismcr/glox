# Notes: box shapes, static bodies, and tilted boxes for physics_world

Addendum to `PLAN_physics_world.md` — covers shape/staticness/rotation
extensions discussed after the initial sphere-only implementation. Not
yet reduced to concrete Go code or verified against the repo (the base
plan's two files were compile-checked; this addendum is design notes to
implement against that foundation).

**Status: implemented, with dynamic boxes descoped.** Sections 2–4
(static bodies, tilted static boxes, and `win.cube_rotated()` drawing)
are implemented as designed. The *dynamic* half of section 1 — an
`add_box()` constructor for axis-aligned boxes that fall/move under
gravity — was built, then removed at the user's request (not something
they'd asked for). `Shape`/`ShapeType` and the box-box/sphere-box
collision math from section 1 below are still accurate as *design*
background, but in the shipped code `ShapeBox` is only ever created by
`add_static_box` — there is no dynamic box constructor, and the box-box
collision path was deleted as dead code (two static bodies never
collide with each other, so it was never reachable). Read section 1 as
history/rationale, not as a description of what `add_box()` does today,
because `add_box()` doesn't exist.

## 1. Axis-aligned boxes (no rotation)

Two decision points on any "add a shape" extension: rotating or not,
static or dynamic. Start here (no rotation) before adding tilt.

**(Historical — the dynamic-box half of this section was removed after
implementation; see the Status note above.)**

**Per-body shape data**

```go
type ShapeType uint8
const (
    ShapeSphere ShapeType = iota
    ShapeBox
)

type Shape struct {
    Type   ShapeType
    Extent PVec3 // sphere: radius in Extent.X only; box: half-extents X/Y/Z
}
```

Store as `shapes []Shape`, parallel to the existing SoA slices, indexed
by body id. Kept as its own slice rather than folded into the hot
integration loop, since shape data is only read during narrow phase.

**Broad phase barely changes — for dynamic bodies.** Grid cell
placement still just needs *a* bounding radius: for a box, use
half-extents' magnitude (`sqrt(hx²+hy²+hz²)`) as a conservative bounding
sphere for cell membership. Slightly pessimistic at the corners, but
keeps grid code untouched.

This assumes a body's extent stays comparable to `cellSize` (already
true today — see `PHYS_CELL_SIZE` sized relative to max ball diameter
in `3d_balls_physics_shaders.lox`). It **breaks down for large static
geometry** (a 10+ unit shelf or ramp): `rebuildGrid()` inserts a body
into exactly one cell keyed by its centroid, and `narrowPhase()` only
scans the 27 cells around a body's *own* cell. A ball near the far end
of a long platform, more than a cell or two from the platform's
centroid, would never find it. Section 2 below addresses this by
routing static bodies around the grid entirely rather than making the
grid itself handle arbitrary-sized cells.

**Narrow phase dispatches on the pair's shape types** — `checkAndResolve`
becomes a dispatcher:

```go
switch {
case shapes[i].Type == ShapeSphere && shapes[j].Type == ShapeSphere:
    resolveSphereSphere(i, j) // existing code, unchanged
case shapes[i].Type == ShapeBox && shapes[j].Type == ShapeBox:
    resolveBoxBox(i, j)
default:
    resolveSphereBox(i, j) // handles either order
}
```

- **Box-box (AABB slab test):** compute overlap on each axis
  independently — `overlapX = (hx_i + hx_j) - |posX_j - posX_i|`, same
  for Y/Z. All three positive → intersecting; the axis with the
  *smallest* overlap is the separating axis — push apart and reflect
  velocity along that single axis only (not a full 3D normal).
- **Sphere-box:** clamp the sphere's center to the box's min/max extents
  per axis to get the closest point on the box surface, then treat that
  point like a second sphere center (radius 0) for the existing
  distance/normal/impulse math in `checkAndResolve` — mostly reusable
  once you have the closest point.

**Lox-facing:** a separate `add_box()` method (parallel to `add()`)
rather than overloading `add` with a shape argument — matches the
existing convention of type-specific constructors (`batch(BATCH_CUBE)`).

## 2. Static (fixed) bodies

For level geometry (walls, platforms) that balls bounce off but which
never move themselves — simpler than dynamic boxes, and orthogonal to
shape.

**A `static []bool` flag**, parallel to the SoA slices. Static bodies
still live in the same arrays and the same grid (need to be found by the
neighbour scan like anything else) — they just skip two passes:

```go
func (w *PhysicsWorld) integrate(dt float64) {
    for i := range w.posX {
        if !w.active[i] || w.static[i] {
            continue // static bodies never move
        }
        ...
    }
}
```

Same guard needed in `boundaryCollisions` — a fixed platform shouldn't
bounce off the world bounds either.

**Static bodies skip the grid entirely, rather than being inserted into
it.** Per the broad-phase gap noted in section 1, a single centroid-keyed
cell can't represent a large platform's true extent, and multi-cell
insertion would mean touching `rebuildGrid()`'s insertion loop and
paying a cost proportional to a static body's size every frame (it never
moves, so that cost is pure waste). Instead:

- `rebuildGrid()` skips bodies where `static[i]` is true — they're never
  inserted, so `narrowPhase()`'s grid-based scan only ever considers
  dynamic-dynamic pairs, unchanged from today.
- A separate `staticIDs []int` slice is appended to (never rescanned)
  whenever `add_static_box` creates a body.
- `Step()` gets one additional direct pass after `narrowPhase()`: for
  each active dynamic body `i`, loop over `staticIDs` and call
  `checkAndResolve(i, staticID)` directly — no cell lookup. Static body
  counts (walls/ramps/platforms) are expected to stay small (single or
  low double digits), so this is O(dynamic × static), which is cheap in
  practice and requires zero changes to the existing grid code.

This keeps the "barely changes" property true for the *dynamic* broad
phase while giving statics an exact (not bounding-sphere-approximated)
membership test.

**Resolution becomes asymmetric when one side is static.** Today's
`checkAndResolve` splits positional correction 50/50 and impulse equally
both ways (equal-mass assumption). With a static body, all correction
and all impulse goes to the dynamic one — this is the correct limit of
the general two-body formula as one mass → infinity, not a bolt-on
special case:

```go
if w.static[i] && w.static[j] {
    return // two static bodies can't meaningfully collide
}
if w.static[j] {
    // push i only, full impulse (not halved) applied to i
} else if w.static[i] {
    // mirror: push j only
} else {
    // existing equal-split code, unchanged
}
```

**Lox-facing:**

```
platform_mat = world.add_material(0.6, 0.5, 1.0)
id = world.add_static_box(pos, vec3(hx, hy, hz), mat_id)
```

No velocity argument — signals fixed-ness without needing to inspect a
boolean flag.

## 3. Tilted (rotated) static boxes

Rotation is much cheaper here than the general rigid-body case, because
these boxes are static: orientation is set once at creation and never
updated — no angular velocity, no torque, no per-frame orientation
integration. It's "test a moving sphere against a tilted box," not
"simulate rotational dynamics."

**Representation** — axis-angle, matching what `batch_instanced`
already uses for rotated instances (`MakeInstance(x,y,z,axisX,axisY,axisZ,angle)`
in `obj_builtin_batch_instanced.go`, built on `rl.MatrixRotate(axis, angle)`).
Reusing the same convention keeps physics and rendering consistent and
avoids introducing a second rotation representation. Quaternion is the
alternative if composing rotations becomes necessary later, but
axis-angle is simplest for "set once, never touched again."

```go
type Shape struct {
    Type   ShapeType
    Extent PVec3   // sphere: radius in Extent.X only; box: half-extents X/Y/Z
    Axis   PVec3   // normalized; unused (zero value) for spheres and axis-aligned boxes
    Angle  float64 // degrees, matches rl.MatrixRotate convention; 0 = axis-aligned
}
```

Folded into the same `shapes []Shape` slice from section 1, rather than
a second `map[int]*BoxShape` keyed by body id. Both would be indexed by
body id and both would store half-extents, so a separate map only
duplicates data and adds a lookup — one parallel slice, paid once per
body regardless of shape type, matches the existing convention (`shapes`
is already read only during narrow phase, not the hot integration loop).

**Collision test — sphere vs. oriented box (transform into box-local space):**

1. `local = sphereCenter - boxCenter`
2. Rotate `local` by the box's **inverse** rotation → box-local frame,
   where it's just an AABB
3. Clamp each axis of `local` to `[-halfExtent, +halfExtent]` → closest
   point, still box-local
4. Rotate that point back by the box's rotation, add `boxCenter` →
   closest point in world space
5. Treat like the sphere-box case above: `normal = (sphereCenter -
   closestPoint) / dist`, check `dist < sphereRadius`, feed into the
   same static-asymmetric resolution from section 2

Steps 2 and 4 are the only new math; everything downstream reuses
existing code.

**Lox-facing** — a pure-Lox helper for axis-angle construction (no Go
needed), so level design reads naturally:

```
ramp_mat = world.add_material(0.4, 0.6, 1.0)
id = world.add_static_box(pos, vec3(hx, hy, hz), vec3(1, 0, 0), 30, ramp_mat)
```

## 4. Drawing tilted boxes

**Raylib has no `DrawCube(rotation)` overload.** Rotated primitives
always go through a mesh + transform matrix. The repo already solves
this in `obj_builtin_batch_instanced.go`:

```go
translation := rl.MatrixTranslate(x, y, z)
rotation := rl.MatrixRotate(axis, angle*rl.Deg2rad)
```

fed into `rl.DrawMeshInstanced`. Tilted boxes should reuse this exact
approach, but **not** via the instanced-thousands system —
`batch_instanced` is built for 100k+ *identical* objects in one draw
call, overkill for a handful of static ramps/platforms. Draw each one
individually per frame instead:

```go
translation := rl.MatrixTranslate(pos.X, pos.Y, pos.Z)
rotation := rl.MatrixRotate(axis, angle*rl.Deg2rad) // same as MakeInstance
transform := rl.MatrixMultiply(rotation, translation)
rl.DrawMesh(cubeMesh, material, transform)
```

`cubeMesh` = a single shared `rl.GenMeshCube(1,1,1)`, scaled per-box via
the transform or via half-extents baked in at creation.

**This lives on `window`, not on `physics_world`.** `physics_world` is
pure simulation today — it never draws anything; every existing example
(`3d_balls_physics_shaders.lox`) reads state back with `get_position()`
and calls a `win.*` primitive itself (`win.cube()`, `win.sphere()`, both
thin wrappers over `rl.DrawCube`/`rl.DrawSphere` in `win_methods.go`).
Tilted boxes should follow that same split: a new `win.cube_rotated(pos,
size, axis, angle, color)` builtin in `win_methods.go`, built on the
`GenMeshCube`+`DrawMesh`+matrix approach above. The shared `cubeMesh`
(and a default material) is created lazily on first call and cached on
the `WindowObject` — not per-`PhysicsWorld` — since a script could in
principle draw rotated boxes without any physics_world at all, and
that's consistent with `win.init()` already owning other GPU-side
resources.

**Single source of truth for orientation.** Store axis+angle exactly
once, at `add_static_box` time, in `PhysicsWorld`. Expose it back to Lox
via a getter (`world.get_box_transform(id)` → position + axis + angle)
so the render call always draws exactly what the physics collided
against. If Lox independently tracked its own copy of the tilt for
drawing, a future edit to one and not the other would make balls
visibly bounce off a surface that doesn't match what's drawn.

**Known dead code, unrelated but worth knowing:** the plain `batch`'s
`BatchEntry` struct already has a `Rotation rl.Vector3` field that
`Draw()` never reads (confirmed by inspecting the draw loop in
`obj_builtin_batch.go`). Not a blocker here since tilted boxes don't go
through `batch`, but don't assume that field does anything today.

## Decisions

- **Rotation scope: static only.** Rotated boxes (section 3) are always
  static. This rules out OBB-OBB (SAT) and angular-velocity integration
  entirely — the only rotated collision test needed is sphere-vs-OBB
  (section 3).
- **No dynamic boxes.** `add_box()` (section 1) was implemented, then
  removed — it wasn't something the user had asked for. Every `ShapeBox`
  in the shipped code comes from `add_static_box`, so box-box collision
  is unreachable (two statics never collide) and was deleted along with
  it. The `Shape`/`ShapeType` split and the box-box AABB math in section
  1 remain here as design background only.
- **Static bodies bypass the grid** (see section 2) rather than being
  grid-inserted via multi-cell coverage — a direct dynamic × static pass
  after `narrowPhase()`, using a cached `staticIDs []int`.
- **Axis-angle**, not quaternion, for box orientation — matches
  `batch_instanced`'s existing convention (`MakeInstance`/`rl.MatrixRotate`),
  and static boxes never compose rotations after creation, so quaternion's
  main advantage doesn't apply here.
- **One `Shape` struct**, not a separate `shapes []Shape` plus
  `map[int]*BoxShape` — axis/angle fold into `Shape` directly (zero value
  = axis-aligned), since both would otherwise duplicate half-extents
  under the same body-id index.
- **Rotated-box drawing is a `win.cube_rotated(...)` builtin**, not a
  `physics_world` method — matches the existing split where
  `physics_world` only simulates and Lox/`win.*` does all drawing. Shared
  `cubeMesh` lives on `WindowObject`, created lazily on first use.

## Still open

- Nothing outstanding. The `add_box` vs. `add_static_box` unification
  question this section used to raise is moot now that `add_box` doesn't
  exist — `add_static_box` is the only box constructor.
