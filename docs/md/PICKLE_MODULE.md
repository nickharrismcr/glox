# pickle Module Documentation

The `pickle` module serialises plain-data Lox values to a byte string and back, similar in spirit
to Python's `pickle` module but deliberately limited to data: `nil`, `bool`, `int`, `float`,
`string`, `list`, `tuple`, `dict`, `vec2`/`vec3`/`vec4`, and class instances, arbitrarily nested.
It's intended for passing values between separate `glox` processes (over a pipe, socket, or file)
or persisting a value to disk.

**Class instances:** `dumps()` encodes an instance as its class *name* plus its fields — never its
class's methods or code. `loads()` reconstructs the instance by looking up a class of that name
already loaded in the decoding process (checking built-in classes first, then the calling script's
own module scope), then reattaching the decoded fields to that live class. This means:
- The receiving side needs a class of the same name already defined — typically true for built-in
  exceptions (`Exception`, `PickleError`, `ProcessError`, ...) and for any project-wide classes both
  sides import, but a class only visible in some other module's scope won't resolve. `loads()`
  raises `PickleError` in that case.
- Class identity is by name only (the same convention `except ClassName` already uses to look up a
  handler) — a same-named class in a different scope is indistinguishable to `loads()`.
- Because the instance is reattached to a real, live class, its methods and inheritance chain work
  normally after unpickling — including `except`-style class checks and overridden `toString()`.
- Classes themselves (not instances of them) still cannot be pickled — `dumps()` raises
  `PickleError`.

**Other limitations:**
- Closures/functions, modules, files, and native/graphics objects cannot be pickled — `dumps()`
  raises `PickleError`.
- Cyclic structures (e.g. a list appended to itself, or an instance field pointing back at its own
  instance) are detected and raise `PickleError` rather than hanging.
- Dict keys are always strings, matching how Lox dicts already work.

## Usage

```lox
import pickle
```

## Module functions

### `pickle.dumps(value)` → string
Serialises `value` to a string holding the raw encoded bytes.
- **value**: any plain-data Lox value or class instance, arbitrarily nested
- **Returns**: a string (may contain non-printable bytes — treat it as opaque data, not text)
- Raises `PickleError` if `value` contains an unpicklable type or a cycle

```lox
data = pickle.dumps({"name": "Rex", "tags": [1, 2, 3]})
```

### `pickle.loads(data)` → value
Deserialises a string produced by `dumps()` back into the original value.
- **data**: a string previously produced by `pickle.dumps()`
- **Returns**: the decoded value
- Raises `PickleError` on truncated or malformed input, or on an encoded instance whose class isn't
  loaded in this process (see "Class instances" above)

```lox
value = pickle.loads(data)
print value["name"]   // "Rex"
print value["tags"]   // [ 1 , 2 , 3 ]
```

## Example

```lox
import pickle

d = {"name": "Rex", "pos": vec2(1, 2), "tags": (1, 2, 3)}
data = pickle.dumps(d)
d2 = pickle.loads(data)
print d2["name"]     // "Rex"
print d2["pos"].x    // 1
print d2["tags"]     // ( 1 , 2 , 3 )

try {
    pickle.dumps(pickle.dumps)   // a function -- not picklable
} except PickleError as e {
    print e.msg
}

class Point {
    init(x, y) {
        this.x = x
        this.y = y
    }
}
p2 = pickle.loads(pickle.dumps(Point(3, 4)))
print p2.x   // 3
print p2.y   // 4
```
