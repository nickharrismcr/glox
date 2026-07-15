# GLOX

**Bob Nystroms CLox bytecode interpreter implemented in Go**


---

The aim of this project is to learn more deeply about programming in Go and the crafting of interpreters by way of implementing Bobs CLox interpreter in Go, adding Python-inspired extensions to Lox along the way.
The extensions to the language include enhanced string operations, lists, dictionaries, exception handling, module imports with bytecode caching, string and list iteration, lambda functions, Raylib bindings for graphics, and I/O.  

📖 **[Full language reference: `docs/language-reference.html`](docs/language-reference.html)** — a guide to the syntax, built-in types and functions, native objects, and library modules. Open it in a browser.  

**Authorship**

The port of Bob Nystrom's clox bytecode interpreter to Go was done **by hand**, along with the language extensions up to and including exception handling. The Raylib graphics bindings and the core VM optimisations — superinstructions, native vector types, and similar — were assisted by **GitHub Copilot**. More recent work was co-authored with **Claude Code** (Anthropic);  language features (lambdas, one-line braced blocks, the full compound-assignment set, the ternary conditional expression, default & variadic parameters, loop-scope and compiler fixes), VM performance (`Value`-struct shrink, faster global lookup, per-call allocation removal), benchmarking, Raylib/physics additions and demos, and tooling, tests, and the HTML language reference.


### Additions to vanilla Lox

Feature summary — see the **[language reference](docs/language-reference.html)** for full syntax, methods, and examples.

**Language**
- **Optional semicolons** — a newline or a closing `}` terminates a statement; braced blocks may be written on one line.
- **Implicit variable declaration** (`a = 1`) and **`const`** immutables.
- **Integer type** with `%` modulus, distinct from float.
- **Destructuring / unpacking assignment** — `a, b, c = [1, 2, 3]`.
- **Compound assignment** — `+=`, `-=`, `*=`, `/=`, `%=`, `++`.
- **Ternary / conditional expression** — `cond ? a : b` (C-style, right-associative).
- **String interpolation** — `"total: ${count} (${pct}%)"` in either quote style; `$$` escapes a literal `$`.
- **`break` / `continue`**, and **`foreach`** over lists, strings, and iterables (`__iter__`/`__next__`).
- **`range(start, end, step)`** — native integer iterator, faster than an equivalent `for`.
- **Anonymous functions (lambdas)** — `func (x) { ... }` as expressions; full closures.
- **Default & variadic parameters** — `func f(a, b=expr)` (defaults evaluated at call time) and a trailing `*rest` that collects surplus positional arguments into a list.
- **Exceptions** — `try` / `except` / `finally`, `raise`, custom `Exception` subclasses, catchable runtime errors.
- **Module imports with bytecode caching** — `import m`, `import m as alias`, `from m import ...`; compiled modules cached as `__loxcache__/<module>.lxc`.

**Types & operators**
- **Lists** — slicing, slice assignment, `&` concatenation, `in` membership, `append`/`remove`.
- **Tuples** — immutable sequences.
- **Dictionaries** — `get(k, default)`, `keys()`, `remove()`.
- **Strings** — `${expr}` interpolation, `format()` (Go `Sprintf`), `&` concat, `*` repeat, slicing, `in`, `replace`, `join`; all interned.
- **Native vectors** `vec2` / `vec3` / `vec4` — heap-allocated objects tagged directly in the `Value` (no interface dispatch to discriminate); `++` addition, `.add()` in-place addition.
- **`float_array`** — fast native 2D float grid.

**Classes**
- **`toString()`** magic method, **static methods**, and the **iterator protocol** (`__iter__` / `__next__`).

**Native & graphics**
- **Raylib `window`** — 2D/3D primitives, camera, textures, shaders, images, keyboard input.
- **Batch rendering** — `batch()` draws thousands of primitives per call; `batch_instanced()` draws 100k+ instanced textured cubes.
- **`physics_world`** — native 3D rigid-body sphere simulation (gravity, boundary bounce, collisions in Go).
- **File & directory I/O** via `os`; PNG output; RGB encode/decode.
- **Built-in modules** — `math`, `random`, `colour`, `string`, `itertools`, `functools`, `particle_sys`, `sys`, `os`, `inspect`, `gfx` (graphics constructors: `window`, `batch`, `texture`, `shader`, `camera`, …), `physics` (`physics_world`). Import with `from gfx import *` or `import gfx`.

---

## Build

```bash
# Fast build (default) -- what bin/glox is built as; also `bash bin/build.sh`
go build -o bin/glox main.go
```

`bin/glox` never compiles in the per-instruction debug hook that `--debug`, `--info`, and `--instrument` need — its mere presence in the hot dispatch loop costs ~25% on dispatch-bound code (see `docs/performance-roadmap.md`). Those flags still run against `bin/glox`, but print a warning and produce empty trace output / zero instruction counts rather than silently doing nothing.

```bash
# Debug build -- hook compiled in, for real --debug/--info/--instrument output
bash bin/build_debug.sh
```

This produces `bin/debug_glox` (and `bin/debug_glox.exe`) and leaves the source tree unmodified afterward — it temporarily uncomments the hook line to build, then restores it.

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

Benchmarks run via `bin/benchmarks.sh` (loxcraft suite, plus `collections`, a glox-specific addition exercising list/dict/string built-in methods in a hot loop). All numbers are from `bin/glox`, the default fast build (see **Build** above), measured back-to-back in one sitting (3-run averages) — this is a thermally-constrained laptop with a measured ±10–17% run-to-run noise floor (see `docs/performance-roadmap.md`), so don't read small deltas between benchmarks as significant.

| benchmark | glox | CPython 3 | ratio |
|---|---|---|---|
| binary_trees | 18.8s | 7.5s | 2.5× |
| collections | 10.5s | 2.9s | 3.6× |
| equality | 52.3s | 19.9s | 2.6× |
| fib | 23.3s | 9.1s | 2.6× |
| instantiation | 41.3s | 21.7s | 1.9× |
| invocation | 16.5s | 9.5s | 1.7× |
| loop | 6.3s | 3.7s | 1.7× |
| method_call | 21.2s | 8.6s | 2.5× |
| properties | 18.1s | 7.6s | 2.4× |
| string_equality | 40.6s | 17.2s | 2.4× |
| trees | 23.9s | 6.7s | 3.6× |
| zoo | 16.7s | 10.0s | 1.7× |
| zoo_batch | 10.0s | 10.0s | 1.0× |

glox is currently 1.7–3.6× slower than CPython across the suite.

**Why a C VM (clox) is faster.** The gap is structural, not a handful of missing tricks. clox is a tagged-union value in ~16 bytes with `ip`/stack pointers pinned in registers, raw pointer arithmetic (no bounds checks), object type dispatched by a single tag byte, instance fields and methods in a purpose-built open-addressing hash table, and no garbage collector on the hot path. glox pays Go's costs for the same work: a 32-byte `Value`, an `Object` **interface** (virtual dispatch) for every heap type, **Go `map`-backed** instance fields and method tables, bounds-checked slice indexing, a pointer-bearing value stack that the **garbage collector must scan** (with write barriers), and per-call allocation for bound methods. `loop` is the closest of the numeric benchmarks to CPython (1.7×) after removing the per-instruction debug hook from the default build's dispatch loop — its mere presence cost ~25% there even as a near-always-false branch. `fib` stays further out because call/return overhead (frame setup, `refreshFrame`) dominates it more than dispatch does. The object-heavy benchmarks (`trees`, `method_call`) run widest because of `map`-backed instance fields and method lookup on top of that, and GC pressure from the per-object allocation they cause — see `docs/performance-roadmap.md` for the profiled breakdown and the planned slot-based-fields fix.

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

 
