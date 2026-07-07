# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

**glox** is a bytecode interpreter for Lox (from Bob Nystrom's *Crafting Interpreters*), implemented in Go. It extends vanilla Lox with: lists, dicts, tuples, slices, exceptions, module imports with bytecode caching, `foreach`/`range`, integer arithmetic, `break`/`continue`, `const`, string interning, native vector types (`vec2`/`vec3`/`vec4`), a `float_array` type, and Raylib bindings for 2D/3D graphics.

## Language reference

`docs/language-reference.html` is the comprehensive, hyperlinked language reference (syntax, built-in types and functions, native objects, and library modules). **Consult it whenever writing `.lox` code** to get syntax and available functions/methods right rather than guessing.

**Keep it in sync:** whenever you add or change a language feature — new syntax, keyword, built-in function, native object/method, or library module — update `docs/language-reference.html` in the same change. The topic Markdown docs it builds on live in `docs/md/`; update those too if the change touches their area.

## Build & run

```powershell
# Build
go build -o bin/glox main.go

# Run a script
.\bin\glox script.lox

# REPL
.\bin\glox --repl

# Set LOX_PATH (needed for module resolution and tests)
. .\setenv.ps1
```

Key flags:
- `-d` / `--debug` — print bytecode and trace execution
- `-c` / `--compile-only` — compile without running
- `-f` / `--force-compile` — recompile cached modules
- `-i` / `--instrument` — print timing and instruction count
- `-n` / `--no-peephole` — skip the peephole optimiser

## Bytecode cache (.lxc files)

Modules are cached as compiled bytecode in `__loxcache__/*.lxc` directories. Stale `.lxc` files (compiled with an older binary) cause hangs or out-of-memory panics when loaded by a newer binary. **Always run `bin/clear_lxc.sh` after any change that affects `.lxc` serialisation** — this includes changes to `Value`, `Chunk`, `bc_cache.go`, or any other type that is written/read by `src/vm/bc_cache.go`.

```bash
bash bin/clear_lxc.sh
```

## Tests

Tests are run via Python (requires `bin/glox` on PATH):

```bash
# Run all tests (bash, from repo root)
. ./setenv && cd tests && python test.py

# Run a single test with diff on failure
cd tests && python test.py lox/fibo.lox --verbose --diff

# Write/update expected output for a test
cd tests && python test.py lox/fibo.lox --write
```

Each `.lox` file in `tests/lox/` has a corresponding expected output in `tests/output/<name>.lox.testoutput`. The test runner runs each script twice: once forcing module recompilation and once loading from cached `.lxc` files.

Files with a `_ns` suffix are the same tests run without Raylib (no-graphics) to allow CI without display.

## Architecture

### Pipeline

```
Source (.lox)
  → Scanner (src/compiler/scanner.go) → Tokens
  → Compiler/Parser (src/compiler/compile.go) → Chunk (bytecode + constants)
  → VM (src/vm/vm.go) → executes opcodes
```

The compiler does a single-pass Pratt parser; there is no AST. Closures, upvalues, and class methods are all resolved at compile time.

### Package layout

| Package | Role |
|---|---|
| `src/core` | All shared types: `Value`, `Chunk`, opcodes, object interfaces, all `Object` implementations (list, dict, string, closure, class, instance, module, iterators, vec2/3/4, etc.), string interning |
| `src/compiler` | Scanner + single-pass Pratt compiler. Emits bytecode into a `Chunk`. |
| `src/vm` | The main run loop (`vm.go`), built-in function dispatch (`builtin.go`), and bytecode cache (`bc_cache.go`) for module `.lxc` files |
| `src/builtin` | Native object implementations: Raylib window, texture, shader, batch, camera, render texture, `float_array`, image. Also core/math/color/os functions callable from Lox |
| `src/debug` | Disassembler, execution tracer, VM introspection for the `inspect` module |
| `src/util` | I/O helpers, colour utilities |

### Value representation

`Value` (in `src/core/value.go`) is a tagged union struct with a `Data uint64` field (holding an int cast, float64 bits, or bool 0/1), an `Obj` (`Object` interface) field, plus `InternedId`, `Type`, and `Immut`. All strings are **interned** — the VM works with integer IDs for string keys (method lookup, globals) to avoid repeated hashing. `Value.InternedId` caches the interned ID directly on the value so the VM doesn't need to cast `Obj` to `StringObject` on every global/method lookup.

The struct is currently **32 bytes** (vs clox's ~16 bytes) — a known performance cost, reduced from an earlier 64 bytes. See `docs/performance-roadmap.md` for the remaining performance gap and optimisation plan.

### Object types

All heap objects implement the `Object` interface in `src/core/object.go`. Concrete types live in `src/core/obj_*.go` and `src/builtin/obj_builtin_*.go`. Adding a new native object means:
1. Define the struct in `src/core/` or `src/builtin/`
2. Implement the `Object` interface (GetType, etc.)
3. Add a constructor function registered via `vm.defineBuiltins()` in `src/vm/builtin.go`
4. Add method dispatch in the corresponding `*_methods.go` file

### Module system

`import modulename` compiles the module source and caches the result as `__loxcache__/<module>.lxc` (binary-serialised `Chunk`). On subsequent imports the `.lxc` is loaded unless the source is newer. `--force-compile` bypasses the cache. Built-in modules (math, random, color, etc.) are registered as `BuiltInModules` on the VM rather than loaded from disk.

### Peephole optimiser

After compilation, a peephole pass (`src/vm/vm.go`) replaces common patterns (two `OP_GET_LOCAL` + `OP_ADD_NUMERIC`) with superinstructions (`OP_ADD_NN`, `OP_ADD_II`, `OP_ADD_FF`, `OP_INCR_CONST_*`). This is especially effective for numeric for-loops.

### VM dispatch loop internals

The `run()` function in `src/vm/vm.go` is the hot path. Key invariants to preserve:

- Five locals are hoisted **before** the `for` loop: `frame *core.CallFrame`, `function *core.FunctionObject`, `chunk *core.Chunk`, `constants []core.Value`, `vm.currCode []uint8`. These are kept in sync by a `refreshFrame()` closure.
- `refreshFrame()` **must** be called after any opcode that changes `vm.frameCount`: `OP_CALL`, `OP_INVOKE`, `OP_SUPER_INVOKE`, `OP_RETURN` (loop-continue path only), `OP_RAISE`, `OP_STR` (toString path), and after `RaiseExceptionByName` succeeds at the `End:` label. Also after `vm.run(RUN_CURRENT_FUNCTION)` returns in `OP_FOREACH`/`OP_NEXT`.
- `readShort()` and `readByte()` helpers have been **deleted** — their logic is inlined at call sites. Do not re-introduce calls to them; inline directly using `vm.currCode[frame.Ip]` and `frame.Ip++`.
- The specialised peephole opcodes (`OP_ADD_II`, `OP_ADD_FF`, `OP_INCR_CONST_I`, `OP_INCR_CONST_F`) use the hoisted `frame` and `constants` locals directly — keep them consistent if refactoring.

### Tests

The active test suite is `tests/new_tests/` (pytest). The old `tests/test.py` is legacy. Run tests with:

```powershell
$env:LOX_PATH = "d:\go\glox"
$env:PATH = "d:\go\glox\bin;" + $env:PATH
python -m pytest tests/new_tests/ -x -q
```

### Benchmarks

`bin/benchmarks.sh [N]` runs the full loxcraft suite (11 benchmarks) against CPython and prints a comparison table. Results are recorded in the README Performance Notes section. Run with `N=1` for a quick pass, `N=3` or more for stable numbers.

## Raylib / graphics

The Raylib binding (`github.com/gen2brain/raylib-go/raylib`) requires `raylib.dll` on Windows to be on `PATH` or in the working directory. Graphics scripts should only be run in environments with a display. Tests suffixed `_ns` ("no-screen") exist for headless CI.
