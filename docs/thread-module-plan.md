# Plan: `thread` module — in-process concurrent workers

## Context

`process`/`pool` give glox worker-based concurrency, but every worker is a
full OS process (spawned via `exec.Command`) communicating over a pipe with
pickle-serialized messages — heavyweight, and `spawn()` is forced to take a
`.lox` script path rather than an in-memory closure because there's no way
to serialize a closure across a process boundary.

This followed from a design discussion after building `pool`
(`src/modules/pool.lox`): could glox instead run multiple `*vm.VM`
instances as goroutines within a single OS process? Cheaper workers, and
critically, workers that accept an actual closure (`thread.spawn(func() {
... })`) instead of a script path, since no serialization is needed within
one address space.

Investigation (direct code reading, confirmed at every citation below)
found the `VM` struct itself is already well-shaped for this — its
execution state (stack, frames, `BuiltIns`) is all per-instance, not
package-global. The real work is: (a) fixing a handful of currently-unsafe
shared global state every VM touches, (b) building a deep-copy discipline
so a closure handed to a new thread doesn't let two VMs race on the same
mutable data, and (c) a new module surface mirroring `process`'s shape but
with different (weaker, and explicitly documented) fault-isolation and
cancellation guarantees, since a goroutine can't be crash-isolated or
force-killed the way an OS process can.

**Important scope limitation, discovered during design and worth stating
up front**: only a spawned closure's own *captured (upvalue) locals* are
isolated (deep-copied) by this design. Top-level `var`s, class statics, and
module attributes are **not** isolated — they live in a `*core.Environment`
shared by every function compiled from the same script/module (confirmed:
`FunctionObject.Environment` is one object shared by a whole compilation
unit, not per-closure). Making that fully isolated too would require
transitively cloning the whole reachable environment graph, which directly
contradicts the (correct, kept) decision to share `*ClassObject` by pointer
across threads — a class's methods close over that same shared environment.
v1 accepts this and documents it plainly: **global/module/class-level
mutable state is shared across threads and not synchronized for you** —
same category of caveat as the existing raylib carve-out, not silently
papered over. This plan also adds a `sync` module (`Mutex`, see below) as
the one tool a script has for making that shared state safe to touch from
more than one thread, when it actually needs to.

## Concurrency prerequisites (do first — nothing else in this plan is safe without these)

Four spots of currently-unsynchronized package/shared state, all one small
fix each:

1. **`src/core/string_intern.go:3-6`** (`nameToID`/`idToName`) — every
   `InternName` call is a raw map write with no lock; hit on essentially
   every identifier/dict-key/method-name lookup. Fix: package-level
   `sync.RWMutex`; `RLock` fast path on hit, upgrade to write lock only on
   a genuinely new name.
2. **`src/vm/vm.go:80-81`** (`globalModuleSource`/`globalModules`, the
   compiled-module cache, written in `importModule` around `:2472`/`:2504`,
   read around `:2436`/`:2460`) — same fix, a `sync.Mutex` around the
   read-check/write; two threads compiling the same not-yet-cached module
   concurrently is acceptable wasted work, not a bug worth preventing.
3. **`src/core/obj_list.go:151`** (`var stringDepth int`, a recursion guard
   for stringifying nested/self-referential lists) — convert to
   `sync/atomic.Int32`. Under concurrent load two threads printing deeply
   nested structures simultaneously could trip the depth guard a little
   early; cosmetic, not a correctness bug, not worth a full lock on a
   `String()` hot path.
4. **`src/core/environment.go`** (`Environment.Vars map[int]Value`,
   read/written unsynchronized by `GetVar`/`SetVar`) — **this one is not
   optional**: `module.attr = x` compiles to `OP_SET_PROPERTY` calling
   `Environment.SetVar` directly (a bare map write). Since every built-in
   module's `Environment` is shared by reference across the parent and
   every spawned thread (same sharing the existing module-import `subvm`
   precedent already relies on), two threads both writing a module
   attribute hits Go's **fatal, unrecoverable** concurrent-map-write
   detector — a process crash `recover()` cannot catch, not just a race.
   Fix: add `mu sync.RWMutex` to `Environment`, lock in `GetVar`/`SetVar`.
   (`Globals`/`Defined` slice element writes via `SetGlobal` are a milder,
   silent-race version of the same sharing — left as part of the
   documented "globals aren't isolated" limitation above, not fixed here,
   since REPL's `GrowGlobals` reallocation — the one place these slices
   really move — is sidestepped by rejecting `spawn()` from the REPL, see
   below.)

## Core deep-copy mechanism

New file **`src/core/copy.go`**, structurally parallel to `pickle.go`'s
`encodeObjectValue` but constructing fresh objects instead of writing
bytes, and using a memo `map[Object]Object` (original → copy) instead of
pickle's reject-on-cycle `visiting` set — a self-referential structure
correctly copies into an equally self-referential copy instead of
erroring.

```go
func CopyValueForSpawn(v Value, memo map[Object]Object) Value
func CopyClosureForSpawn(closure *ClosureObject) *ClosureObject
func copyUpvalueForSpawn(uv *UpvalueObject, memo map[Object]Object) *UpvalueObject
```

Rules:
- `VAL_NIL/BOOL/INT/FLOAT`: return as-is (already value types).
- `VAL_VEC2/VAL_VEC3/VAL_VEC4`: **must be cloned**, not aliased — these
  carry their payload through `Value.Obj` like `VAL_OBJ` does, but are
  mutable in place (`OP_SET_PROPERTY` on a vec calls e.g. `o.SetW(...)`).
  Easy to miss if only branching on `v.Type == VAL_OBJ`.
- `*StringObject`: share (interned, immutable).
- `*ClassObject`: share by pointer (deliberate — see the scope limitation
  above; class defs/statics are not isolated).
- `*ListObject`/`*DictObject`: clone the container, recurse into items/values.
- `*InstanceObject`: clone `{Class: shared, Fields: cloned+recursed}` — no
  pickle-style by-name class resolution needed at all, since both VMs share
  the same compiled, live `*ClassObject`.
- `*ClosureObject`: clone `{Function: shared (immutable bytecode), Upvalues:
  each copied via copyUpvalueForSpawn}`.
- `*BoundMethodObject`: clone `{Receiver: recursed, Method: shared}`.
- Anything else (module/file/native/graphics handles): share by pointer —
  same caveat class as raylib, just narrower in scope.
- `copyUpvalueForSpawn`: `*uv.Location` always yields the current captured
  value regardless of open/closed state (confirmed:
  `UpvalueObject{Location *Value, Closed Value}`, `Location` points at the
  live stack slot while open or at `&u.Closed` once closed). The copy is
  born **already closed** — `clone.Closed = CopyValueForSpawn(*uv.Location,
  memo); clone.Location = &clone.Closed` — and is never linked into any
  VM's `openUpValues` list, since it's a one-time snapshot, not a live
  binding.

## VM plumbing

**`src/core/thread.go`** (new) — two small types, one per side of a spawn,
since the parent and worker views have non-overlapping capabilities
(`wait`/`cancel` only make sense from the parent):

```go
type ThreadMessage struct { Val Value; Err error }

type ThreadChannels struct { // worker's view
    In        <-chan Value
    Out       chan<- ThreadMessage
    Cancelled <-chan struct{}
}

type ThreadHandle struct { // parent's view
    ToWorker   chan<- Value
    FromWorker <-chan ThreadMessage
    Done       <-chan struct{}
    Cancel     func()
    Err        error
    Result     Value
}
```

**`src/core/object.go`**: extend `VMContext` with
`SpawnThread(closure Value, args []Value) (*ThreadHandle, error)`,
`ThreadChannels() (*ThreadChannels, bool)` (`ok` false unless this VM was
itself created by `SpawnThread`), and `CallClosure(closure Value, args
[]Value) (Value, error)` — same pattern as `ResolveClass` was added for
pickle, since `src/builtin` can't reach `*vm.VM`'s unexported `call`/`run`
directly (would need to import `vm`, which imports `builtin` — cycle).
`CallClosure` is the synchronous case (no new VM, no copy, no goroutine —
just push+call+run on the *current* VM, reusing exactly the shape
`callLoadedChunk`/`runThreadWorker` already use); it exists as its own
primitive because the `sync` module's `Mutex.locked()` below needs it too,
and `runThreadWorker`'s body should itself just call it instead of
duplicating the push+call+run sequence a second time. Also append
`NATIVE_THREAD`/`NATIVE_THREAD_CHANNEL` to the `NativeType` const block
(append only, don't renumber existing entries).

**`src/vm/vm.go`**: one new unexported field, `threadChans *core.ThreadChannels`
(nil unless this VM was created by `SpawnThread`).

**`src/vm/thread.go`** (new):
```go
func (vm *VM) SpawnThread(closureVal core.Value, args []core.Value) (*core.ThreadHandle, error)
func (vm *VM) ThreadChannels() (*core.ThreadChannels, bool)
func runThreadWorker(worker *VM, closure *core.ClosureObject, args []core.Value,
    handle *core.ThreadHandle, fromWorker chan core.ThreadMessage, doneCh chan struct{})
```
`SpawnThread`: rejects if `vm.Repl` (see REPL note below) or if `closureVal`
isn't a `*core.ClosureObject`; deep-copies the closure and each arg (own
memo maps); builds buffered channels (`make(chan ..., 16)`, matching
`process`'s `recvCh` buffer size), a `context.WithCancel`, and a `doneCh`;
constructs `worker := NewVM(vm.script, false)` sharing `BuiltIns`/
`BuiltInModules`/`Args()` — the exact pattern the existing module-import
`subvm` precedent already uses (`src/vm/vm.go:2473-2477`); sets
`worker.threadChans`; launches `go runThreadWorker(...)`.

`runThreadWorker` mirrors `callLoadedChunk`'s push+call+run shape
(`src/vm/vm.go:2517-2528`), run on the copied closure instead of a loaded
`.lxc` chunk, wrapped in its own `recover()` — **mandatory**, since nothing
else catches a non-main-goroutine panic (confirmed: no `panic`/`recover`
inside `run()` itself; the only `recover()` is in `main.go`'s `runFile`,
guarding just the one calling goroutine). On panic, or on an unhandled Lox
exception escaping `run()` (`INTERPRET_RUNTIME_ERROR`), send a
`ThreadMessage{Err: ...}` on `fromWorker` and store `handle.Err`; on a
normal return, store `handle.Result` from `run()`'s second return value.
Close `fromWorker` then `doneCh`, in that order — the happens-before
guarantee of channel close is the only synchronization `wait()`/`recv()`
need, no mutex.

## Builtin module surface (mirrors `process`'s 3-file split)

**`src/builtin/thread_functions.go`** (new): `SpawnBuiltIn` (validates arg
0 is a closure, calls `vm.SpawnThread`, wraps the handle in `ThreadObject`,
registers methods) and `ChannelBuiltIn` (calls `vm.ThreadChannels()`,
raises `ThreadError` if not inside a spawned thread, else wraps in
`ThreadChannelObject`).

**`src/builtin/obj_builtin_thread.go`** (new): `ThreadObject{core.BuiltInObject;
Handle *core.ThreadHandle; Methods map[int]*core.BuiltInObject}` and
`ThreadChannelObject{core.BuiltInObject; Chans *core.ThreadChannels;
Methods ...}` — two types rather than one dual-purpose type, since the
parent and worker sides wrap different underlying Go types with
non-overlapping methods. Each implements `String()`/`GetType()`/
`GetNativeType()`/`GetMethod()`/`RegisterMethod()`/`IsBuiltIn()`, same
shape as `ProcessObject`.

**`src/builtin/thread_methods.go`** (new):
- `Thread` (parent side): `send(v)` (select on `ToWorker` vs. `Done`,
  raises `ThreadError` if the thread's already finished), `recv()`/
  `try_recv()` (mirror `process`'s tri-state pattern), `wait()` (blocks on
  `Done`, then raises `ThreadError` if `Handle.Err != nil` else returns
  `Handle.Result` — giving threads a real return value, unlike processes),
  `cancel()` (calls `Handle.Cancel()`; returns immediately).
- `ThreadChannel` (worker side, from `thread.channel()`): `send`/`recv`/
  `try_recv`, identical shape but selecting against `Cancelled` instead of
  `Done`.

**`cancel()` is cooperative only** — it unblocks a worker currently parked
in a `channel().send`/`recv` call (both select against `Cancelled`), but
cannot interrupt a worker stuck in a tight non-channel loop; there's no
instrumented preemption point in `run()`, and adding one is out of scope.
Document this plainly as a real limitation, not a soft edge case — it's
the direct consequence of no longer having an OS-level `kill()`.

**`ThreadError`** class: add next to `ProcessError` in `exceptionSource`
(`src/vm/builtin.go`), same shape (`msg`/`name`/`toString()`). Raised for:
non-closure `spawn()` argument, worker arity mismatch, an unhandled Lox
exception or recovered panic surfacing via `wait()`/`recv()`, `channel()`
called outside a spawned thread, and send/recv after the other side is
gone.

**Registration** (`src/vm/builtin.go`, alongside existing `process` wiring):
`makeBuiltInModule(vm, "thread")`, `defineBuiltIn(vm, "thread", "spawn",
builtin.SpawnBuiltIn)`, `defineBuiltIn(vm, "thread", "channel",
builtin.ChannelBuiltIn)`.

## REPL restriction

`SpawnThread` rejects (raises `ThreadError`) if `vm.Repl` is true. Reason:
the REPL's `Environment.GrowGlobals` (`src/vm/vm.go:129-134`) *reallocates*
`Globals`/`Defined` on every new line, but `run()` caches a slice header
from them per frame push — a thread still running when the user enters
another REPL line would be working against a stale backing array, a
split-brain bug, not just a race. Simplest correct fix is disallowing it
at the boundary rather than trying to make `GrowGlobals` thread-aware.

## `sync` module: `Mutex` for shared global state

The scope limitation stated up top — globals, class statics, and module
attributes are shared across threads and **not** isolated by the deep-copy
mechanism — is deliberate, but it means a script that actually wants two
threads to safely share and mutate one of those needs a real synchronization
primitive; today there's nothing to reach for. No new statement/keyword is
needed for this — glox already has `finally` (confirmed: a real reserved
token, `docs/language-reference.html`'s exception-handling row lists `try /
except / finally / raise`), so a plain `acquire()`/`release()` pair used
with `try { ... } finally { m.release(); }` is already a safe, standard
pattern, the same idiom Java's `Lock` uses. This fits the rest of this
plan's approach exactly: new capability = a module + a native object, no
compiler/grammar changes.

**`src/builtin/sync_functions.go`** (new): `MutexBuiltIn` — `sync.Mutex()`
constructs a `MutexObject` wrapping a real Go `sync.Mutex`.

**`src/builtin/obj_builtin_sync.go`** (new):
`MutexObject{core.BuiltInObject; mu sync.Mutex; Methods map[int]*core.BuiltInObject}`.
**Important**: `CopyValueForSpawn`'s existing "anything else: share by
pointer" default bucket already does the right thing here without any
special-casing — a `Mutex` captured by a spawned closure's upvalue must be
*shared*, not cloned, or every thread ends up locking its own private copy
and the lock stops meaning anything. Worth a one-line comment at that
default case in `copy.go` calling this out explicitly, so a future change
to "clone everything by default" doesn't silently break it.

**`src/builtin/sync_methods.go`** (new):
- `acquire()` → `mu.Lock()`.
- `release()` → `mu.Unlock()`, wrapped in its own `recover()` — Go's
  `sync.Mutex.Unlock()` on an already-unlocked mutex panics, which should
  surface as a catchable `SyncError` (new exception class, same shape as
  `ProcessError`/`ThreadError`), not crash the calling thread's goroutine.
- `locked(closure)`: `mu.Lock(); defer mu.Unlock()`, then
  `vm.CallClosure(closure, nil)` inside that defer's scope — guarantees the
  unlock runs even if the closure raises or panics, without the caller
  needing to remember `finally` at all. The convenience form for the
  common single-critical-section case; `acquire()`/`release()` stay
  available for a lock that needs to span multiple statements or
  functions, matching Go's own `sync.Mutex` offering both styles.

**`SyncError`** class: add next to `ThreadError` in `exceptionSource`
(`src/vm/builtin.go`). Raised for: `release()` without a matching
`acquire()`, and any panic escaping a `locked()` closure that isn't already
a Lox exception.

v1 scope is a single `Mutex` — no `RWMutex`/semaphore/`WaitGroup`-equivalent
yet, same tight-scoping instinct as the rest of this plan. Natural follow-ons
if `Mutex` alone proves insufficient in practice.

## Tests

New `tests/new_tests/lox/thread_*.lox` + `tests/new_tests/test_thread.py`
(same `run_lox`/`@pytest.mark.parametrize("force_compile", ...)` pattern as
`test_process.py`):

- **`thread_basic.lox`**: spawn a closure, `thread.channel()` inside it,
  send/recv/wait round trip.
- **`thread_isolation.lox`**: capture a **local** (not global — a global
  version would correctly show no isolation, which isn't a bug, just the
  documented scope limit) in the spawned closure; both the parent and the
  worker mutate their own copy after spawn; assert neither observes the
  other's mutation. Include a captured **list** (not just a scalar) to
  exercise the `ListObject` clone path.
- **`thread_panic.lox`**: worker raises an uncaught Lox exception; parent's
  `try { t.wait(); } except ThreadError as e { ... }` proves it surfaces
  cleanly instead of crashing the process.
- **`thread_cancel.lox`**: worker blocks in `channel().recv()`; parent
  calls `cancel()`; asserts the worker's `recv()` raises `ThreadError`
  rather than hanging, and `wait()` returns promptly.
- **Go-level** `src/vm/thread_test.go`: a deliberately-malformed closure
  (e.g. empty `Chunk.Code`, making `run()`'s first instruction fetch an
  out-of-range index panic) to exercise a genuine Go panic — hard to
  construct from valid Lox source, since Lox-level recursion is already
  capped by `FRAMES_MAX` before it could overflow the Go stack — asserting
  `wait()`/`Handle.Err` surfaces it and the test process itself survives.
- **`sync_mutex.lox`**: spawn several threads (e.g. 8) that each acquire
  the same `Mutex`, increment a value read via `thread.channel()` from a
  shared source, and send the result back — with the lock, the parent
  reconstructs a correct, non-racy final tally; a variant without
  `acquire()`/`release()` at all is *not* included as a "must fail" test
  (a race isn't guaranteed to manifest deterministically, so asserting on
  it would be flaky) — the case for `Mutex` is made in the docs, not by a
  test designed to prove the absence of one is broken.
- **`sync_mutex_finally.lox`**: a closure run via `locked()` raises
  partway through; asserts a *different* thread can still successfully
  `acquire()` the same mutex afterward (proving the unlock-on-panic
  guarantee actually holds, not just that the happy path releases it).

## Docs

`docs/md/THREAD_MODULE.md` (new, mirroring `PROCESS_MODULE.md`'s
structure) and a `language-reference.html` section + nav link + exceptions
table entry for `ThreadError`, per this repo's own convention for
documenting new modules (CLAUDE.md). Explicitly state the "globals/class
statics are shared, not isolated" limitation and the "cancel() is
cooperative, not a real kill" limitation in both places — these are the
two facts a user of this module most needs to not get wrong.

`docs/md/SYNC_MODULE.md` (new, same structure) + a `language-reference.html`
section + nav link + exceptions table entry for `SyncError`, covering both
`acquire()`/`release()` and `locked()`, and stating plainly that `Mutex` is
the *only* tool this plan gives you for the "globals/class statics are
shared, not isolated" gap — nothing else makes that shared state safe on
its own.

## Verification

- `go build -o bin/glox main.go` after each phase.
- `go test ./src/vm/...` for the new `thread_test.go` panic-recovery case.
- `cd tests && python3 -m pytest new_tests/ -q` for the full suite
  including new `test_thread.py` cases, both `force_compile` values.
- Manually run `thread_isolation.lox` and `thread_cancel.lox` a few times
  in a loop (`for i in 1 2 3 4 5; do ./bin/glox ...; done`) to catch
  goroutine-scheduling-order flakiness before trusting a single green run,
  same diligence applied to `pool_reuse.lox` earlier this session.
