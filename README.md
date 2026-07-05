# GLOX

**Bob Nystroms CLox bytecode interpreter implemented in Go**


---

The aim of this project is to learn more deeply about programming in Go and the crafting of interpreters by way of implementing Bobs CLox interpreter in Go, adding Python-inspired extensions to Lox along the way.
The extensions to the language include enhanced string operations, lists, dictionaries, exception handling, module imports with bytecode caching, string and list iteration, lambda functions, Raylib bindings for graphics, and I/O.  

📖 **[Full language reference: `docs/language-reference.html`](docs/language-reference.html)** — a guide to the syntax, built-in types and functions, native objects, and library modules. Open it in a browser.  

**Authorship**

The port of Bob Nystrom's clox bytecode interpreter to Go was done **by hand**, along with the language extensions up to and including exception handling. The Raylib graphics bindings and the core VM optimisations — superinstructions, native vector types, and similar — were assisted by **GitHub Copilot**. More recent work was co-authored with **Claude Code** (Anthropic);  language features (lambdas, one-line braced blocks, loop-scope and compiler fixes), VM performance (`Value`-struct shrink, faster global lookup, per-call allocation removal), benchmarking, Raylib/physics additions and demos, and tooling, tests, and the HTML language reference.


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
 

##### Native Physics Simulation

`physics_world()` is a native 3D rigid-body simulation for spheres. Each body has a
position, velocity, radius and material; a single `step()` call applies gravity, bounces
bodies off the boundary box, and resolves body-to-body collisions natively in Go — far
faster than driving per-object physics from Lox.

```lox
import random

// world(min, max, cell_size, gravity)
world = physics_world(vec3(-10, 0, -10), vec3(10, 100, 10), 2.0, vec3(0, -0.01, 0))
mat = world.add_material(0.5, 0.3, 0.99)   // restitution, friction, damping

ids = []
foreach (i in range(50)) {
    pos = vec3(random.float(-5, 5), random.float(5, 20), random.float(-5, 5))
    ids.append(world.add(pos, vec3(0, 0, 0), 0.5, mat))
}

while (!win.should_close()) {
    world.step(1.0)                      // advance the whole simulation
    foreach (id in ids) {
        p = world.get_position(id)       // read a body's position back
        // ...draw a sphere at p...
    }
}
```
- `add(pos, vel, radius, material_id)` returns a stable id used by every other method.
- `add_impulse(id, vec3)` applies an instantaneous velocity change to one body.
- `collisions()` reports pairs `(a, b, normal, impulse)` that newly touched during the last `step()`.
- See the [language reference](docs/language-reference.html#physics-world) for the full method list.

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

#### Anonymous Functions (Lambdas)

##### Syntax
```lox
var add = func (a, b) { return a + b }
print add(2, 3)   // 5
```
- Omit the name to use a function as an expression (block body, explicit `return`).
- `fun` is accepted as an alias for `func`.

##### Closures
```lox
func make_scaler(factor) {
    return func (x) { return x * factor }   // captures `factor`
}
var double = make_scaler(2)
print double(21)   // 42
```
- Anonymous functions are ordinary closures — they capture surrounding variables, and `this` inside a method.

##### Inline callbacks
```lox
import functools
print functools.map([1, 2, 3], func (x) { return x * x })       // [ 1 , 4 , 9 ]
print functools.filter(xs, func (x) { return x > 0 })
print (func () { return 7 })()                                  // immediately-invoked
```
- A statement that begins with `func` is always parsed as a *named declaration*, so write lambdas where an expression is expected (after `=`, as an argument, in a list/dict, after `return`). Wrap in parentheses to invoke one immediately.
- Having no name, a lambda cannot recurse by name — assign it to a variable first.

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


## Testing

The project has two test suites under `tests/`:

### Assert-based suite (recommended)

```bash
# from repo root, after building
. ./setenv
bash bin/run_tests.sh          # run all tests
bash bin/run_tests.sh -v       # verbose
bash bin/run_tests.sh -k fibo  # run a single test by keyword
```

Or run directly with pytest:

```bash
. ./setenv
cd tests
python -m pytest new_tests/ -v
```

Tests live in `tests/new_tests/` — one Python module per language feature, each running a `.lox` script and making semantic assertions on the output. Large-output tests (Mandelbrot, sine table, mapfilter) use structural assertions rather than exact line-by-line comparison.

The `.lox` scripts used by the tests are in `tests/new_tests/lox/`.

### Regression suite (legacy)

```bash
. ./setenv
bash bin/run_tests.old.sh
```

The legacy runner (`tests/old/test.py`) does exact byte comparison against stored output files in `tests/old/output/`. Use `python test.py lox/foo.lox --write` to record expected output for a new script.

---

## Performance Notes:

Benchmarks run via `bin/benchmarks.sh` (loxcraft suite).  

| benchmark | glox | CPython 3 | ratio |
|---|---|---|---|
| binary_trees | 18.8s | 7.5s | 2.5× |
| equality | 52.3s | 20.1s | 2.6× |
| fib | 20.6s | 9.3s | 2.2× |
| instantiation | 39.7s | 22.5s | 1.8× |
| invocation | 14.9s | 9.2s | 1.6× |
| loop | 8.0s | 3.6s | 2.2× |
| method_call | 22.4s | 8.9s | 2.5× |
| properties | 16.2s | 7.5s | 2.2× |
| string_equality | 36.9s | 17.4s | 2.1× |
| trees | 24.5s | 6.8s | 3.6× |
| zoo | 15.1s | 10.4s | 1.5× |
| zoo_batch | 10.0s | 10.0s | 1.0× |

glox is currently 1.5–3.6× slower than CPython across the suite.

Known costs:
- **`Value` struct is 32 bytes** — clox's is ~16 bytes. Every stack push/pop copies 32 bytes.
- **No computed goto** — Go's `switch` dispatch is slower than clox's `COMPUTED_GOTO` threaded dispatch, which jumps directly to the next handler without re-entering the switch.

Optimisations in place:
- **`Value` struct reduced 64→32 bytes** in three steps:
  - Removed `Bool bool` — booleans stored as `Data` 0/1, saving 8 bytes (padding).
  - Merged `Int int` + `Float float64` into `Data uint64` — `math.Float64bits`/`math.Float64frombits` are amd64 intrinsics (single `MOVQ`), saving 8 bytes.
  - Shrunk `Type ValueType` from `int` (8 bytes) to `uint8` (1 byte) and `InternedId` from `int` (8 bytes) to `int32` (4 bytes); reordered fields to pack the small fields into the tail of the struct, saving 12 bytes.
  - Total: 5–15% improvement across benchmarks.
- **Global variable indexing** — globals are stored in a `[]Value` slice indexed by a compiler-assigned integer slot rather than a `map[int]Value` keyed by interned string ID. `OP_GET_GLOBAL` / `OP_SET_GLOBAL` go from a hash-map lookup to a direct slice index. ~10–27% improvement on global-variable-heavy benchmarks.
- String interning with integer IDs for fast method and global lookup
- Peephole pass replaces `OP_GET_LOCAL, OP_GET_LOCAL, OP_ADD` with a single `OP_ADD_NN` superinstruction, with runtime specialisation to `OP_ADD_II` / `OP_ADD_FF` on first execution. A similar optimisation handles `local = local + constant`.
- Call frames stored inline in the VM struct (not heap-allocated) to avoid per-call GC pressure.
- Frame context (`frame`, `function`, `chunk`, `constants`, `currCode`) hoisted before the dispatch loop and refreshed only at opcodes that change the active frame (`OP_CALL`, `OP_INVOKE`, `OP_SUPER_INVOKE`, `OP_RETURN`, `OP_RAISE`, toString path).
- `readShort()` and `readByte()` inlined at all call sites in the dispatch loop, eliminating indirect frame fetches on every jump and loop opcode.
- GC interval check uses a bitmask (`& 0xFFFF`) rather than modulo, avoiding a multiply-high sequence on every opcode.

 
