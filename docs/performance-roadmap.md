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
- **Forced stop-the-world collections.** Two `runtime.GC()` calls used to fire:
  a *periodic* one in the dispatch loop (every 65 536 instructions, on a 5 s
  interval) and one on *every top-level return*. The periodic loop GC has been
  **removed** — Go's own pacer handles heap growth, so the forced collection was
  pure overhead. The top-level-return GC ([`src/vm/vm.go`](../src/vm/vm.go),
  `OP_RETURN` `frameCount == 0` branch) still fires each time the outermost frame
  unwinds (also per module import / REPL line); gating it to debug-only remains a
  cheap win.
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

✅ **Done.** `--cpuprofile <file>` and `--memprofile <file>` flags added to
`main.go`, writing standard `runtime/pprof` CPU and heap-allocation profiles
(open with `go tool pprof -top -focus="glox/" bin/glox.exe <file>`).

Ran against `trees` and `method_call`. Results confirm family A as the
dominant *attributable* cost, and family B close behind:

- **`trees`** (map-backed instance fields, deep object graphs):
  `runtime.mapaccess2_fast64` + `internal/runtime/maps.ctrlGroup.matchFull`
  together ≈ **32% of cumulative CPU** — this is field reads/writes and
  method lookup going through `map[int]Value` / `class.Methods[id]`, exactly
  the Step 2 hypothesis. Allocation profile: **`MakeInstanceObject` is 66.6%
  of all allocated objects** — every `Tree(...)` call allocates both the
  struct and an empty `map[int]Value{}` (family B, and the direct driver of
  family A's map cost).
- **`method_call`** (field + method access, shallower graph, tighter loop):
  `mapaccess2_fast64`/`mapassign_fast64`/`ctrlGroup.matchFull` together ≈
  12% cumulative, `invoke`/`invokeFromClass` ≈ 14% cumulative — smaller
  relative share than `trees` (no per-call instance allocation here, just
  repeated field/method lookup), but the same map-backed mechanism.
  `Value.IsStringObject` also shows up, confirming the family-C note about
  the `<`/`>` fast-path check.

**Both profiles confirm Step 2 (slot-based instance fields) as the
highest-value next fix** — map access/assign plus the per-instance map
allocation it requires are the largest single attributable cost in both
object-heavy benchmarks.

**Windows profiling caveat:** on this machine, CPU profiles also show
`runtime.stdcall1`/`notewakeup`/`schedule`/`stoplockedm` at a combined
~30–34% of cumulative time in `trees`. This is very likely a **sampling
artifact of `pprof` on Windows** (each sample suspends/resumes the thread via
`SuspendThread`/`GetThreadContext`, which is itself a `stdcall`), not real
interpreter cost — confirmed by two A/B wall-clock tests that each isolate a
candidate cause and find no effect: `GODEBUG=asyncpreemptoff=1` (24.5s vs
24.3s baseline) and a build with `runtime.LockOSThread()` removed from
`main.go`'s `init()` (24.9s vs 25.4s baseline) — both within noise. Filter
these frames out with `-focus="glox/"` when reading a profile on Windows;
don't chase them as a real cost.

### Step 1 — Cheap, high-confidence wins

- ✅ **Periodic loop `runtime.GC()` removed** (was every 65 536 instructions on a
  5 s interval). The remaining forced GC on top-level return
  ([`src/vm/vm.go`](../src/vm/vm.go), `OP_RETURN` `frameCount == 0`) can still be
  gated to debug-only; measure `binary_trees` / `trees` before and after.
- ✅ **Collection and string method tables cached package-level** (shared,
  keyed by interned id) instead of rebuilding them per object. `ListObject`,
  `DictObject`, and `StringObject` no longer carry a per-instance
  `Methods map[int]*BuiltInObject`; each now has a package-level
  `listMethods`/`dictMethods`/`stringMethods` map built once in an `init()`,
  and each method recovers its receiver from `vm.Stack(arg_stackptr - 1)`
  instead of closing over the specific object.

  Added `benchmarks/lox/collections.lox` (+ `benchmarks/python/collections.py`)
  to the suite — the loxcraft benchmarks are all class/method-call-heavy and
  don't exercise this path, so this fix was otherwise invisible to
  `bin/benchmarks.sh`. Pre/post (3-run average, same machine):

  | phase | before | after | speedup | CPython 3 | after/CPython |
  |---|---|---|---|---|---|
  | list | 4.75s | 3.44s | 1.38× | 0.93s | 3.68× |
  | dict | 7.33s | 5.04s | 1.45× | 1.18s | 4.27× |
  | string | 2.82s | 2.34s | 1.20× | 0.73s | 3.19× |
  | **total** | **14.89s** | **10.82s** | **1.38×** | **2.85s** | **3.80×** |

  A purer synthetic (2M iterations of just `var l = []; l.append(i);`, no
  other work) showed a larger ~2.4× wall-time win and ~5.8× fewer allocated
  objects (23.9M → 4.1M — was ~12 allocs per list: 4 closures + map + map
  inserts; now 2: the list struct + the append's backing-array grow). The
  smaller 1.2–1.45× win in `collections.lox` reflects that real method calls
  are diluted by loop dispatch, index-assignment, and `len()` overhead not
  touched by this change — representative of mixed real-world code rather
  than an allocation-only microbenchmark.
- ✅ **Per-instruction `DebugHook` check hoisted out of the hot loop —
  via a source-toggled fast/debug build, not loop duplication.** Measured
  first: removing the check entirely cost `fib` (call-heavy) only −2.1% but
  cost `loop` (pure dispatch, no calls/allocation) **−25.0%** (8.51s →
  6.38s, 3-run averages). The cheap version — swap `vm.DebugHook != nil` for
  a plain `bool` field — recovered **none** of it (8.53s, statistically
  identical to baseline): the cost isn't the comparison, it's the mere
  presence of a branch at that point perturbing the Go compiler's codegen
  for the dispatch switch. A true fix needs the branch physically absent
  from the compiled hot path, and full loop duplication (~1000+ lines kept
  in sync by hand) was too much risk for the payoff.

  Landed a lighter-weight version of the same idea: the hook line in
  `src/vm/vm.go`'s `run()` is commented out by default (`go build -o
  bin/glox main.go`, unchanged), and `bin/build_debug.sh` mechanically
  uncomments that exact line plus flips `core.HotLoopDebugHookCompiled` in
  `src/core/config.go`, builds `bin/debug_glox`, then restores both files
  (via a `trap ... EXIT`, so it cleans up even if the build fails) — the
  working tree is never left modified. `main.go` warns on stderr
  (`warnIfNoDebugHook`) if `--debug`/`--info`/`--instrument` are used on a
  build where `HotLoopDebugHookCompiled` is false, instead of silently
  producing empty trace/zero instruction counts. Confirmed the fast default
  now gets the full win (`loop.lox`: 6.39s, matching the A/B "removed
  entirely" measurement) and the debug build's `-i` correctly reports
  non-zero instruction counts. All 176 `new_tests` still pass on the fast
  build.

  One sharp edge hit and fixed along the way: `sed -i` on this platform
  silently flattens CRLF → LF across the *whole* file on any in-place edit
  (Windows-originated files here use CRLF) — `bin/build_debug.sh` uses
  `sed -i -b` (binary mode) to avoid this; a plain `sed -i` on a CRLF file
  in this repo will produce a spurious whole-file diff.

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
- Re-run the profiler (`bin/glox --cpuprofile out.prof --memprofile out.mprof
  benchmarks/lox/<name>.lox`, then `go tool pprof -top -focus="glox/"
  bin/glox.exe out.prof`) on the same benchmark and confirm the targeted
  symbol shrank. On Windows, always pass `-focus="glox/"` — see the Step 0
  caveat about scheduler-frame sampling noise.
- `bin/benchmarks.sh 3` for stable ratios; update the README Performance table.
- Full correctness gate: `python -m pytest tests/new_tests/ -x -q` (with
  `LOX_PATH` and `bin` on `PATH`). Slot-based fields and shared method tables
  are behaviour-sensitive — the OO tests must stay green.
