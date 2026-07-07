# GLox Shader System Documentation

> **Note:** `shader`, `window`, `texture` and the other graphics constructors now live in the built-in `gfx` module. Add `from gfx import *` (keeps the unqualified names used below) or use `import gfx` and `gfx.shader(...)`.

## Overview

The GLox interpreter now includes comprehensive shader support using raylib-go bindings. This allows for custom GLSL shaders to be loaded, configured, and used in 3D rendering applications.

## Features Implemented

### 1. Shader Object Type
- New `shader()` builtin function creates shader objects
- Supports loading from files or memory
- Automatic cleanup and resource management

### 2. Shader Methods
- `load_from_memory(vertex_code, fragment_code)` - Load shader from GLSL code strings
- `get_location(uniform_name)` - Get uniform location by name
- `set_value_float(location, value)` - Set float uniform
- `set_value_vec2(location, vec2_value)` - Set vec2 uniform
- `set_value_vec3(location, vec3_value)` - Set vec3 uniform
- `set_value_vec4(location, vec4_value)` - Set vec4 uniform
- `is_valid()` - Check if shader loaded successfully
- `unload()` - Free shader resources

### 3. Window Methods for Shader Mode
- `begin_shader_mode(shader)` - Activate custom shader
- `end_shader_mode()` - Return to default shader

## Usage Examples

### Basic Shader Loading from Files
```lox
// Load shader from files
shader_obj = shader("vertex.vs", "fragment.fs")
if (shader_obj.is_valid()) {
    print("Shader loaded successfully!")
}
```

### Loading Shader from Memory
```lox
vertex_code = "#version 330\n..."
fragment_code = "#version 330\n..."

shader_obj = shader()
shader_obj.load_from_memory(vertex_code, fragment_code)
```

### Setting Uniforms
```lox
// Get uniform locations
time_loc = shader_obj.get_location("time")
color_loc = shader_obj.get_location("color")
position_loc = shader_obj.get_location("position")

// Set uniform values
shader_obj.set_value_float(time_loc, current_time)
shader_obj.set_value_vec3(color_loc, vec3(1.0, 0.5, 0.2))
shader_obj.set_value_vec4(position_loc, vec4(x, y, z, w))
```

### Using Shaders in Rendering
```lox
// Apply shader to rendering
win.begin_shader_mode(shader_obj)
win.sphere(position, radius, color)
win.end_shader_mode()
```

## Demo Applications

### 1. 3D Shader Demo (`3d_shader_demo.lox`)
Simple demo showing rainbow shader effects on bouncing spheres.

### 2. Enhanced Physics Demo (`3d_balls_physics_shaders.lox`)
Full physics simulation with 30% of objects using shader effects.

### 3. Advanced Shader Demo (`advanced_shader_demo.lox`)
Multiple shaders with different effects:
- Rainbow shader with time-based color cycling
- Pulse shader with distance-based energy effects
- Demonstrates vec3 uniform setting for pulse center

### 4. Memory Shader Demo (`memory_shader_demo.lox`)
Shows loading shaders from code strings with brightness animation.

## Shader Examples

### Rainbow Effect (`rainbow.vs` / `rainbow.fs`)
Time-based rainbow colors with sinusoidal color cycling.

### Pulse Effect (`pulse.vs` / `pulse.fs`)
Energy pulse effect with configurable center position and ring animations.

### Basic Lighting (`basic_lighting.vs` / `basic_lighting.fs`)
Multi-light Blinn-Phong lighting model with up to 4 light sources.

## Technical Implementation

### Go Integration
- New `ShaderObject` type in `obj_builtin_shader.go`
- Added `NATIVE_SHADER` type to core object system
- Window methods in `win_methods.go` for shader mode control
- Registered `shader` builtin function in main builtin system

### Performance Considerations
- Shader validation checks prevent crashes from invalid shaders
- Proper resource cleanup with `unload()` method
- Efficient uniform location caching
- Compatible with existing raylib rendering pipeline

### Error Handling
- Graceful fallback to default rendering if shaders fail
- Validation of uniform locations before setting values
- Type checking for uniform value types
- Descriptive error messages for debugging

## Rendering Pipeline: What `begin_shader_mode` Actually Affects

`begin_shader_mode(shader)` / `end_shader_mode()` binds the shader raylib's `rlgl`
layer currently has active. Whether a given draw call is affected depends on
*how* that call renders under the hood:

- **Immediate-mode primitives respect it.** `win.plane(...)`, sphere/cube/line/
  rectangle draws, and `win.draw_render_texture(...)` are all implemented in
  raylib as direct `rlBegin`/`rlVertex3f`/`rlEnd` calls with no `Material` of
  their own, so they render with whatever shader is currently bound.
- **Plain `batch` (non-instanced) also respects it.** `DrawBatch.Draw()`
  (`src/builtin/obj_builtin_batch.go`) draws each entry with `rl.DrawCube`,
  `rl.DrawSphere`, or `rl.DrawTriangle3D` — again all immediate-mode, no
  per-entry `Material` — so wrapping `my_batch.draw()` in `begin_shader_mode`
  works exactly like the reference docs' example shows. This is the key
  distinction from `batch_instanced` below: same-sounding name, very
  different render path and shader support.
- **`batch_instanced` draws do not respect it.** `cubeBatch.draw(cam)` calls
  `BatchInstancedObject.Draw` (`src/builtin/obj_builtin_batch_instanced.go`),
  which renders via `rl.DrawMeshInstanced(mesh, material, transforms, count)`.
  raylib's `DrawMesh`/`DrawMeshInstanced` always render with the mesh's own
  `Material.Shader` field, ignoring whatever shader `begin_shader_mode` has
  currently bound. Wrapping a `batch.draw(cam)` call in
  `win.begin_shader_mode(my_shader)` has **no visible effect**.

### The instanced-batch shader is a fixed, shared singleton

Every `batch_instanced(...)` object gets its `Material.Shader` set in
`MakeModel()` to a package-level singleton, `shaderInstanced`
(`src/builtin/obj_builtin_batch_instanced.go`). It is loaded once, lazily, from
`src/shaders/instanced/base_lighting_instanced.vs` and
`src/shaders/instanced/lighting.fs` — a basic per-instance-transform,
ambient-plus-up-to-4-lights Blinn-Phong shader (in practice, only `ambient` is
ever set by `InitShader()`, so cubes render as flat-lit textured geometry with
no directional/point lights active). Consequences:

- **No per-script customisation.** A `.lox` script cannot swap in a different
  GLSL shader for its cube batches — `shader()` + `begin_shader_mode` only
  reaches immediate-mode draws, not the batch's mesh material.
- **No per-batch customisation either.** All `BatchInstancedObject`s in the
  same process — even ones created from different textures/sizes — literally
  share the one `*rl.Shader` instance. Two batches cannot have two different
  looks at the Go level without changing this singleton pattern.
- **No per-instance colour/uniform data.** Instances only carry a transform
  matrix (`translation`/`rotation`); there's no per-instance colour or custom
  attribute channel, so effects that vary per-cube (e.g. tinting by distance,
  by stack, or by height) aren't expressible even by editing the shared
  shader — every instance in a batch is shaded identically aside from its
  transform and the (single, per-texture) diffuse texture.

To give instanced batches real custom-shader or per-instance-uniform support
would require an engine change: e.g. a `batch.set_shader(shader)` method that
overrides `Model.material.Shader` per `BatchInstancedObject` (replacing the
shared singleton), plus replicating the `instanceTransform`/`viewPos`/`mvp`
wiring `InitShader()` currently does by hand.

### No depth texture — true distance/depth effects aren't reachable from `.lox`

`render_texture(...)` (`RenderTextureObject`) exposes only its colour
`get_texture()`; there is no accessor for a sampleable depth attachment. That
means a post-process fragment shader bound via `begin_shader_mode` +
`win.draw_render_texture(fb, ...)` has no way to know true per-pixel scene
depth — it only ever sees the already-flattened colour image. Real
distance-based effects on either primitives or instanced batches need either
a depth-texture-capable render target exposed to `.lox`, or per-vertex
world-position/distance computed directly in a custom vertex shader bound to
the actual draw call (which, per above, is only possible today for
immediate-mode primitives, not `batch_instanced`).

## Future Enhancements

Potential areas for expansion:
1. Texture binding for shader samplers
2. Matrix uniform support
3. Geometry and compute shader support
4. Shader uniform blocks
5. Instanced rendering with shader support
6. Post-processing effects pipeline

## Performance Impact

The shader system adds minimal overhead:
- Shader objects are lightweight wrappers around raylib structures
- Uniform updates are efficient native operations
- Begin/end shader mode has negligible cost
- Compatible with existing collision detection and physics systems

The enhanced 3D demo maintains 60+ FPS with 100+ objects and shader effects enabled.
