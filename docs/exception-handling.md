# Exception handling: bytecode shape, VM dispatch, and the finally trampoline

How `try`/`except`/`finally` actually works, end to end — the bytecode layout,
the VM-side matching machinery, and why the compiler-side `finally` support
looks the way it does. Grounded in the current code:
[`src/compiler/compile.go`](../src/compiler/compile.go) (`tryExceptStatement`
and friends), [`src/vm/vm.go`](../src/vm/vm.go) (`raiseException`,
`nextHandler`, the `OP_TRY`/`OP_END_TRY`/`OP_EXCEPT`/`OP_END_EXCEPT`/
`OP_FINALLY`/`OP_RAISE` dispatch cases), and
[`src/core/types.go`](../src/core/types.go) (`ExceptionHandler`).

## Bytecode shape

```
try {
    <try body>
}
except TypeA as a {
    <clause A body>
}
except TypeB as b {
    <clause B body>
}
finally {
    <finally body>
}
```

compiles to, roughly:

```
        OP_TRY          <exceptIP>      ; exceptIP = address of the first OP_EXCEPT
                                         ; (or straight at OP_FINALLY if no except)
        <try body>
        OP_END_TRY      <offset>        ; normal completion: pop handler, jump to <A>
        OP_EXCEPT       <"TypeA">
        <clause A body>
        OP_JUMP         <A>             ; normal completion of this clause
        OP_END_EXCEPT
        OP_EXCEPT       <"TypeB">
        <clause B body>
        OP_JUMP         <A>
        OP_END_EXCEPT
        OP_FINALLY                      ; always-matching catch-all/reraise handler
        <finally body>  (copy #2)
        OP_GET_LOCAL    <exc slot>
        OP_RAISE
   <A>: <finally body>  (copy #1 — the shared normal-completion landing point)
        ...
```

Two invariants the VM depends on, both easy to break by accident:

- **`OP_END_EXCEPT` must be immediately followed by the next clause's
  `OP_EXCEPT` or `OP_FINALLY`, with nothing in between.** `nextHandler()`
  (`src/vm/vm.go:2420-2437`) finds "the next clause" by scanning forward
  byte-by-byte for this exact adjacency — it has no other way to locate
  clause boundaries. `tryExceptStatement()`'s per-clause exit jump
  (`OP_JUMP <A>`) is emitted *before* `OP_END_EXCEPT`, not after, specifically
  to preserve this.
- **`OP_TRY`'s operand is patched exactly once**, to the address of the
  *first* clause (`OP_EXCEPT` or, for a bare `try { } finally { }`,
  `OP_FINALLY` directly) — never re-patched as later clauses compile. This
  used to be a bug (patched once per clause, so it always ended up pointing
  at the *last* one); see "Bugs fixed by this design" below.

## The VM-side matching loop

`OP_TRY` (`vm.go:1420-1433`) pushes an `ExceptionHandler{ExceptIP, StackTop,
Prev}` onto `frame.Handlers` — a cons-list, so nested `try`s just push more
entries. `OP_END_TRY` (`vm.go:1434-1441`) pops one entry and jumps past the
whole except/finally chain on normal completion; the jump offset is real and
must be consumed (it wasn't, once — see below).

`raiseException` (`vm.go:2333-2418`) is the whole matching engine:

```go
for {
    for handler := vm.frame().Handlers; handler != nil; handler = handler.Prev {
        vm.stackTop = handler.StackTop
        vm.stack[vm.stackTop] = err; vm.stackTop++
        vm.frame().Ip = int(handler.ExceptIP)
        for {  // "inner": walk this try's own clause chain
            if <at OP_FINALLY> { pop handler, return true }
            <read classname constant, resolve to a class value>
            if err_class.IsSubclassOf(handler_class) { pop handler, return true }
            if !nextHandler() { break }  // no more clauses in *this* try
        }
        // handler = handler.Prev: fall back to the next enclosing try
        // in the *same frame* before giving up on it
    }
    if !popFrame() { report "Uncaught exception"; return false }
    // retry from the top in the caller's frame
}
```

Two things worth calling out because they weren't always true:

- **`OP_FINALLY` is just an always-matching clause**, slotted into the exact
  same `Handlers`/`nextHandler()` chain `except` clauses use. This is why
  `finally` correctly runs even when the exception escapes a *nested function
  call* made from inside the `try` — `popFrame()` unwinds the call stack one
  frame at a time, and each ancestor frame's own `Handlers` (including any
  `OP_FINALLY`) gets a fair chance via the same loop above. A pure
  compile-time jump-splice, with no VM-level participation, could not do
  this — it only ever sees the current function's own bytecode.
- **A handler whose own clause chain doesn't match now falls back to
  `handler.Prev`** (an enclosing `try`, if any, *in the same frame*) before
  the whole loop gives up and unwinds to the caller frame. This is the
  `for handler := ...; handler != nil; handler = handler.Prev` structure
  above; it used to be a plain `if`, which meant a nested `try` whose own
  `except` didn't match skipped straight past a syntactically-enclosing,
  actually-matching outer `try` in the same frame and reported the exception
  as uncaught. See "Bugs fixed by this design" below.

### Resolving the exception class name

The classname lookup in the `inner` loop tries, in order: the current
function's own `Environment.GetVar` (checks `Environment.Vars`, the
module-export/import-all binding table — not locals or upvalues),
`vm.BuiltIns`, then a
fast-globals-slice fallback. That last one used to consult
`function.Chunk.SlotForName(name)` — but `Chunk.GlobalNames` is **only ever
populated for the top-level script's own chunk** (`endCompiler()`,
`compile.go` — "inner function chunks don't own globals, they're all in the
script's environment"). So an `except SomeUserClass` clause compiled inside
*any* other function always failed this lookup, regardless of where the
`raise` came from. The fix was `Environment.SlotForName`
(`src/core/environment.go:40-47`), which scans `Environment.GlobalNames` —
populated once from the top-level chunk and then shared by every function in
the compilation unit (the same table `Environment.NameForSlot` already used
for the reverse direction, in error messages). Built-in exception classes
(`RunTimeError`, `SyncError`, ...) were never affected, since they resolve
through `vm.BuiltIns` first.

## Why `finally` recompiles its body instead of duplicating bytecode

A `finally` block needs to run at several different points: the shared
normal-completion path, the always-matching `OP_FINALLY` handler, and once
per `return`/`break`/`continue` that crosses it. Each of these needs its own
copy of the compiled body — but copying raw bytecode between them doesn't
work, because every `OP_GET_LOCAL`/`OP_SET_LOCAL`/upvalue-index operand is an
*absolute* frame-relative slot number, fixed at the point the code was
originally compiled (`addLocal`, `c.localCount`). Splicing the same bytes in
at a different point (different `c.localCount` context) would require
relocating every such operand — a linker-style fixup, and a fragile one.

Instead, the parser snapshots the `finally` block's *token position*
(`parserSnapshot`, `compile.go:58-61`; `snapshotPos`/`restorePos`,
`compile.go:442-457`) and **recompiles it from source** at each landing
site. This is safe because `Scanner.Tokens` is fully materialized up front
(`NewScanner`) and `NextToken()` is a pure index into that fixed slice —
replaying is side-effect-free, and each replay gets correct locals/upvalues
for whatever scope it's compiled into.

## The trampoline: why return/break/continue can't just splice inline

The obstacle here is ordering, not logic: `finally` is the *last* thing
`tryExceptStatement()` parses, after the `try` body and every `except`
clause. So a `return`/`break`/`continue` written inside the `try` body is
compiled *before* the parser even knows whether a `finally` follows it, let
alone what's in one. It can't decide there and then whether to emit a plain
jump/return or "run cleanup first" — that decision, and the cleanup code
itself, both have to wait.

`TryFinally` (`compile.go:67-82`) is pushed onto `Compiler.tries` (mirroring
how `Loop` tracks enclosing loops) when a `try` is entered, and every
`return`/`break`/`continue` that might be crossing one defers into its
`pending` list as a `trampolineSite` (`compile.go:84-92`) — *unconditionally*,
regardless of whether `hasFinally` turns out true, since that isn't known yet
either. Once `tryExceptStatement()` actually reaches `finally` (or confirms
there isn't one), `compilePendingTrampolines` (`compile.go:673-740`) resolves
every deferred site: patch its jump, replay the `finally` body if there is
one, then either chain onward (if the same jump also needs to cross a
further *outer* `try`) or finally emit the real terminal instruction
(`OP_RETURN`, the loop's real `break`/`continue` jump).

A few supporting details worth knowing if touching this code:

- **`return` doesn't need per-local popping** — `OP_RETURN` already does an
  O(1) wholesale `stackTop = frame.Slots` reset regardless of what's above
  it, so crossing a `finally` on the way out just needs the cleanup code to
  run first, not any stack bookkeeping. The return value is protected across
  that gap by anchoring it in a synthetic local (`__retval`) *before* the
  (not-yet-parsed) `finally` block can declare locals of its own — see
  `localCountAtCrossing` below.
- **`break`/`continue` do need it**, since the frame survives and the same
  code can re-fire every loop iteration — `crossTries` (`compile.go:742-758`)
  emits `OP_END_TRY, 0, 0` (pop one handler, jump zero bytes) for *every*
  `try` crossed, whether or not it has a `finally`. This is also the fix for
  a second bug: `break`/`continue` never used to touch `frame.Handlers` at
  all, leaving a stale handler (with a stack height from inside the
  now-exited loop iteration) that could wrongly intercept a later, unrelated
  exception.
- **`localCountAtCrossing`** on each `trampolineSite` records the exact local
  count (== runtime stack height relative to `frame.Slots`) at the moment
  that jump lands — constant across every hop of a chain, since each hop's
  own `finally` replay pushes and pops in balance. `compilePendingTrampolines`
  reserves dummy locals up to that count before replaying, so the replayed
  body's own locals can't alias whatever's still live higher on the real
  stack (most importantly, a pending return's `__retval`).
- **A crossed `try`'s own normal-completion path must jump *past* the
  trampolines compiled right after it** (`compileTrampolinesAfterNormalPath`,
  `compile.go:650-671`) — they sit next to each other in the bytecode stream,
  so without this jump, ordinary fallthrough would run straight into
  break/continue/return cleanup code that isn't relevant to it.

## Known limitation

If an `except` clause's own body raises a fresh exception, the enclosing
`finally` does **not** run for it — only exceptions escaping the `try` body
itself (uncaught here) and normal/`return`/`break`/`continue` exits are
covered. Covering this too would mean wrapping every `except` clause body in
its own private nested `try` tied to the same `finally`, multiplying compiled
copies and trampoline chains — deliberately out of scope.

## Bugs fixed by this design (for context, not because they're expected to recur)

- `OP_TRY`'s operand was patched once per `except` clause instead of once,
  so it always ended up pointing at the *last* clause.
- Falling through a `try` body without raising, when an `except` clause was
  present, panicked the VM outright — `OP_END_TRY` never consumed its own
  jump offset.
- After a non-last `except` clause matched and completed, execution fell
  through into the next clause's own body and ran that too.
- `break`/`continue` never popped `frame.Handlers` (see above).
- A nested `try` whose own clauses didn't match unwound straight to the
  *caller* frame instead of falling back to an enclosing `try`'s handler in
  the same frame (see "same frame" fallback above).
- `except SomeUserClass` failed to resolve inside any non-top-level function
  (see "Resolving the exception class name" above).
