# GLOX

**Bob Nystroms CLox bytecode interpreter implemented in Go**


---

The aim of this project is to learn more deeply about programming in Go and the crafting of interpreters by way of implementing Bobs CLox interpreter in Go, adding Python-inspired extensions to Lox along the way.
The extensions to the language include enhanced string operations, lists, dictionaries, exception handling, module imports with bytecode caching, string and list iteration, lambda functions, Raylib bindings for graphics, and I/O.  

📖 **[Full language reference: `docs/language-reference.html`](docs/language-reference.html)** — a guide to the syntax, built-in types and functions, native objects, and library modules. Open it in a browser.  

**Authorship**

The port of Bob Nystrom's clox bytecode interpreter to Go was done **by hand**, along with the language extensions up to and including exception handling. The Raylib graphics bindings and the core VM optimisations — superinstructions, native vector types, and similar — were assisted by **GitHub Copilot**. More recent work was co-authored with **Claude Code** (Anthropic);  language features (lambdas, one-line braced blocks, loop-scope and compiler fixes), VM performance (`Value`-struct shrink, faster global lookup, per-call allocation removal), benchmarking, Raylib/physics additions and demos, and tooling, tests, and the HTML language reference.


### Additions to vanilla Lox

Feature summary — see the **[language reference](docs/language-reference.html)** for full syntax, methods, and examples.

**Language**
- **Optional semicolons** — a newline or a closing `}` terminates a statement; braced blocks may be written on one line.
- **Implicit variable declaration** (`a = 1`) and **`const`** immutables.
- **Integer type** with `%` modulus, distinct from float.
- **Destructuring / unpacking assignment** — `a, b, c = [1, 2, 3]`.
- **Compound assignment** — `+=`, `-=`, `++`.
- **`break` / `continue`**, and **`foreach`** over lists, strings, and iterables (`__iter__`/`__next__`).
- **`range(start, end, step)`** — native integer iterator, faster than an equivalent `for`.
- **Anonymous functions (lambdas)** — `func (x) { ... }` as expressions; full closures.
- **Exceptions** — `try` / `except` / `finally`, `raise`, custom `Exception` subclasses, catchable runtime errors.
- **Module imports with bytecode caching** — `import m`, `import m as alias`, `from m import ...`; compiled modules cached as `__loxcache__/<module>.lxc`.

**Types & operators**
- **Lists** — slicing, slice assignment, `&` concatenation, `in` membership, `append`/`remove`.
- **Tuples** — immutable sequences.
- **Dictionaries** — `get(k, default)`, `keys()`, `remove()`.
- **Strings** — `format()` (Go `Sprintf`), `&` concat, `*` repeat, slicing, `in`, `replace`, `join`; all interned.
- **Native vectors** `vec2` / `vec3` / `vec4` — inlined (no heap allocation); `++` concatenation.
- **`float_array`** — fast native 2D float grid.

**Classes**
- **`toString()`** magic method, **static methods**, and the **iterator protocol** (`__iter__` / `__next__`).

**Native & graphics**
- **Raylib `window`** — 2D/3D primitives, camera, textures, shaders, images, keyboard input.
- **Batch rendering** — `batch()` draws thousands of primitives per call; `batch_instanced()` draws 100k+ instanced textured cubes.
- **`physics_world`** — native 3D rigid-body sphere simulation (gravity, boundary bounce, collisions in Go).
- **File & directory I/O** via `os`; PNG output; RGB encode/decode.
- **Built-in modules** — `math`, `random`, `colour`, `string`, `itertools`, `functools`, `particle_sys`, `sys`, `os`, `inspect`.

---

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

This is a toy project written in go, its expected that it will perform poorly compared to Cpython or Clox. However it has been instructive and fun to implement optimisations to squeeze more performance out of the VM, or lift often used lox functions into the language library in go to get a native performance boost.

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

**Why a C VM (clox) is faster.** The gap is structural, not a handful of missing tricks. clox is a tagged-union value in ~16 bytes with `ip`/stack pointers pinned in registers, raw pointer arithmetic (no bounds checks), object type dispatched by a single tag byte, instance fields and methods in a purpose-built open-addressing hash table, and no garbage collector on the hot path. glox pays Go's costs for the same work: a 32-byte `Value`, an `Object` **interface** (virtual dispatch) for every heap type, **Go `map`-backed** instance fields and method tables, bounds-checked slice indexing, a pointer-bearing value stack that the **garbage collector must scan** (with write barriers), and per-object allocation for collection method tables and bound methods. The allocation-free numeric benchmarks (`fib`, `loop`) sit near the ~2.2× floor that dispatch and copy overhead impose; the object-heavy ones (`trees`, `method_call`) run wider because of the map lookups and GC pressure on top.

A prioritised plan to close the gap — profiling steps, cheap wins, and the larger structural changes (slot-based instance fields, cached method tables) — is in **[docs/performance-roadmap.md](docs/performance-roadmap.md)**.

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

 
