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
10. [Vector Objects](#vector-objects)
11. [File Operations](#file-operations)
12. [System Modules](#system-modules)
13. [Color Utilities Module](#color-utilities-module)

---

## Builtin Functions

### Core Functions


- **`type(value)`** - Returns the type of a value as a string
- **`len(container)`** - Returns the length of a container (string, list, etc.)


### Mathematical Functions

- **`sin(x)`** - Returns the sine of x (in radians)
- **`cos(x)`** - Returns the cosine of x (in radians)
- **`sqrt(x)`** - Returns the square root of x
- **`atan2(y, x)`** - Returns the arctangent of y/x in radians
- **`rand()`** - Returns a random number between 0 and 1

### Type Conversion Functions

- **`float(value)`** - Converts value to floating point number
- **`int(value)`** - Converts value to integer

### Container Functions

- **`append(list, item)`** - Appends an item to a list
- **`range(n)`** - Creates a range iterator from 0 to n-1

### String Functions

- **`replace(string, old, new)`** - Replaces occurrences of 'old' with 'new' in string

### Color Functions

- **`encode_rgb(r, g, b)`** - Encodes RGB values (0-255) into a single integer
- **`decode_rgb(color)`** - Decodes an RGB integer into [r, g, b] components

**Note:** For advanced color manipulation (fade, tint, brightness, HSV conversion, etc.), see the [Color Utilities Module](#color-utilities-module).

### Graphics Functions

- **`draw_png(filename, width, height, data)`** - Writes PNG image data to file
- **`batch(type)`** - Creates a new batch object for optimized rendering of primitives

### Special Functions

- **`lox_mandel_array(width, height, zoom, center_x, center_y, max_iter)`** - Generates Mandelbrot set data as a float array

---

## Window Object

The window object provides the main interface for graphics rendering and input handling.

### Window Creation

```lox
var win = window(width, height);
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
- **`draw_texture_rect(texture, x, y, src_x, src_y, src_w, src_h, color_vec4)`** - Draw part of a texture
- **`draw_render_texture(render_texture, x, y, color_vec4)`** - Draw a render texture
- **`draw_render_texture_ex(render_texture, x, y, rotation, scale, color_vec4)`** - Draw render texture with transformation

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
- **`win.BATCH_TEXTURED_CUBE`** - For textured cube batch creation
- **`win.BATCH_SPHERE`** - For sphere batch creation
- **`win.BATCH_PLANE`** - For plane batch creation

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
 
var cube_batch = batch(win.BATCH_CUBE);      // Create a batch for cubes
var sphere_batch = batch(win.BATCH_SPHERE);  // Create a batch for spheres
var plane_batch = batch(win.BATCH_PLANE);    // Create a batch for planes

 
```

**Supported batch types:**
- `win.BATCH_CUBE`  - For rendering cubes
- `win.BATCH_TEXTURED_CUBE` - For rendering textured cubes
- `win.BATCH_SPHERE`   - For rendering spheres  
- `win.BATCH_PLANE`   - For rendering planes

### Adding Primitives

```lox
// Add primitives to batch (returns index for later modification)
var index = cube_batch.add(position, size, color);

// For textured cubes, use add_textured_cube method
var textured_batch = batch(win.BATCH_TEXTURED_CUBE);
var textured_index = textured_batch.add_textured_cube(texture, position, size, base_color);

// Examples:
var idx1 = cube_batch.add(vec3(0, 0, 0), vec3(1, 1, 1), vec4(255, 0, 0, 255));
var idx2 = sphere_batch.add(vec3(2, 0, 0), vec3(0.5, 0.5, 0.5), vec4(0, 255, 0, 255));
var idx3 = plane_batch.add(vec3(0, -1, 0), vec3(10, 1, 10), vec4(100, 100, 100, 255));

// Textured cube example:
var my_texture = texture(image("myimage.png"), 1, 1, 1);
var idx4 = textured_batch.add_textured_cube(my_texture, vec3(0, 0, 0), vec3(1, 1, 1), vec4(255, 255, 255, 255));
```

### Updating Primitives

Update existing primitives by index:

```lox
// Update position, size, or color of existing entries
cube_batch.set_position(index, vec3(1, 2, 3));
cube_batch.set_size(index, vec3(2, 2, 2));
cube_batch.set_color(index, vec4(0, 0, 255, 255));

// Get current values
var pos = cube_batch.get_position(index);
var size = cube_batch.get_size(index);
var color = cube_batch.get_color(index);
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
plane_batch.draw();     // Renders all planes in one call
win.end_3d();
```

#### Optimized Rendering with Culling

For large scenes, use culling methods to automatically skip primitives that are too far away or outside the camera's view:

```lox
// Distance-based culling - skips objects beyond max_distance
cube_batch.draw_culled(camera, max_distance);

// Frustum culling - skips objects outside camera's field of view
cube_batch.draw_frustum_culled(camera);

// Example usage:
win.begin_3d(camera);
cube_batch.draw_frustum_culled(camera);  // Only renders visible cubes
sphere_batch.draw_culled(camera, 50.0);  // Only renders spheres within 50 units
win.end_3d();
```

**Culling Method Parameters:**
- **`draw_culled(camera, max_distance)`** - Culls objects beyond `max_distance` from camera position
- **`draw_frustum_culled(camera)`** - Culls objects outside the camera's viewing frustum (field of view)

**Performance Benefits:**
- Distance culling: Skip distant objects to maintain frame rate
- Frustum culling: Only render what's actually visible on screen
- Automatic optimization: No need to manually track object visibility

### Batch Management

```lox
// Get information about the batch
var count = cube_batch.count();           // Number of entries
var capacity = cube_batch.capacity();     // Current capacity

// Memory management
cube_batch.reserve(5000);                 // Pre-allocate space for 5000 entries
cube_batch.clear();                       // Remove all entries

// Validation
var valid = cube_batch.is_valid_index(idx); // Check if index exists
```

### Performance Benefits

**Without Batching (Traditional):**
```lox
// 1000 individual draw calls - SLOW!
for (var i = 0; i < 1000; i = i + 1) {
    win.cube(positions[i], sizes[i], colors[i]);  // 1000 draw calls
}
```

**With Batching (Optimized):**
```lox
// 1 optimized draw call - FAST!
var cube_batch = batch("cube");
for (var i = 0; i < 1000; i = i + 1) {
    cube_batch.add(positions[i], sizes[i], colors[i]);  // Add to batch
}
cube_batch.draw();  // Single draw call for ALL 1000 cubes!
```

**Performance Comparison:**
- **1,000 cubes:** 1,000 draw calls → 1 draw call (1000x improvement)
- **10,000 cubes:** 10,000 draw calls → 1 draw call (10,000x improvement)
- **Real-world result:** 10,000+ animated cubes running at 60+ FPS

### Complete Example

```lox
import colour_utils;

// Create batch and add cubes
var cube_batch = batch(win.BATCH_CUBE);
var indices = [];

// Add 100 cubes in a grid
for (var x = 0; x < 10; x = x + 1) {
    for (var z = 0; z < 10; z = z + 1) {
        var pos = vec3(x * 2, 0, z * 2);
        var size = vec3(1, 1, 1);
        var hue = (x + z) * 20;
        var color = colour_utils.hsv_to_rgb(hue, 0.8, 1.0);
        
        var idx = cube_batch.add(pos, size, color);
        indices.append(idx);
    }
}

// Animation loop
while (!win.should_close()) {
    var time = sys.clock();
    
    // Update cube positions with wave effect
    for (var i = 0; i < len(indices); i = i + 1) {
        var wave = sin(time + i * 0.1) * 2;
        var oldPos = cube_batch.get_position(indices[i]);
        var newPos = vec3(oldPos.x, wave, oldPos.z);
        cube_batch.set_position(indices[i], newPos);
    }
    
    // Render
    win.begin();
    win.clear(vec4(0, 0, 0, 255));
    win.begin_3d(camera);
    
    cube_batch.draw();  // Single call renders all 100 cubes!
    
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
- Use `draw_culled()` with distance limits to maintain smooth frame rates

### Large-Scale Culling Example

For scenes with thousands of objects, use culling to maintain performance:

```lox
import colour_utils;

var CITY_SIZE = 50;  // 50x50 = 2500 buildings
var camera = camera_3d(vec3(0, 20, 0), vec3(0, 0, 0), vec3(0, 1, 0));

// Create batch and generate a city
var cube_batch = batch(win.BATCH_CUBE);
cube_batch.reserve(CITY_SIZE * CITY_SIZE);  // Pre-allocate for performance

for (var x = 0; x < CITY_SIZE; x = x + 1) {
    for (var z = 0; z < CITY_SIZE; z = z + 1) {
        var height = rand() % 15 + 2;  // Random building height
        var pos = vec3(x * 4 - CITY_SIZE * 2, height / 2, z * 4 - CITY_SIZE * 2);
        var size = vec3(1.5, height, 1.5);
        var color = colour_utils.hsv_to_rgb(rand() % 360, 0.7, 0.9);
        
        cube_batch.add(pos, size, color);
    }
}

print("Generated " + cube_batch.count() + " buildings");

// Animation loop with automatic culling
while (!win.should_close()) {
    var time = sys.clock();
    
    // Fly camera through the city
    var radius = CITY_SIZE * 1.5;
    var cam_x = cos(time * 0.3) * radius;
    var cam_z = sin(time * 0.3) * radius;
    var cam_y = 25 + sin(time * 0.5) * 10;
    
    camera.position = vec3(cam_x, cam_y, cam_z);
    camera.target = vec3(0, 10, 0);
    
    // Render with automatic frustum culling
    win.begin();
    win.clear(vec4(0.1, 0.1, 0.2, 1.0));
    win.begin_3d(camera);
    
    // Only renders buildings visible in camera's field of view
    cube_batch.draw_frustum_culled(camera);
    
    win.end_3d();
    win.end();
}
```

**Culling Results:**
- **2,500 buildings:** Typically only 200-400 rendered (80%+ culling efficiency)
- **Performance:** Maintains 60+ FPS even with complex scenes
- **Automatic:** No manual distance checking or visibility calculations needed

---

## Texture Object

Textures are used for storing and displaying 2D image data.

### Texture Creation

```lox
var img = image("filename.png");
var tex = texture(img, frames, start_frame, end_frame);
```

- **`image(filename)`** - Load an image from file
- **`texture(image, frames, start_frame, end_frame)`** - Create a texture from an image with animation support

### Texture Methods

- **`width()`** - Returns texture width in pixels
- **`height()`** - Returns texture height in pixels
- **`frame_width()`** - Returns frame width (for animated textures)
- **`animate(frame_time)`** - Set automatic frame animation ( ticks per frame )

---

## RenderTexture Object

RenderTextures allow rendering to an off-screen buffer that can be used as a texture.

### RenderTexture Creation

```lox
var rt = render_texture(width, height);
```

### RenderTexture Methods

- **`width()`** - Returns render texture width
- **`height()`** - Returns render texture height
- **`clear(color_vec4)`** - Clear the render texture with specified color

#### Drawing Methods (same as window but to render texture)
- **`line(x1, y1, x2, y2, color_vec4)`** - Draw line to render texture
- **`line_ex(x1, y1, x2, y2, thickness, color_vec4)`** - Draw thick line to render texture
- **`rectangle(x, y, width, height, color_vec4)`** - Draw rectangle to render texture
- **`circle_fill(x, y, radius, color_vec4)`** - Draw filled circle to render texture
- **`circle(x, y, radius, color_vec4)`** - Draw circle outline to render texture
- **`pixel(x, y, color_vec4)`** - Draw pixel to render texture

---

## Camera Object

Cameras define the viewpoint for 3D rendering.

### Camera Creation

```lox
var cam = camera(position_vec3, target_vec3, up_vec3);
```

### Camera Methods

- **`set_position(vec3_position)`** - Set camera position using a Vec3
- **`set_target(vec3_target)`** - Set camera target (what it's looking at) using a Vec3
- **`set_fovy(field_of_view)`** - Set field of view in degrees
- **`update()`** - Update camera (enables free-look controls)

---

## Shader Object

Shaders allow custom GPU programs for advanced rendering effects.

### Shader Creation

```lox
var shader = shader();
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
var img = image("filename.png");
```

### Image Methods

- **`width()`** - Returns image width in pixels
- **`height()`** - Returns image height in pixels

---

## FloatArray Object

FloatArrays store 2D arrays of floating-point values, useful for mathematical computations and data visualization.

### FloatArray Creation

```lox
var arr = float_array(width, height);
```

### FloatArray Methods

- **`width()`** - Returns array width
- **`height()`** - Returns array height
- **`get(x, y)`** - Get value at coordinates (x, y)
- **`set(x, y, value)`** - Set value at coordinates (x, y)
- **`clear(value)`** - Fill entire array with specified value

---

## Vector Objects

Vector objects represent mathematical vectors for 2D, 3D, and 4D operations.

### Vector Creation

```lox
var v2 = vec2(x, y);
var v3 = vec3(x, y, z);
var v4 = vec4(x, y, z, w);
```

Vectors support standard mathematical operations and are used extensively in 3D graphics and physics calculations.

---



## System Modules

### sys Module

The sys module provides system-level functionality (accessed via sys.function_name).

`import sys` 

- **`sys.args()`** - Returns command line arguments
- **`sys.clock()`** - Returns current time in seconds
- **`sys.sleep(seconds)`** - Pauses execution for the specified number of seconds

## File Operations

### File I/O Functions

- **`sys.open(filename, mode)`** - Open a file for reading/writing
  - Modes: "r" (read), "w" (write), "a" (append)
- **`sys.close(file)`** - Close an open file
- **`sys.readln(file)`** - Read a line from file
- **`sys.write(file, text)`** - Write text to file

### Example File Usage

```lox
var file = sys.open("data.txt", "w");
sys.write(file, "Hello, World!\n");
sys.close(file);

file = sys.open("data.txt", "r");
var line = sys.readln(file);
print line;
sys.close(file);
```

---

### inspect Module

The inspect module provides debugging and introspection capabilities (accessed via inspect.function_name).

```
import inspect

inspect.dump_frame() 
```
- print current frame name, stack/locals, globals 

`d=inspect.get_frame()` returns frame data dictionary with keys:
`function`   - function name 
`line`       - current line
`file`       - current script 
`args`       - list of arguments
`locals`     - dictionary of locals
`globals`    - dictionary of globals 
`prev_frame` - calling frame dict (or nil) 


---

## Color Utilities Module

The `colour_utils` module provides native high-performance color manipulation functions (accessed via colour_utils.function_name).

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
var win = window(800, 600);
win.init();
win.set_target_fps(60);

while (!win.should_close()) {
    win.begin();
    win.clear(vec4(50, 50, 50, 255));
    
    // Draw a filled circle
    win.circle_fill(400, 300, 50, vec4(255, 0, 0, 255));
    
    // Draw text
    win.text("Hello, GLox!", 350, 200, 20, vec4(255, 255, 255, 255));
    
    win.end();
}

win.close();
```

### 3D Rendering Example

```lox
var win = window(800, 600);
var cam = camera(vec3(5, 5, 5), vec3(0, 0, 0), vec3(0, 1, 0));

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
var width = 800;
var height = 600;
var win = window(width, height);

win.init();

var mandel_data = lox_mandel_array(width, height, 1.0, -0.5, 0.0, 100);

while (!win.should_close()) {
    win.begin();
    win.draw_array(mandel_data);
    win.end();
}

win.close();
```

 
 

 