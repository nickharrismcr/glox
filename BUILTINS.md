# GLox Builtin Functions and Objects Documentation

 

## Table of Contents

1. [Builtin Functions](#builtin-functions)
2. [Window Object](#window-object)
3. [Texture Object](#texture-object)
4. [RenderTexture Object](#rendertexture-object)
5. [Camera Object](#camera-object)
6. [Shader Object](#shader-object)
7. [Image Object](#image-object)
8. [FloatArray Object](#floatarray-object)
9. [Vector Objects](#vector-objects)
10. [File Operations](#file-operations)
11. [System Modules](#system-modules)

---

## Builtin Functions

### Core Functions

- **`args()`** - Returns command line arguments
- **`clock()`** - Returns current time in seconds
- **`type(value)`** - Returns the type of a value as a string
- **`len(container)`** - Returns the length of a container (string, list, etc.)
- **`sleep(seconds)`** - Pauses execution for the specified number of seconds

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

### Graphics Functions

- **`encode_rgb(r, g, b)`** - Encodes RGB values (0-255) into a single integer
- **`decode_rgb(color)`** - Decodes an RGB integer into [r, g, b] components
- **`draw_png(filename, width, height, data)`** - Writes PNG image data to file

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
- **`ellipse(center_x, center_y, center_z, radius_x, radius_z, color_vec4)`** - Draw a 3D ellipse
- **`text(text, x, y, size, color_vec4)`** - Draw text

#### Advanced Drawing
- **`draw_array(float_array)`** - Draw a float array as grayscale image
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
- **`sphere(x, y, z, radius, color_vec4)`** - Draw a 3D sphere
- **`cylinder(x, y, z, radius_top, radius_bottom, height, slices, color)`** - Draw a 3D cylinder
- **`grid(slices, spacing)`** - Draw a 3D grid
- **`plane(x, y, z, width, length, color)`** - Draw a 3D plane

### Shader Support
- **`begin_shader_mode(shader)`** - Begin custom shader mode
- **`end_shader_mode()`** - End custom shader mode

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

## File Operations

### File I/O Functions

- **`open(filename, mode)`** - Open a file for reading/writing
  - Modes: "r" (read), "w" (write), "a" (append)
- **`close(file)`** - Close an open file
- **`readln(file)`** - Read a line from file
- **`write(file, text)`** - Write text to file

### Example File Usage

```lox
var file = open("data.txt", "w");
write(file, "Hello, World!\n");
close(file);

file = open("data.txt", "r");
var line = readln(file);
print line;
close(file);
```

---

## System Modules

### sys Module

The sys module provides system-level functionality (accessed via sys.function_name).

### inspect Module

The inspect module provides debugging and introspection capabilities (accessed via inspect.function_name).

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

 
 

 