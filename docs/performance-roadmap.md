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

## Deep dive: the `Value.Obj` interface representation

`Value.Obj` ([`src/core/value.go:33`](../src/core/value.go)) is a Go interface —
a **two-word fat pointer** `(itab, data)` = 16 bytes. It is the single largest
field, driving the struct to 32 bytes, and it imposes **four distinct taxes**.
Separating them matters because different fixes target different taxes:

1. **Virtual dispatch on `GetType()`** — 21 call sites in the run loop
   ([`src/vm/vm.go`](../src/vm/vm.go)), including inside `OP_ADD`
   ([`820`](../src/vm/vm.go)), `OP_GET_PROPERTY` ([`1029`](../src/vm/vm.go),
   [`1146`](../src/vm/vm.go)), `OP_INVOKE` ([`1820`](../src/vm/vm.go)), and
   equality. Each is an itab load + indirect call that Go rarely devirtualises,
   where clox switches on one tag byte.
2. **Type assertions** — every `v.Obj.(*ListObject)` etc.
   ([`src/core/value.go:315-482`](../src/core/value.go)) does an itab comparison
   before yielding the concrete pointer.
3. **Width** — 16 of the 32 bytes, copied on every stack push/pop and `[]Value`
   shuffle.
4. **GC scanning + write barriers** — because `Obj` is a pointer, the entire
   `[16384]Value` stack and every `[]Value` are pointer-bearing: the collector
   scans all of it, and every store of an object-bearing `Value` hits a write
   barrier during marking. This is a large slice of family B.

**Existing precedent:** `vec2`/`vec3`/`vec4` already dodge tax #1 — they carry
their own top-level `ValueType` tags (`VAL_VEC2/3/4`) *and* an `Obj`, so the VM
discriminates them from the tag byte without a `GetType()` call. The pattern
below just generalises that.

### Options, cheapest to most radical

**Option 1 — Discriminate object subtype from a tag byte (keep the interface).**
✅ **Implemented.** Added an `ObjType ObjectType` byte to `Value` (in the free
padding at [`src/core/value.go`](../src/core/value.go), so **zero size growth** —
still 32 bytes; `ObjectType` narrowed to `uint8`), set in the five `Make*Value`
constructors, and replaced all 21 `v.Obj.GetType()` run-loop reads with
`v.ObjType`. *Kills #1* (each virtual call becomes a byte compare). *Leaves
#2–#4.* Correctness-equivalent (assertions still validate); all 155 `new_tests`
pass. An isolated Go microbenchmark measured the discrimination step itself at
**~1.89 ns → ~0.84 ns per call (~2.25×)**; end-to-end this is a small
single-digit-percent win, below the wall-clock noise floor of a thermally-
constrained laptop (control benchmarks `fib`/`loop` drifted ±10–17% per run).
Highest confidence-to-effort item; generalises the vec2/3/4 precedent.

**Option 2 — Replace the `Object` interface with `unsafe.Pointer` + the tag.**
Store `Ptr unsafe.Pointer` instead of `Obj Object`; the Option 1 tag selects the
concrete type and `(*ListObject)(v.Ptr)` recovers it with no itab check. *Kills
#1, #2, and half of #3* — `Value` drops 32 → 24 bytes (one word not two).
*Leaves #4* (still a real pointer, still scanned). Effort: **medium-high** and
invasive — every `.Obj` access across `core`/`vm`/`builtin`/`debug` changes, Go
type-checking is lost on those conversions (a wrong tag becomes memory
corruption, not a panic), polymorphic interfaces (`Iterator`, `HasMethods`,
`Iterable`) still need a narrow dispatch path, and serialisation
([`src/core/value.go:415`](../src/core/value.go)) touches `.Obj` so `.lxc` code
changes → `clear_lxc.sh`.

**Option 3 — Handle/index instead of pointer (the GC win).** Store a `uint32`
index into per-type object pools (`[]*ListObject`, `[]*InstanceObject`, …)
selected by the tag byte. If `Value` then contains *no pointers*, the value
stack and every `[]Value` become **pointer-free**: the GC skips them and write
barriers vanish on the hottest stores. *Kills all four*, including #4 — the one
tax that is family B, the cost *above* the 2.2× floor. `Value` could shrink to
~16 bytes (near clox parity). Effort: **high** — needs lifetime management
(pools keep objects alive; slot reuse or you leak), essentially a semi-manual
heap. Risk: **high** — a stale handle is a use-after-free. Interned strings
(`InternedId`) are a partial precedent, but reclaimable mutable instances are a
different beast.

**Option 4 — NaN-boxing into a single `uint64`.** Largely a **dead end in Go**:
Go's precise GC must identify every pointer by static type, so a pointer
smuggled into a NaN payload is invisible to the collector and gets freed
underneath you. clox can NaN-box because it manages its own GC roots; we cannot,
short of adopting Option 3's pools anyway. Skip unless paired with handles.

**Option 5 — Inline-cache the dispatch *result* (orthogonal).** Don't shrink the
representation; cache what the interface call resolves to — a one-entry
`(classID → slot/method)` cache on `OP_GET_PROPERTY`/`OP_INVOKE` (see Step 2).
Removes the *repeated* cost at monomorphic sites and attacks map-backed fields
(family A) at the same time. Complementary to Option 1.

### How this maps to the taxes

| Option | #1 dispatch | #2 assert | #3 width | #4 GC scan | Effort | Risk |
|---|---|---|---|---|---|---|
| 1 tag byte | ✅ | — | — | — | low | low |
| 2 unsafe.Pointer | ✅ | ✅ | ½ (→24B) | — | med-high | med |
| 3 handles | ✅ | ✅ | ✅ (→~16B) | ✅ | high | high |
| 4 NaN-box | — dead end in Go — | | | | | |
| 5 inline cache | (caches result) | | | | med | med |

**Sequencing:** Option 1 first (free, low-risk, generalises vec2/3/4), then
profile. Options 1–2 attack dispatch/width (families A + C); **Option 3 is the
only one that attacks GC scanning of the stack (family B)** — the roadmap's
prime suspect for the above-2.2× cost. So if allocation/GC profiling confirms
family B dominates *after* 1–2, the interface's *pointer-ness* (#4) matters more
than its *dispatch* (#1), and handles become the priority despite the hazard.
Skip Option 4 in Go.

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

- ✅ **Object-subtype tag byte (Option 1 above) — done.** `ObjType` byte added to
  `Value` in the free padding; all 21 hot-loop `v.Obj.GetType()` calls replaced
  with a byte compare — kills tax #1 at zero size cost. Discrimination measured
  ~2.25× faster in isolation; see the deep-dive above.
- Use the `Value.Type` tag instead of `GetType()` interface calls where the tag
  already distinguishes the case; move the `IsStringObject()` check in `<` / `>`
  behind the numeric fast path.
- Reduce string re-interning on concat — intern only when a string is used as a
  key, not on every concat result.

### Step 4 — Research-level (only if profiles justify it)

- **`unsafe.Pointer` `Value` (Option 2 above).** Drop the interface for a raw
  pointer + tag → 24-byte `Value`, no itab checks. Invasive; trades Go type
  safety for width. Do only if profiling still shows `assertI2T`/itab cost after
  Step 3.
- **Handle/index representation (Option 3 above).** Replace object pointers with
  `uint32` pool indices to make `Value` pointer-free — the only option that
  removes GC scanning of the value stack (tax #4, family B). Biggest win, biggest
  hazard (semi-manual heap, use-after-free risk); pursue only if allocation/GC
  profiling confirms family B still dominates. **NaN-boxing is a dead end in Go**
  (see Option 4) — pursue handles instead.
- Threaded dispatch via a `[]func()` jump table (Go has no computed goto; this
  trades the switch for indirect calls — measure, it can lose).
- Register-based bytecode. Large effort; defer until Steps 1–3 are exhausted.

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
