# GLox performance roadmap

Why glox runs 1.5–3.6× slower than CPython, and a prioritised, profile-first
plan to narrow the gap. Grounded in the current code — the run loop is in
[`src/vm/vm.go`](../src/vm/vm.go), the value representation in
[`src/core/value.go`](../src/core/value.go), and the object types in
[`src/core/`](../src/core/).

The README's Performance Notes name two costs — the 32-byte `Value` struct and
the lack of computed-goto dispatch. Those are real, but they only account for a
**flat ~2.2× floor**, visible in the allocation-free numeric benchmarks (`fib`,
`loop`). Everything *above* that floor — `trees` at 3.6×, `method_call` at 2.5×,
`equality` at 2.6× — comes from three families of cost the README omits: the
**object model**, **GC/allocation churn**, and **Go's per-operation safety
overheads**.

---

## The factors

### A. Object model — interface dispatch + map-backed fields

*Drives `trees`, `method_call`, `properties`, `invocation`, `instantiation`.*

- **Interface virtual dispatch for every heap object.** The `Object` interface
  ([`src/core/object.go:43-48`](../src/core/object.go)) is implemented by every
  heap type, and `GetType()` is called constantly to discriminate types —
  through the itab, where clox switches on a single tag byte. Even the numeric
  `<` / `>` opcodes call `IsStringObject()` (an interface call) before reaching
  the numeric fast path ([`src/vm/vm.go:421,440`](../src/vm/vm.go)).
- **Instance fields are a Go `map[int]Value`, allocated per instance.** Each
  field read is `Fields[id]` (hash + probe), each write a map insert
  ([`src/core/obj_instance.go:14-16`](../src/core/obj_instance.go),
  [`src/vm/vm.go:1078-1088`](../src/vm/vm.go)). This is the same map-vs-slot
  problem the globals optimisation already solved (globals moved from a map to a
  slice indexed by a compiler-assigned slot) — but it has not yet been done for
  instances. This is the likely top contributor to `trees` and `properties`.
- **Method lookup is also a map**, and binding allocates. `class.Methods[id]`
  ([`src/vm/vm.go:1834`](../src/vm/vm.go)); `bindMethod` heap-allocates a
  `BoundMethodObject` on every method-value access
  ([`src/vm/vm.go:1925-1930`](../src/vm/vm.go)). There is no inline caching of
  property offsets or method targets, so every access re-does the full lookup.

### B. Allocation & GC pressure

*Drives `binary_trees`, `trees`, `instantiation`; taxes any collection loop.*

- **Every list/dict literal rebuilds its method table.** `MakeListObject` calls
  `RegisterAllListMethods`, allocating a `map[int]*BuiltInObject` plus four fresh
  closures for *every* list created
  ([`src/core/obj_list.go:15-21,45-93`](../src/core/obj_list.go)); dicts do the
  same. clox's collection methods are static functions with zero per-object cost.
- **An explicit `runtime.GC()` fires on every top-level return**
  ([`src/vm/vm.go:962-965`](../src/vm/vm.go)) — a stop-the-world collection each
  time the outermost frame unwinds.
- **The value stack is GC-scanned.** A `Value` contains an `Obj` interface
  pointer, so the whole `[16384]Value` stack and every `[]Value` are
  pointer-bearing and scanned by the collector; stores of pointer-containing
  Values incur write barriers while marking. clox runs no GC on the hot path.
- **String materialisation allocates and re-hashes.** `OP_CONCAT` / `OP_STR`
  build a new Go string then re-intern it (a `map[string]int` hash) every time
  ([`src/vm/vm.go:819,1434`](../src/vm/vm.go)); dict keys are re-interned on
  every get/set. Object equality can fall back to `a.String() == b.String()`,
  allocating two Go strings just to compare
  ([`src/core/value.go:148`](../src/core/value.go)).

### C. Go per-operation safety & indirection

*A flat tax on every benchmark, amplified by stack-based dispatch.*

- **Bounds checks on every slice index** — `stack[...]`, `currCode[...]`,
  `constants[...]`, `globals[...]` — where clox uses raw pointer arithmetic.
- **`frame.Ip` is a heap field, not a register.** Every operand read is
  `currCode[frame.Ip]; frame.Ip++` — a pointer-chased read-modify-write
  ([`src/vm/vm.go:504-505`](../src/vm/vm.go)) — where clox keeps `ip` in a local
  the C compiler pins to a register. The stack pointer is likewise an index.
- **`refreshFrame()` reloads six locals after every call/return**
  ([`src/vm/vm.go:372-380`](../src/vm/vm.go)), chasing pointers through
  `Closure.Function.Chunk` / `.Environment`. This taxes call-heavy `fib`,
  `invocation`, `method_call`, and `trees`.
- **A `DebugHook != nil` check on every instruction**
  ([`src/vm/vm.go:393`](../src/vm/vm.go)), multiplied by billions of dispatches.

### Why the two named factors under-explain the gap

`fib` and `loop` are allocation-free integer work and sit near ~2.2× — that is
roughly the pure dispatch + 32-byte-copy + bounds-check tax (family C, plus the
two README factors). Everything above 2.2× is the object model and GC cost
(families A and B). `trees` at 3.6× is all of them stacked: map-backed fields
(A) + bound-method allocation (A) + per-node instance/map garbage (B) + the
forced GC (B).

---

## Prioritised, profile-first roadmap

Don't optimise blind — the factors above are hypotheses ranked by evidence.
Confirm attribution with a profiler, then fix in impact order.

### Step 0 — Measure

Add `runtime/pprof` CPU and allocation profiling behind a flag, and run the
worst offenders (`trees`, `method_call`, `binary_trees`, `string_equality`).
Confirm whether the cost is `mapaccess`/`mapassign` (family A), `mallocgc` /
`runtime.GC` (family B), or `assertI2T`/itab (family A). The profile sets the
order of everything below.

### Step 1 — Cheap, high-confidence wins

- **Gate the forced `runtime.GC()`** at
  [`src/vm/vm.go:964`](../src/vm/vm.go) to debug-only; measure `binary_trees` /
  `trees` before and after.
- **Cache collection and string method tables** package-level (shared, keyed by
  interned id) instead of rebuilding them per object in
  `RegisterAllListMethods` / `StringObject.GetMethod`.
- **Hoist the per-instruction `DebugHook` check** out of the hot loop (build tag
  or a debug/non-debug loop variant).

### Step 2 — The structural win: slot-based instance fields

Apply the globals treatment to instances: give each `ClassObject` a compiled
field-name → slot map, store fields in a `[]Value` on the instance, and emit
`OP_GET_FIELD_SLOT` / `OP_SET_FIELD_SLOT` when the field is statically known
(`this.x` inside methods, monomorphic sites), falling back to the map for
dynamic/unknown fields. This targets the likely top contributor to
`trees`/`properties`/`method_call`. Optionally add a one-entry inline cache
(class-id → slot/method) on `OP_GET_PROPERTY` / `OP_INVOKE`.

### Step 3 — Dispatch & object-model micro-opts

- Use the `Value.Type` tag instead of `GetType()` interface calls where the tag
  already distinguishes the case; move the `IsStringObject()` check in `<` / `>`
  behind the numeric fast path.
- Reduce string re-interning on concat — intern only when a string is used as a
  key, not on every concat result.

### Step 4 — Research-level (only if profiles justify it)

- Threaded dispatch via a `[]func()` jump table (Go has no computed goto; this
  trades the switch for indirect calls — measure, it can lose).
- Further `Value` shrink toward NaN-boxing, or a register-based bytecode. Large
  efforts; defer until Steps 1–3 are exhausted.

---

## Verification

For each change:

- `go build -o bin/glox main.go`, then `bash bin/clear_lxc.sh` (serialisation-
  adjacent changes invalidate cached `.lxc` files).
- Re-run the profiler on the same benchmark and confirm the targeted symbol
  shrank.
- `bin/benchmarks.sh 3` for stable ratios; update the README Performance table.
- Full correctness gate: `python -m pytest tests/new_tests/ -x -q` (with
  `LOX_PATH` and `bin` on `PATH`). Slot-based fields and shared method tables
  are behaviour-sensitive — the OO tests must stay green.
