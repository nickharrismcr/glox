# Plan: `physics_world` native type for glox

## Context

`lox_examples/3d_balls_physics_shaders.lox` currently implements per-object
physics (integration, boundary bounce, spatial-hash broad phase, impulse
collision resolution) entirely in Lox: a `MovingObject` class with an
`update()` method called per-ball per-frame, and a hand-rolled uniform grid
rebuilt every 3rd frame. This is the interpreter's hot path — method
dispatch and field access through the VM for every ball, every frame.

This plan moves the simulation itself into Go as a native type,
`physics_world`, following the exact convention already used by `batch`
(see `src/builtin/obj_builtin_batch.go` and `src/builtin/batch_methods.go`).
Lox keeps ownership of rendering (shadow/batch drawing, per-ball shader or
colour choice) and gameplay logic (explosion triggering) — only the
numeric simulation moves.

## Scope for this pass

**In:**
- Body storage (SoA: parallel position/velocity/radius/material slices)
- Materials table (restitution, friction, damping), referenced by id
- `step(dt)`: gravity integration, boundary bounce, uniform-grid broad
  phase, impulse-based collision resolution
- `collisions()`: pairs that **newly** started touching during the last
  `step()` call only — resting/still-touching pairs from prior frames are
  not re-reported, so a settled pile of balls doesn't spam the list

**Out (stays in Lox for now, revisit later):**
- Explosion triggering / `explode_at` / force application — gameplay
  policy, not core physics
- Direct-to-batch rendering (`draw_into(batch)`) — Lox still loops over
  ids and calls `get_position` per ball to feed the existing
  shadow/batch code
- Friction actually applied in impulse resolution (computed via
  `combineMaterials` but not yet used — tangential friction is a
  follow-up)
- Non-equal-mass bodies (resolution currently assumes equal mass, 50/50
  positional correction — matches what the current Lox version does)

## Files to add

Both files go in `src/builtin/` alongside `obj_builtin_batch.go` /
`batch_methods.go`, which they mirror stylistically.

- `src/builtin/obj_builtin_physics_world.go` — constructor, SoA
  simulation (`PhysicsWorld`), and the native object wrapper
  (`PhysicsWorldObject`)
- `src/builtin/physics_world_methods.go` — method registration
  (`RegisterAllPhysicsWorldMethods`), following `batch_methods.go`'s
  arg-checking and stack-unmarshalling style exactly

Full contents are attached alongside this plan
(`obj_builtin_physics_world.go`, `physics_world_methods.go`) — copy them
in verbatim as a starting point; adjust names/behaviour as needed during
implementation.

## Edits to existing files

1. **`src/core/object.go`** — add a new native type constant, after
   `NATIVE_BATCH_INSTANCED`:

   ```go
   const (
       NATIVE_FLOAT_ARRAY NativeType = iota
       NATIVE_VEC2
       NATIVE_VEC3
       NATIVE_VEC4
       NATIVE_WINDOW
       NATIVE_IMAGE
       NATIVE_TEXTURE
       NATIVE_RENDER_TEXTURE
       NATIVE_CAMERA
       NATIVE_SHADER
       NATIVE_BATCH
       NATIVE_BATCH_INSTANCED
       NATIVE_PHYSICS_WORLD   // <-- add this
   )
   ```

2. **`src/vm/builtin.go`** — register the constructor, next to the
   `batch`/`batch_instanced` registrations:

   ```go
   defineBuiltIn(vm, "", "physics_world", builtin.PhysicsWorldBuiltIn)
   ```

## API surface

```
world = physics_world(min_vec3, max_vec3, cell_size, gravity_vec3)

mat_id = world.add_material(restitution, friction, damping)   // -> int
id     = world.add(pos, vel, radius, mat_id)                  // -> int
world.remove(id)
pos    = world.get_position(id)                                // -> vec3
world.step(dt)
pairs  = world.collisions()   // -> list of [a, b, normal_vec3, impulse]
n      = world.count()        // -> int (active bodies)
```

`collisions()` returns each pair as a 4-element immutable list
(`pair[0]`=a id, `pair[1]`=b id, `pair[2]`=contact normal as vec3,
`pair[3]`=impulse magnitude) rather than a named-field struct, since Lox
doesn't appear to have a lightweight record type — confirm this is
acceptable, or add a small wrapper class in Lox if named-field access
(`pair.a`, `pair.impulse`) is wanted instead.

## Implementation steps

1. Add the two new files to `src/builtin/`.
2. Apply the two edits above (`core/object.go`, `vm/builtin.go`).
3. `go build ./...` — must compile clean. (Confirmed compiling against
   this repo as of this plan's authoring, go1.24 + the currently pinned
   raylib-go version.)
4. Write a small standalone `.lox` smoke test: create a world, add ~10
   bodies with overlapping starting positions/velocities, step it a few
   times, print positions and any collisions. Confirm bodies separate,
   bounce off boundaries, and `collisions()` reports each impact exactly
   once (not every frame while resting).
5. Port `lox_examples/3d_balls_physics_shaders.lox` to use `physics_world`
   in place of the `MovingObject` class + hand-rolled spatial hash:
   - Replace per-object `update()` calls with a single `world.step(dt)`
     per frame.
   - Replace the Lox-side grid rebuild/query with nothing — it's gone,
     subsumed into `step()`.
   - Keep the `ids` list (from `world.add(...)`) as the handle registry
     for rendering: still loop over `ids`, call `get_position(id)`, feed
     into the existing shadow/batch drawing code.
   - Keep per-ball shader/colour selection and explosion-timer logic in
     Lox, driven off `world.collisions()` for impact detection instead of
     the old `check_collision`/`resolve_collision` methods.
6. Benchmark before/after frame time at the current ball count (450) and
   at a higher count (e.g. 2000+) to quantify the win and check whether
   `narrowPhase`'s per-step `map[pairKey]bool` allocations
   (`checked`, `currContacts`) show up as a bottleneck worth replacing
   with a flat slice/bitset.

## Known follow-ups (not blocking this pass)

- `checked` and `currContacts` in `narrowPhase`/`Step` allocate a new map
  every step — fine to start, but worth profiling once ball counts grow;
  a reusable flat structure keyed by packed `(a,b)` ints would avoid the
  per-frame allocation.
- No free-list for removed body ids — `Remove` just tombstones via
  `active[id] = false`; slots aren't reused. Add a free-list if churn
  (frequent add/remove) becomes common.
- Unequal masses: would need a `mass` field alongside `radius` and a
  mass-weighted split in `checkAndResolve`'s positional correction and
  impulse response, instead of the current 50/50 equal-mass assumption.
- Tangential friction: `combineMaterials` already computes a combined
  friction value but nothing consumes it yet.
- `explode_at`/force application and `pop_exploded`/`mark_exploded`
  event-queue pattern were discussed and deferred — see conversation
  history for the fuller sketch if/when explosion logic also moves to Go.
