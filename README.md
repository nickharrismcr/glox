# GLOX

**Bob Nystroms CLox bytecode interpreter implemented in Go**

> **Note for AI/Copilot**: Read `.copilot-instructions.md` and `TERMINAL_INSTRUCTIONS.md` before assisting with this project.

The aim of this project is to learn more deeply about programming in Go and the crafting of interpreters by way of implementing Bobs CLox interpreter in Go, adding Python-inspired extensions to Lox along the way.
The extensions to the language include enhanced string operations, lists, dictionaries, exception handling, module imports with bytecode caching, string and list iteration, Raylib bindings for graphics, and I/O.  


### Additions to vanilla Lox:

#### Module Imports

##### Syntax
```lox
import modulename
import modulename as alias
import modulename1, modulename2 as alias2
```
##### Example
```lox
import math
import graphics as gfx
import random, color as clr
```
- Modules are cached as bytecode after first import for fast loading.
- Aliases allow you to refer to a module with a different name.
- Compiled modules are stored in `__loxcache__/<module>.lxc` and reloaded unless the source is newer.
- `sys` module for system functions (args, clock)
- `os` module for file and directory operations
- `inspect` module for vm state dumps
---

#### Variable Declarations

##### Implicit Declaration
```lox
a = 1
```
No `var` required; assignment creates the variable.

##### Immutable Variables
```lox
const PI = 3.14159
```
`const` creates a variable that cannot be reassigned.

---

#### Numeric Types

##### Integer with Modulus
```lox
a = 10
b = 3
c = a % b   // c is 1
```

---

#### Control Flow

##### Break and Continue
```lox
for (var i = 0; i < 10; i = i + 1) {
    if (i == 5) break
    if (i % 2 == 0) continue
    print i
}
```

##### Foreach Loop
```lox
foreach (item in [1, 2, 3]) {
    print item
}
```
- Works with lists, strings, and any object implementing `__iter__` and `__next__`.
- The iterable's `next` method is called until end is reached.
- Iterables can be native lists/strings or Lox classes that implement `__iter__` (returning an iterator that implements `__next__` and returns a value or nil for end).

##### Range
```lox
foreach (i in range(0, 10, 2)) {
    print i
}
```
- `range(start, end, step)` returns an efficient integer iterator.
- `foreach ( a in range... )` is much faster than the equivalent for loop.s

---

#### Compound numeric assignment+add/subtract Operators

##### Syntax
```lox
a+=1
obj.x+=2
```
##### Example
```lox
a = 5
a+=1
print a // 6

point = vec2(1, 2)
point.x+=1
print point.x // 2
```

---

#### Native Functions

##### Type Conversion
```lox
int("42")      // 42
float("3.14")  // 3.14
str(123)       // "123"
```

##### Length
```lox
len([1,2,3])   // 3
len("hello")   // 5
```

##### Math
```lox
sin(3.14)
cos(0)
```

##### Command-line Arguments
```lox
sys.args() // returns list of command-line arguments
```

##### Type Inspection
```lox
type(123)      // "int"
type([1,2,3])  // "list"
```

##### PNG Drawing
```lox
draw_png("out.png", float_array)
```

##### RGB Encoding/Decoding
```lox
f = encode_rgb(255, 128, 0)
r, g, b = decode_rgb(f)
```

---

#### Native Objects

##### Fast 2D Float Array
```lox
a = float_array(100, 100)
a.set(10, 10, 0.5)
b = a.get(10, 10)
```

##### Fast Native Vectors
```lox
v = vec2(1, 2)
v3 = vec3(1, 2, 3)
v4 = vec4(1, 2, 3, 4)
```
Addition 
```lox
v = v + vec2(3,4)
v3 = v3 + vec3(1,2,3)
v4 = v4 + vec4(3,4,5,6)
```

##### Raylib Graphics Window
```lox
const width = 1500
const height = 900
var win = window(width, height)
win.init()

while (!win.should_close()) {
    win.begin()
    win.clear(10, 10, 10, 255)
    win.begin_blend_mode("BLEND_ADD")
    win.circle_fill(100, 100, 50, 255, 0, 0, 255)
    win.end_blend_mode()
    win.end()
}
win.close()
```
- Supports drawing 2d and 3d primitives, camera, images, textures, shaders, and reading keyboard state.
- **Batch rendering** - Render thousands of objects with a single draw call using `batch()` objects. For particle systems, large scenes, and real-time simulations.
`batch_instanced()` uses mesh instancing to draw 100k+ textured cubes in one call. 

##### Batch Rendering Example
```lox
// Create a batch for cubes using constants
var cube_batch = batch(win.BATCH_CUBE);

// Add 1000 cubes to the batch
for (var i = 0; i < 1000; i = i + 1) {
    var pos = vec3(random.float(-50, 50), 0, random.float(-50, 50));
    var size = vec3(1, 1, 1);
    var color = vec4(255, random.integer(0, 255), 0, 255);
    cube_batch.add(pos, size, color);
}

// Render ALL 1000 cubes in a single draw call!
win.begin_3d(camera);
cube_batch.draw();  // Replaces 1000 individual draw calls
win.end_3d();
```
 

---

#### Built-in Lox Modules

- Iterator tools, function tools, math, random, color, string utilities, PNG plotting, graphics particle system, and more.

---

#### Lists

##### Initialization
```lox
a = []
b = [1, 2, 3]
```

##### Append
```lox
a.append(4)
```

##### Indexing and Assignment
```lox
b = a[0]
a[1] = 42
```

##### Slicing
```lox
b = a[1:3]
c = a[:2]
d = a[2:]
e = a[:]
```

##### Slice Assignment
```lox
a[2:5] = [7, 8, 9]
```

##### Concatenation
```lox
c = a & b
```

##### Membership Test
```lox
if 3 in a {
    print "Found"
}
```

##### Remove
```lox
a.remove(2)
```

##### Unpacking
```lox
a, b, c = [1, 2, 3]
```

---

#### Tuples

##### Syntax
```lox
a = (1, 2, 3)
```
- Immutable, supports same operations as lists except append/assignment.

##### Unpacking
```lox
x, y, z = (1, 2, 3)
```

---

#### Dictionaries

##### Initialization
```lox
a = {}
b = {"b": "c", "d": "e"}
```

##### Get and Set
```lox
v = a[key]
v = a.get(key, default)
a[key] = b
```

##### Keys
```lox
keys = a.keys()
```

##### Remove
```lox
a.remove(key)
```

---

#### Strings

##### Formatting (wrapper for Go fmt.Sprintf)

```lox
a=math.sqrt(2.0)
b="hello"
c="world
print format("%s %s %f",a,b,c)
```

##### Concatenation
```lox
s = "hello" & "4"
```

##### Replace
```lox
s2 = s.replace("hello", "world")
```

##### Join
```lox
sep = "|"
joined = sep.join(["a", "b", "c"]) // "a|b|c"
```
or
```lox
joined = join(["a", "b", "c"], "|") // "a|b|c"
```

##### Multiplication
```lox
s = "@" * 3    // "@@@"
s = 3 * "@"    // "@@@"
```

##### Slicing
```lox
a = "abcd"
b = a[0]      // "a"
c = a[:2]     // "ab"
```

##### Substring Test
```lox
if "bc" in a {
    print "found"
}
```

- All VM strings are interned for fast lookup and runtime refers to integer string ID keys.

---

 


---

#### Classes

##### toString Magic Method
```lox
class Point {
    toString() {
        return "Point"
    }
}
p = Point()
print p // prints "Point"
```
- If present and returns a string, will be used for print class / str(class).

---

#### Exceptions

##### Syntax
```lox
try {
    // code
} except ExceptionType as e {
    // handler
} except AnotherType as e2 {
    // another handler
}
```
- Built-in Exception class, subclass custom exception classes from it.
- Can nest try/except blocks.
- Multiple handlers for different exception types.
- `raise [exception instance]` statement.
- Runtime can raise catchable exceptions e.g. RunTimeError, EOFError.

---

#### I/O

##### File Operations
```lox
import os
f = os.open("file.txt", "r")
line = os.readln(f)
os.write(f,"hello\n")
os.close(f)
```
- Native file open, close, readln, write.
- `readln` throws EOFError on end of file.
- File operations are part of the `os` module along with directory operations.

##### Directory Operations
```lox
import os

// Directory listing
files = os.listdir(".")
for file in files {
    if (os.isdir(file)) {
        print "[DIR]  " + file
    } else if (os.isfile(file)) {
        print "[FILE] " + file
    }
}

// Path manipulation
full_path = os.join("assets", "images", "sprite.png")
dir = os.dirname(full_path)      // "assets/images"
filename = os.basename(full_path) // "sprite.png"
parts = os.splitext(filename)    // ["sprite", ".png"]

// Directory operations
os.mkdir("new_directory")
current_dir = os.getcwd()
os.chdir("../parent")
```
- File system operations: `listdir`, `mkdir`, `rmdir`, `remove`
- Path testing: `exists`, `isdir`, `isfile`
- Path manipulation: `join`, `dirname`, `basename`, `splitext`
- Working directory: `getcwd`, `chdir`

#### VM inspection
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


## Performance Notes:

My implementation is slow compared to CLox. Fibonacci benchmark averages 1s, CLox is around 0.5.  Python3 averages around half that.

The VM :
- does a lot of function calls in place of C macros, not all of which get inlined
    - TODO : manual inlining of calls.
- has a large switch/case inner loop which Go compiler doesn't optimise at all well ( no computed goto ) 
- uses slow map for globals - function code runs much quicker 
- uses interface{} for objects ( values are tagged union structs for speed but contain a pointer for objects ) 
- GC is handled by the Go runtime. 

 
  
There are some optimisations such as string interning to allow integer hash keys for method lookup, singleton NIL_VALUE, inlined functions in the main run loop. 
A peephole optimiser compile step replaces two get locals and a numeric add with a single superinstruction OP_ADD_NN working directly on the stack, with a further runtime specialisation opcode rewrite to int+int or float+float if possible. This is good for for loops. A similar optimisation is done for local = local + constant.  

 
