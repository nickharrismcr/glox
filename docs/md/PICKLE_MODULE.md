# pickle Module Documentation

The `pickle` module serialises plain-data Lox values to a byte string and back, similar in spirit
to Python's `pickle` module but deliberately limited to data: `nil`, `bool`, `int`, `float`,
`string`, `list`, `tuple`, `dict`, `vec2`/`vec3`/`vec4`, arbitrarily nested. It's intended for
passing values between separate `glox` processes (over a pipe, socket, or file) or persisting a
value to disk.

**Limitations:**
- Closures/functions, class instances, classes, modules, files, and native/graphics objects cannot
  be pickled — `dumps()` raises `PickleError`.
- Cyclic structures (e.g. a list appended to itself) are detected and raise `PickleError` rather
  than hanging.
- Dict keys are always strings, matching how Lox dicts already work.

## Usage

```lox
import pickle
```

## Module functions

### `pickle.dumps(value)` → string
Serialises `value` to a string holding the raw encoded bytes.
- **value**: any plain-data Lox value, arbitrarily nested
- **Returns**: a string (may contain non-printable bytes — treat it as opaque data, not text)
- Raises `PickleError` if `value` contains an unpicklable type or a cycle

```lox
data = pickle.dumps({"name": "Rex", "tags": [1, 2, 3]})
```

### `pickle.loads(data)` → value
Deserialises a string produced by `dumps()` back into the original value.
- **data**: a string previously produced by `pickle.dumps()`
- **Returns**: the decoded value
- Raises `PickleError` on truncated or malformed input

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
```
