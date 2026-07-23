# GLox Builtins 

## Table of Contents

1. [Builtin Functions](#builtin-functions)
2. [Window Object](#window-object)
3. [Batch Object](#batch-object)
4. [Texture Object](#texture-object)
5. [RenderTexture Object](#rendertexture-object)
6. [Camera Object](#camera-object)
7. [Shader Object](#shader-object)
8. [Image Object](#image-object)
9. [FloatArray Object](#floatarray-object)
10. [PhysicsWorld Object](#physicsworld-object)
11. [Vector Objects](#vector-objects)
12. [System Modules](#system-modules)
13. [Color Utilities Module](#color-utilities-module)

---

## Builtin Functions

### Core Functions

- **`type(value)`** - Returns the type of a value as a string

### Type Conversion Functions

- **`float(value)`** - Converts value to floating point number
- **`int(value)`** - Converts value to integer

### Container Functions

- **`len(container)`** - Returns the length of a container (string, list, dict, tuple)
- **`range(n)` / `range(start, end)` / `range(start, end, step)`** - Native integer iterator, for use with `foreach`
- **`append(list, item)`** - Appends an item to a list. `list.append(item)` (the method form) is the more idiomatic spelling — both do the same thing.

### String Functions

- **`replace(string, old, new)`** - Replaces occurrences of 'old' with 'new' in string. `string.replace(old, new)` (the method form) is equivalent.
- **`str(value)`** - Converts any value to its string form (uses a class's `toString` for instances)
- **`format(fmt, args...)`** - printf-style formatting using Go verbs (`%s`, `%d`, `%f`, `%v`); wraps `fmt.Sprintf`

**String interpolation:** any `${ expr }` inside a string literal (either `"…"` or `'…'`) is evaluated and stringified into the result, e.g. `print "total: ${count} (${pct}%)"`. Values are stringified the same way as `str()`. Write a literal `$` as `$$`. It is pure sugar: `"a${x}b"` desugars to `("a" & str(x) & "b")`. Note that `&`, not `+`, is the string concatenation operator — `"a" + "b"` raises a `RunTimeError`.

### Graphics Functions

> **Note:** graphics constructors and helpers live in the built-in **`gfx`** module (and `physics_world` in **`physics`**). Import them before use — `from gfx import *` (names stay unqualified, as in the examples below) or `import gfx` then `gfx.window(...)`. Only `vec2`/`vec3`/`vec4` remain global.

- **`window(width, height)`** - Creates a window object (`gfx`)
- **`batch(type)`** - Creates a new batch object for optimized rendering of primitives (`gfx`)
- **`texture` / `render_texture` / `shader` / `camera` / `image` / `batch_instanced` / `float_array`** - native graphics objects (`gfx`)
- **`draw_png(filename, width, height, data)`** - Writes PNG image data to file (`gfx`)
- **`encode_rgba(r, g, b)` → float / `decode_rgba(value)` → `[r, g, b]`** - pack/unpack an RGB triple (0-255 each) into the single float `float_array`/`draw_array` expect for color-encoded data (`gfx`)

### Random & Color Functions

- **`rand()`** - Global, no import needed. Returns a float in `[0.0, 1.0)`, like Python's `random.random()` — **not** an integer, so `rand() % n` raises `RunTimeError` (`%` requires integer operands). Use [`random.integer(min, max)`](#system-modules) for a random integer in a range.
- For RGB manipulation (fade, tint, brightness, HSV conversion, etc.), see the [Color Utilities Module](#color-utilities-module).

### Special Functions

- **`lox_mandel_array(width, height, zoom, center_x, center_y, max_iter)`** - Generates Mandelbrot set data as a float array (`gfx`)

---

## Window Object

The window object provides the main interface for graphics rendering and input handling.

### Window Creation

```lox
import gfx;

var win = gfx.window(800, 600);
win.init();
```

### Window Management Methods

- **`init()`** - Initialize the window with dimensions specified during creation
- **`begin()`** - Begin drawing frame
- **`end()`** - End drawing frame and present to screen
- **`close()`** - Close the window
- **`should_close()`** - Returns true if window should close
- **`toggle_fullscreen()`** - Toggle between fullscreen and windowed mode
- **`get_screen_width()`** - Returns current screen width
- **`get_screen_height()`** - Returns current screen height
- **`set_target_fps(fps)`** - Set target frame rate
- **`get_fps()`** - Get current frame rate 

### Drawing Methods

#### Basic Drawing
- **`clear(color_vec4)`** - Clear screen with specified Vec4 color
- **`pixel(x, y, color_vec4)`** - Draw a single pixel
- **`line(x1, y1, x2, y2, color_vec4)`** - Draw a line
- **`line_ex(x1, y1, x2, y2, thickness, color_vec4)`** - Draw a line with specified thickness
- **`rectangle(x, y, width, height, color_vec4)`** - Draw a rectangle outline
- **`circle(x, y, radius, color_vec4)`** - Draw a circle outline
- **`circle_fill(x, y, radius, color_vec4)`** - Draw a filled circle
- **`triangle(x1, y1, x2, y2, x3, y3, color_vec4)`** - Draw a triangle
- **`text(text, x, y, size, color_vec4)`** - Draw text

#### Advanced Drawing
- **`draw_array(float_array)`** - Draw an RGB encoded float array as colour image
- **`draw_texture(texture, x, y, color_vec4)`** - Draw a texture
- **`draw_texture_flip(texture, x, y, color_vec4, flip_x)`** - Draw a texture, mirrored horizontally when `flip_x` is true
- **`draw_texture_scaled(texture, x, y, color_vec4, flip_x, scale)`** - Like `draw_texture_flip`, but scales the drawn size by `scale` (dest size = frame size × `scale`); `x, y` anchor the top-left corner of the *scaled* sprite
- **`draw_texture_rect(texture, x, y, src_x, src_y, src_w, src_h, color_vec4)`** - Draw part of a texture
- **`draw_texture_pro(texture, src_x, src_y, src_w, src_h, dest_x, dest_y, dest_w, dest_h, origin_x, origin_y, rotation, color_vec4)`** - Full-control texture draw: arbitrary source/destination rectangles, rotation origin, and rotation angle
- **`draw_render_texture(render_texture, x, y, color_vec4)`** - Draw a render texture
- **`draw_render_texture_ex(render_texture, x, y, rotation, scale, color_vec4)`** - Draw render texture with transformation

#### Off-screen Texture Mode
- **`begin_texture_mode(render_texture)`** - Redirect subsequent drawing calls to a `render_texture` instead of the screen
- **`end_texture_mode()`** - Stop redirecting; drawing calls go back to the window

### Blend Modes
- **`begin_blend_mode(mode)`** - Begin custom blend mode (use win.BLEND_* constants)
- **`end_blend_mode()`** - End custom blend mode

#### Blend Mode Constants
- **`win.BLEND_ALPHA`** - Alpha blending (default)
- **`win.BLEND_ADD`** - Additive blending
- **`win.BLEND_MULTIPLY`** - Multiplicative blending  
- **`win.BLEND_SUBTRACT`** - Subtractive blending
- **`win.BLEND_DEFAULT`** - Default blend mode (same as BLEND_ALPHA)

#### Batch Type Constants
- **`win.BATCH_CUBE`** - For cube batch creation
- **`win.BATCH_SPHERE`** - For sphere batch creation
- **`win.BATCH_TRIANGLE3`** - For 3D triangle batch creation
- **`win.BATCH_CIRCLE3`** - For 3D circle batch creation (see [Batch Object](#batch-object))

#### Texture Wrap Mode Constants
- **`win.WRAP_REPEAT`** / **`win.WRAP_CLAMP`** / **`win.WRAP_MIRROR_REPEAT`** / **`win.WRAP_MIRROR_CLAMP`** - For `texture.set_wrap_mode()`

### Input Methods
- **`key_down(key_code)`** - Check if key is currently pressed (use win.KEY_* constants)
- **`key_pressed(key_code)`** - Check if key was just pressed this frame (use win.KEY_* constants)

#### Key Constants
Key constants are available as `win.KEY_*` (e.g., `win.KEY_SPACE`, `win.KEY_ESCAPE`, `win.KEY_A`, `win.KEY_ENTER`, etc.)

### 3D Rendering
- **`begin_3d(camera)`** - Begin 3D mode with camera
- **`end_3d()`** - End 3D mode
- **`cube(x, y, z, width, height, length, color)`** - Draw a 3D cube
- **`cube_wires(x, y, z, width, height, length, color)`** - Draw a 3D cube wireframe
- **`cube_rotated(position, size, axis, angle, color)`** - Draw a solid 3D cube rotated `angle` degrees around `axis` (position/size/axis are vec3, angle is a number in degrees, color is vec4)
- **`cube_wires_rotated(position, size, axis, angle, color)`** - Draw a 3D cube wireframe rotated `angle` degrees around `axis` - the rotated counterpart of `cube_wires`
- **`textured_cube(texture, position, size, base_color)`** - Draw a 3D cube with base color and texture overlay (supports both texture and render_texture objects)
- **`sphere(x, y, z, radius, color_vec4)`** - Draw a 3D sphere
- **`cylinder(x, y, z, radius_top, radius_bottom, height, slices, color)`** - Draw a 3D cylinder
- **`grid(slices, spacing)`** - Draw a 3D grid
- **`plane(x, y, z, width, length, color)`** - Draw a 3D plane
- **`ellipse3(center_x, center_y, center_z, radius_x, radius_z, color_vec4)`** - Draw a 3D ellipse
- **`triangle3(x1,y1,z1,z2,y2,z2,x3,y3,z3,color_vec4)`** - Draw a triangle in 3D

### Shader Support
- **`begin_shader_mode(shader)`** - Begin custom shader mode
- **`end_shader_mode()`** - End custom shader mode

---

## Batch Object

The batch object provides high-performance rendering for large numbers of similar primitives. Instead of making individual draw calls for each object, batching allows thousands of objects to be rendered in a single optimized draw call.

### Batch Creation

```lox
import gfx;

var win = gfx.window(800, 600);
win.init();

var cube_batch = gfx.batch(win.BATCH_CUBE);      // Create a batch for cubes
var sphere_batch = gfx.batch(win.BATCH_SPHERE);  // Create a batch for spheres
```

**Supported batch types:**
- `win.BATCH_CUBE` - For rendering cubes
- `win.BATCH_SPHERE` - For rendering spheres
- `win.BATCH_TRIANGLE3` - For rendering 3D triangles (`add_triangle3`/`set_triangle3*` methods)
- `win.BATCH_CIRCLE3` - For rendering flat 3D circles (`add_circle3`/`set_circle3*`/`set_circle_texture` methods)

### Adding Primitives

```lox
// Add primitives to batch (returns index for later modification)
var idx1 = cube_batch.add(vec3(0, 0, 0), vec3(1, 1, 1), vec4(255, 0, 0, 255));
var idx2 = sphere_batch.add(vec3(2, 0, 0), vec3(0.5, 0.5, 0.5), vec4(0, 255, 0, 255));
```

### Updating Primitives

Update existing primitives by index:

```lox
// Update position, size, or color of existing entries
cube_batch.set_position(idx1, vec3(1, 2, 3));
cube_batch.set_size(idx1, vec3(2, 2, 2));
cube_batch.set_color(idx1, vec4(0, 0, 255, 255));

// Get current values
var pos = cube_batch.get_position(idx1);
var size = cube_batch.get_size(idx1);
var color = cube_batch.get_color(idx1);
```

### Rendering

#### Basic Rendering

```lox
// Render ALL primitives in the batch with a single draw call
cube_batch.draw();

// Typical usage in render loop:
win.begin_3d(camera);
cube_batch.draw();      // Renders all cubes in one call
sphere_batch.draw();    // Renders all spheres in one call
win.end_3d();
```

#### Frustum-culled Rendering

For large scenes, `draw_frustum_culled` automatically skips primitives outside the camera's field of view. It takes the camera's position and forward direction directly, rather than the `camera` object itself — the camera object has no getter for its current target/forward direction, so track it in Lox alongside `camera.set_target(...)`:

```lox
var cam_pos = vec3(15, 15, 15);
var cam_target = vec3(0, 0, 0);
var cam_forward = math.normalize3(cam_target - cam_pos);

win.begin_3d(camera);
cube_batch.draw_frustum_culled(cam_pos, cam_forward, 100.0, 60.0);  // max_distance=100, fov=60 degrees
win.end_3d();
```

- **`draw_frustum_culled(camera_position, camera_forward, max_distance, fov_degrees)`** - Culls objects beyond `max_distance` or outside a `fov_degrees`-wide cone in front of `camera_forward`. `camera_forward` **must be a unit vector** — normalize it yourself (`math.normalize3(target - position)`) before passing it in, since the culling math compares it directly against a cosine value.

### Batch Management

```lox
// Get information about the batch
var count = cube_batch.count();           // Number of entries
var capacity = cube_batch.reserve(5000);  // Pre-allocate space for 5000 entries

// Validation
var valid = cube_batch.is_valid_index(idx1); // Check if index exists

// Remove all entries
cube_batch.clear();
```

### Performance comparison

**Without batching** — one draw call per object, slow at scale:
```lox
foreach (i in range(len(positions))) {
    win.cube(positions[i].x, positions[i].y, positions[i].z, 1, 1, 1, colors[i]);
}
```

**With batching** — one draw call for the whole set:
```lox
var cube_batch = gfx.batch(win.BATCH_CUBE);
foreach (i in range(len(positions))) {
    cube_batch.add(positions[i], vec3(1, 1, 1), colors[i]);
}
cube_batch.draw();  // Single draw call for every cube added above
```

### Complete Example

```lox
import gfx;
import colour_utils;
import math;

var win = gfx.window(800, 600);
win.init();
var camera = gfx.camera(vec3(15, 15, 15), vec3(0, 0, 0), vec3(0, 1, 0));

var cube_batch = gfx.batch(win.BATCH_CUBE);
var indices = [];

// Add 100 cubes in a grid
for (var x = 0; x < 10; x = x + 1) {
    for (var z = 0; z < 10; z = z + 1) {
        var pos = vec3(x * 2, 0, z * 2);
        var size = vec3(1, 1, 1);
        var hue = (x + z) * 20;
        var color = colour_utils.hsv_to_rgb(hue, 0.8, 1.0);
        indices.append(cube_batch.add(pos, size, color));
    }
}

// Animation loop
while (!win.should_close()) {
    var time = sys.clock();

    // Update cube positions with a wave effect
    foreach (idx in indices) {
        var wave = math.sin(time + idx * 0.1) * 2;
        var old_pos = cube_batch.get_position(idx);
        cube_batch.set_position(idx, vec3(old_pos.x, wave, old_pos.z));
    }

    win.begin();
    win.clear(vec4(0, 0, 0, 255));
    win.begin_3d(camera);
    cube_batch.draw();  // Single call renders all 100 cubes
    win.end_3d();
    win.end();
}
```

**💡 Pro Tips:**
- Use batching for any scene with more than ~50 similar objects
- Pre-allocate batch capacity with `reserve()` for better performance
- Update only what changes - positions, colors, or sizes independently
- Combine multiple batch types in the same scene
- Clear batches between frames only if needed (reusing is more efficient)
- Use `draw_frustum_culled()` for large scenes to automatically skip off-screen objects

### Large-Scale Example

For scenes with thousands of objects, frustum culling keeps frame time down even when most of the scene is off-screen:

```lox
import gfx;
import colour_utils;
import random;
import math;

var CITY_SIZE = 50;  // 50x50 = 2500 buildings
var win = gfx.window(800, 600);
win.init();
var camera = gfx.camera(vec3(0, 20, 0), vec3(0, 0, 0), vec3(0, 1, 0));

var cube_batch = gfx.batch(win.BATCH_CUBE);
cube_batch.reserve(CITY_SIZE * CITY_SIZE);  // Pre-allocate for performance

for (var x = 0; x < CITY_SIZE; x = x + 1) {
    for (var z = 0; z < CITY_SIZE; z = z + 1) {
        var height = random.integer(2, 16);  // Random building height
        var pos = vec3(x * 4 - CITY_SIZE * 2, height / 2, z * 4 - CITY_SIZE * 2);
        var size = vec3(1.5, height, 1.5);
        var color = colour_utils.hsv_to_rgb(random.integer(0, 359), 0.7, 0.9);
        cube_batch.add(pos, size, color);
    }
}

// Animation loop with automatic culling
var cam_target = vec3(0, 10, 0);
while (!win.should_close()) {
    var time = sys.clock();

    // Fly the camera through the city
    var radius = CITY_SIZE * 1.5;
    var cam_pos = vec3(math.cos(time * 0.3) * radius, 25 + math.sin(time * 0.5) * 10, math.sin(time * 0.3) * radius);
    camera.set_position(cam_pos);
    camera.set_target(cam_target);

    win.begin();
    win.clear(vec4(0.1, 0.1, 0.2, 1.0));
    win.begin_3d(camera);
    // draw_frustum_culled takes the camera's position/forward directly, not
    // the camera object -- see the note in the Batch Object section above.
    cube_batch.draw_frustum_culled(cam_pos, math.normalize3(cam_target - cam_pos), radius * 2.0, 60.0);
    win.end_3d();
    win.end();
}
```

---

## Texture Object

Textures are used for storing and displaying 2D image data.

### Texture Creation

```lox
import gfx;

var img = gfx.image("filename.png");
var tex = gfx.texture(img, frames, start_frame, end_frame);
```

- **`gfx.image(filename)`** - Load an image from file
- **`gfx.texture(image, frames, start_frame, end_frame)`** - Create a texture from an image with animation support

### Texture Methods

- **`width()`** - Returns texture width in pixels
- **`height()`** - Returns texture height in pixels
- **`frame_width()`** - Returns frame width (for animated textures)
- **`animate(frame_time)`** - Set automatic frame animation (ticks per frame)
- **`set_wrap_mode(mode)`** - Set the texture's wrap mode, one of `win.WRAP_REPEAT`, `win.WRAP_CLAMP`, `win.WRAP_MIRROR_REPEAT`, `win.WRAP_MIRROR_CLAMP`
- **`unload()`** - Free the texture's GPU resources

---

## RenderTexture Object

RenderTextures allow rendering to an off-screen buffer that can be used as a texture.

### RenderTexture Creation

```lox
import gfx;

var rt = gfx.render_texture(width, height);
```

### RenderTexture Methods

- **`width()`** - Returns render texture width
- **`height()`** - Returns render texture height
- **`get_texture()`** - Returns the underlying texture, for passing to `draw_texture`/`textured_cube`/etc.
- **`unload()`** - Free the render texture's GPU resources

#### Drawing methods (same as window, but targeting the render texture)
- **`line(x1, y1, x2, y2, color_vec4)`**
- **`line_ex(x1, y1, x2, y2, thickness, color_vec4)`**
- **`rectangle(x, y, width, height, color_vec4)`**
- **`circle_fill(x, y, radius, color_vec4)`**
- **`circle(x, y, radius, color_vec4)`**
- **`triangle(x1, y1, x2, y2, x3, y3, color_vec4)`**
- **`pixel(x, y, color_vec4)`**
- **`text(text, x, y, size, color_vec4)`**
- **`draw_texture(texture, x, y, color_vec4)`**
- **`draw_texture_pro(texture, src_x, src_y, src_w, src_h, dest_x, dest_y, dest_w, dest_h, origin_x, origin_y, rotation, color_vec4)`**
- **`draw_array_fast(float_array)`** - Faster variant of `window.draw_array` for use inside `begin_texture_mode`/`end_texture_mode`
- **`clear(color_vec4)`** - Clear the render texture with specified color

---

## Camera Object

Cameras define the viewpoint for 3D rendering.

### Camera Creation

```lox
import gfx;

var cam = gfx.camera(position_vec3, target_vec3, up_vec3);
```

### Camera Methods

- **`set_position(vec3_position)`** - Set camera position using a vec3
- **`get_position()`** - Get position vec3
- **`set_target(vec3_target)`** - Set camera target (what it's looking at) using a vec3
- **`set_fovy(field_of_view)`** - Set field of view in degrees
- **`update()`** - Update camera (enables free-look controls)

---

## Shader Object

Shaders allow custom GPU programs for advanced rendering effects.

### Shader Creation

```lox
import gfx;

var shdr = gfx.shader();
```

### Shader Methods

- **`load_from_memory(vertex_shader_code, fragment_shader_code)`** - Load shader from source code strings
- **`get_location(uniform_name)`** - Get location of a uniform variable
- **`set_value_float(location, value)`** - Set a float uniform value
- **`set_value_vec2(location, vec2_value)`** - Set a Vec2 uniform value
- **`set_value_vec3(location, vec3_value)`** - Set a Vec3 uniform value
- **`set_value_vec4(location, vec4_value)`** - Set a Vec4 uniform value
- **`is_valid()`** - Check if shader compiled successfully
- **`unload()`** - Free shader resources

---

## Image Object

Images represent raw pixel data that can be loaded and manipulated.

### Image Creation

```lox
import gfx;

var img = gfx.image("filename.png");
```

### Image Methods

- **`width()`** - Returns image width in pixels
- **`height()`** - Returns image height in pixels

---

## FloatArray Object

FloatArrays store 2D arrays of floating-point values, useful for mathematical computations and data visualization.

### FloatArray Creation

```lox
import gfx;

var arr = gfx.float_array(width, height);
```

### FloatArray Methods

- **`width()`** - Returns array width
- **`height()`** - Returns array height
- **`get(x, y)`** - Get value at coordinates (x, y)
- **`set(x, y, value)`** - Set value at coordinates (x, y)
- **`clear(value)`** - Fill entire array with specified value

---

## PhysicsWorld Object

`physics_world` is a native 3D rigid-body simulation for spheres: body
storage, gravity integration, boundary bounce, broad-phase collision culling
(uniform grid), and impulse-based collision resolution all run natively in
Go instead of per-object Lox method calls. It's a good fit for scenes with
many identical, roughly-equal-mass moving bodies (particle bursts, ball
pits, debris) where the Lox-level cost of a class + per-object `update()`
call becomes the hot path. See `lox_examples/3d_balls_physics_shaders.lox`
for a full example (spawning, rendering, and explosion forces on top of a
`physics_world`), and `docs/plans/PLAN_physics_world.md` for the design rationale.

Fixed level geometry (floors, shelves, ramps) can be added as static boxes
with `add_static_box` — optionally tilted, never moving — for dynamic
spheres to bounce off. See `docs/md/plans/NOTES_box_physics.md` for the
design rationale behind the shape/static/rotation split.

Gameplay policy — deciding *when* something explodes, *where* the blast
center is, and the distance/falloff curve — stays in Lox; only the low-level
simulation (integration, collision, and the `add_impulse` velocity nudge)
lives in the native type.

### PhysicsWorld Creation

```lox
import physics;

var world = physics.physics_world(min_vec3, max_vec3, cell_size, gravity_vec3);
```

- `min_vec3` / `max_vec3` - Opposite corners of the simulation's boundary box. Bodies bounce off these bounds (each body's own radius is accounted for automatically).
- `cell_size` - Cell size for the internal broad-phase grid. Should exceed the largest body's diameter.
- `gravity_vec3` - Constant world-space acceleration applied to every body every `step()`.

### PhysicsWorld Methods

- **`add_material(restitution, friction, damping)`** - Register a material and return its integer id. `restitution` controls bounciness (both boundary bounces and body-body collisions, combined via `sqrt(a * b)` when two different materials collide); `damping` is a per-step velocity multiplier (air resistance); `friction` is accepted but not yet applied to collision response.
- **`add(pos_vec3, vel_vec3, radius, material_id)`** - Add a dynamic sphere body and return its integer id (a stable handle used by every other method).
- **`add_static_box(pos_vec3, half_extents_vec3, axis_vec3, angle, material_id)`** - Add a fixed box that never moves, optionally tilted `angle` degrees around `axis` (`angle = 0` for axis-aligned). No velocity argument — fixed-ness is unconditional. For floors, shelves, ramps.
- **`get_box_transform(id)`** - For a static box body, returns `(position, half_extents, axis, angle)` as a tuple. Feed straight into `win.cube_rotated()` so drawing always matches exactly what physics collided against.
- **`remove(id)`** - Remove a body from the simulation. Ids are tombstoned, not reused.
- **`get_position(id)`** - Get a body's current position as a vec3.
- **`add_impulse(id, impulse_vec3)`** - Add an instantaneous velocity change (`vel += impulse_vec3`) to a body. This is the primitive for one-off forces such as explosions — compute the direction/falloff in Lox, then call this once per affected body.
- **`step(dt)`** - Advance the simulation: gravity integration, boundary bounce, and collision resolution, all in one native call.
- **`collisions()`** - Returns a list of tuples `(a_id, b_id, normal_vec3, impulse)`, one per body pair that **newly** started touching during the last `step()` call. Pairs that are still resting/touching from a previous frame are not repeated.
- **`count()`** - Returns the number of currently active (non-removed) bodies.

Static bodies bypass the broad-phase grid entirely instead of being inserted
into it — a single grid cell can't represent a large platform's true extent,
so static bodies are checked directly against every dynamic sphere each
`step()` instead. This means `cell_size` only needs to exceed the largest
*dynamic* sphere's diameter; static geometry can be arbitrarily large. Two
static bodies never collide with each other.

### PhysicsWorld Example

```lox
import physics;

var world = physics.physics_world(vec3(-10, 0, -10), vec3(10, 100, 10), 2.0, vec3(0, -0.01, 0));
var mat = world.add_material(0.5, 0.3, 0.99);
var id = world.add(vec3(0, 5, 0), vec3(0.1, 0, 0), 0.5, mat);

while (true) {
    world.step(1.0);
    print world.get_position(id);

    foreach (pair in world.collisions()) {
        a_id, b_id, normal, impulse = pair;
        print "collision: " & str(a_id) & " " & str(b_id) & " " & str(normal) & " " & str(impulse);
    }
}
```

---

## Vector Objects

Vector objects represent mathematical vectors for 2D, 3D, and 4D operations — `vec2`/`vec3`/`vec4` are always global, no import needed.

### Vector Creation

```lox
var v2 = vec2(x, y);
var v3 = vec3(x, y, z);
var v4 = vec4(x, y, z, w);
```

### Vector properties
```lox
v2.x = 1;
v2.y = 2;
```

### Vector addition 
```lox
var v1 = vec3(1, 3, 4);
var v2 = vec3(2, 4, 5);
var v3 = v2 ++ v1;
```

### In-place addition

Each vector type also has an `.add(other)` method that mutates the receiver's components directly instead of allocating a new vector — useful in hot loops (e.g. updating an entity's position every frame).

```lox
var pos = vec2(0, 0);
var dp = vec2(1, 1);
pos.add(dp);        // pos becomes vec2(1, 1); dp is unchanged
```

**Caution:** unlike `++` (which always allocates a fresh vector), `.add()` mutates the existing object, so any other reference to that same vector — e.g. a value stored via `this.pos = pos` in a constructor, when the caller kept its own reference to the argument — is mutated too. Prefer `.add()` only where the receiver is known not to be aliased elsewhere (a field set from an unshared literal, for example); otherwise use `++`.

---

## System Modules

### sys Module

The sys module provides system-level functionality (accessed via `sys.function_name`). No import needed for `sys` — it's always available.

- **`sys.args()`** - Returns command line arguments
- **`sys.clock()`** - Returns elapsed time in seconds since the interpreter started (for timing/animation, not a real timestamp)
- **`sys.sleep(seconds)`** - Pauses execution for the specified number of seconds
- **`sys.today()`** - Returns the current date as a string, `"YYYY-MM-DD"`
- **`sys.now()`** - Returns the current time as a string, `"HH:MM:SS"` (no date — pair with `sys.today()` for both; used by the [`logging`](#system-modules) module's timestamps)

File I/O is not part of `sys` — see [`os`](OS_MODULE.md) for `open`/`close`/`readln`/`write`/`read_all` and the rest of the filesystem API.

### random Module

`import random` — extends the global [`rand()`](#random--color-functions) with common patterns:

- **`random.integer(min, max)`** - Random integer in `[min, max]` (inclusive both ends)
- **`random.float(min, max)`** - Random float in `[min, max)`
- **`random.choice(list)`** - Randomly select one element from a list

### inspect Module

The inspect module provides debugging and introspection capabilities (accessed via `inspect.function_name`).

```lox
import inspect;

inspect.dump_frame();
```
Prints the current frame's name, stack, and locals/globals to stdout.

```lox
var d = inspect.get_frame();
```
Returns a dict describing the current frame, with keys:
- `function` - function name
- `line` - current line
- `file` - current script
- `args` - list of arguments
- `locals` - dict of locals
- `globals` - dict of globals
- `prev_frame` - calling frame's dict (or `nil`)

---

## Color Utilities Module

The `colour_utils` module provides native high-performance color manipulation functions (accessed via `colour_utils.function_name`).

### Color Utility Functions

- **`colour_utils.fade(r, g, b, alpha)`** - Apply alpha transparency to RGB values
  - `r, g, b`: RGB color components (0-255)
  - `alpha`: Alpha value (0.0 to 1.0)
  - Returns: vec4 with alpha applied to RGB components

- **`colour_utils.tint(r1, g1, b1, r2, g2, b2)`** - Tint a color with another color
  - `r1, g1, b1`: Base RGB color components (0-255)
  - `r2, g2, b2`: Tint RGB color components (0-255)
  - Returns: vec4 with tinted color (multiplies RGB components)

- **`colour_utils.brightness(r, g, b, factor)`** - Adjust brightness of RGB values
  - `r, g, b`: RGB color components (0-255)
  - `factor`: Brightness factor (1.0 = normal, >1.0 = brighter, <1.0 = darker)
  - Returns: vec4 with adjusted brightness

- **`colour_utils.lerp(r1, g1, b1, r2, g2, b2, amount)`** - Linear interpolation between two colors
  - `r1, g1, b1`: First RGB color components (0-255)
  - `r2, g2, b2`: Second RGB color components (0-255)
  - `amount`: Interpolation amount (0.0 to 1.0)
  - Returns: vec4 with interpolated color

- **`colour_utils.hsv_to_rgb(h, s, v)`** - Convert HSV to RGB color
  - `h`: Hue (0 to 360 degrees)
  - `s`: Saturation (0.0 to 1.0)
  - `v`: Value/Brightness (0.0 to 1.0)
  - Returns: vec4 with RGB color converted from HSV

- **`colour_utils.random()`** - Generate a random color
  - Returns: vec4 with random RGB color values

### Usage Example

```lox
import colour_utils;
import gfx;

var win = gfx.window(800, 600);
win.init();

// Apply effects directly with RGB values - all functions return vec4s
var faded_red = colour_utils.fade(255, 0, 0, 0.5);
var purple = colour_utils.lerp(255, 0, 0, 0, 0, 255, 0.5);
var bright_red = colour_utils.brightness(255, 0, 0, 1.5);
var tinted = colour_utils.tint(255, 0, 0, 0, 255, 0);
var random_color = colour_utils.random();

// Convert HSV to RGB
var orange = colour_utils.hsv_to_rgb(30, 1.0, 1.0);

// Use directly with graphics functions (no need to decode)
win.circle_fill(100, 100, 50, purple);
win.rectangle(50, 50, 100, 100, faded_red);
win.text("Colorful!", 10, 10, 20, orange);
```

---

## Example Usage

### Basic Graphics Program

```lox
import gfx;

var win = gfx.window(800, 600);
win.init();
win.set_target_fps(60);

while (!win.should_close()) {
    win.begin();
    win.clear(vec4(50, 50, 50, 255));

    win.circle_fill(400, 300, 50, vec4(255, 0, 0, 255));
    win.text("Hello, GLox!", 350, 200, 20, vec4(255, 255, 255, 255));

    win.end();
}

win.close();
```

### 3D Rendering Example

```lox
import gfx;

var win = gfx.window(800, 600);
var cam = gfx.camera(vec3(5, 5, 5), vec3(0, 0, 0), vec3(0, 1, 0));

win.init();
cam.set_fovy(45);

while (!win.should_close()) {
    cam.update();

    win.begin();
    win.clear(vec4(100, 150, 200, 255));

    win.begin_3d(cam);
    win.cube(0, 0, 0, 2, 2, 2, vec4(255, 0, 0, 255));
    win.grid(10, 1);
    win.end_3d();

    win.end();
}

win.close();
```

### Mandelbrot Set Visualization

```lox
import gfx;

var width = 800;
var height = 600;
var win = gfx.window(width, height);

win.init();

var mandel_data = gfx.lox_mandel_array(width, height, 1.0, -0.5, 0.0, 100);

while (!win.should_close()) {
    win.begin();
    win.draw_array(mandel_data);
    win.end();
}

win.close();
```
