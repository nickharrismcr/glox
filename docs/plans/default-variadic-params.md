# Plan: Default & variadic parameters

Status: **implemented** (shipped). Kept for historical reference. Implementation
matches this plan: `UNDEFINED` sentinel + `OP_JUMP_IF_DEFINED` prologue guards,
`MinArity`/`IsVariadic` on `FunctionObject`, call-time arg shaping in `vm.call`,
and `.lxc` serialisation. See the Functions section of
`docs/language-reference.html` and `tests/new_tests/test_default_params.py` /
`test_variadic.py`.

## Context

Functions currently take a **fixed** parameter list: `function()` in
`src/compiler/compile.go` increments `Arity` per parameter, and `vm.call`
(`src/vm/vm.go`, ~line 1991) rejects any call where `argCount != Arity`. Every
call must pass exactly N arguments. This plan adds default and variadic
parameters:

```lox
func greet(name, greeting="hi") { return greeting & ", " & name }
greet("Sam")            // "hi, Sam"
greet("Sam", "yo")      // "yo, Sam"

func log(msg, tags=[]) { tags.append(msg); return tags }   // fresh [] each call

func sum(*xs) { var t = 0; foreach (x in xs) { t = t + x }; return t }
sum(1, 2, 3)            // 6
func first_then(a, *rest) { ... }
```

**Defaults are evaluated at call time**: a default is an arbitrary expression
compiled into the function prologue and run only when the argument is omitted.
This supports computed defaults (`m=n*2`), fresh mutable defaults (`x=[]`, with
no shared-state trap), and defaults that call functions. `*rest` collects surplus
positional arguments into a list; it must be the last parameter.

## Design

**Parameter shape:** `p1, …, [opt1=expr, opt2=expr, …], [*rest]`. Rules enforced
at compile time: a non-default fixed parameter may not follow a defaulted one;
`*rest` must be last.

**`FunctionObject`** (`src/core/obj_function.go`) gains two fields:

- `Arity` (reused) = total named parameter local slots, **including** `*rest`.
- `MinArity int` = minimum args a caller must supply (fixed params without defaults).
- `IsVariadic bool` = last named parameter is `*rest`.

Derived in the VM: `fixedCount = Arity - (IsVariadic ? 1 : 0)`.

**Call-time argument shaping (`vm.call`).** The caller pushes `argCount` args as
today. `call()` now validates a *range* and shapes the stack so exactly `Arity`
parameter slots sit above the closure:

- Validate: variadic → `argCount >= MinArity`; else `MinArity <= argCount <= fixedCount`. Otherwise `RunTimeError`.
- Missing optional fixed params (slots `argCount … fixedCount-1`): push a new `UNDEFINED` sentinel value.
- `*rest`: if `argCount <= fixedCount`, push an empty `ListObject`; if `argCount > fixedCount`, pop the surplus `argCount-fixedCount` values (in order) into a `ListObject` and push it.
- Then set `Slots = stackTop - Arity - 1` and continue as now.

**Prologue fills the defaults.** For each defaulted parameter the compiler emits,
at the very start of the function body, a guard that runs the default expression
only if the slot is still `UNDEFINED`:

```
OP_JUMP_IF_DEFINED <slot> <offset>   ; if local[slot] != UNDEFINED, skip
  <default expression bytecode>
  OP_SET_LOCAL <slot>
  OP_POP
<offset:>
```

This is safe to emit inline during parameter parsing: parameters emit **no**
bytecode of their own (locals are just slot assignments, and the args are already
on the stack), so this code lands exactly at the start of the chunk = the
prologue. Because guards emit in parameter order, a later default may reference an
earlier parameter (`m=n*2`). The `UNDEFINED` sentinel is VM-internal and is always
replaced before the body runs, so it never reaches user code.

## Implementation

1. **`UNDEFINED` sentinel — `src/core/value.go`.** Add `VAL_UNDEFINED` to the
   `ValueType` enum and `UNDEFINED_VALUE = Value{Type: VAL_UNDEFINED}`. Add
   defensive cases (never expected to fire): `String()`/print → `"<undefined>"`,
   `isFalsey` → true. Not a chunk constant, so no bytecode-cache value-tag change.

2. **New opcode `OP_JUMP_IF_DEFINED` — `src/core/chunk.go`.** Operands: 1-byte
   local slot + 2-byte forward offset.

3. **`FunctionObject` fields — `src/core/obj_function.go`.** Add `MinArity int`,
   `IsVariadic bool` (defaults `0`/`false` in `MakeFunctionObject`).

4. **Compiler parameter loop — `src/compiler/compile.go` `function()`.** Replace
   the fixed loop. Track `minArity`, `sawDefault`. Per parameter:
   - `if p.match(TOKEN_STAR)`: parse `*rest` (`parseVariable`+`defineVariable`),
     set `IsVariadic`, count it in `Arity`; require `)` next (error if a parameter follows).
   - else `parseVariable`+`defineVariable`; `slot := currentCompiler.localCount - 1`.
     - `if p.match(TOKEN_EQUAL)`: `sawDefault=true`; emit the guard —
       `emitByte(OP_JUMP_IF_DEFINED); emitByte(slot);` then two `0xff` placeholder
       bytes (record `off := len(code)-2`); `p.expression()`;
       `emitBytes(OP_SET_LOCAL, slot)`; `emitByte(OP_POP)`; `patchJump(off)`
       (reuse existing `patchJump`).
     - else: `if sawDefault` → `error("non-default parameter after default")`; else `minArity++`.
   - Set `function.MinArity = minArity`, `function.IsVariadic = sawRest`. `Arity`
     keeps counting every named parameter local (incl. rest) as it does now.

5. **VM — `src/vm/vm.go`.**
   - `OP_JUMP_IF_DEFINED` handler (near `OP_JUMP_IF_FALSE`): inline-read the slot
     byte and 2-byte offset; if `vm.stack[frame.Slots+slot].Type != VAL_UNDEFINED`
     add offset to `frame.Ip`; no stack push/pop. Follow the existing inlined
     `readShort` pattern.
   - Rewrite `vm.call()` per the argument-shaping section: range check, pad
     `UNDEFINED`, build/append the `*rest` list (reuse the list construction used
     by list literals / `OP_BUILD_LIST`), then set `Slots`.

6. **Bytecode cache — write `src/core/value.go` (~line 426), read
   `src/vm/bc_cache.go` (~line 148).** Serialise `MinArity` (uint32) and
   `IsVariadic` (1 byte) after `UpvalueCount`, before the chunk, on both sides.
   **Run `bash bin/clear_lxc.sh`** (serialisation changed).

7. **Disassembler — `src/debug/debug.go`.** Add an `OP_JUMP_IF_DEFINED` case
   (slot byte + jump offset) so `-d` output works.

8. **Docs.** Update the Functions section of `docs/language-reference.html`
   (defaults, `*rest`, call-time-evaluation note, trailing-order rules) and add the
   two features to the README feature list.

## Verification

1. `go build -o bin/glox main.go` then `bash bin/clear_lxc.sh`.
2. Ad-hoc smoke test:
   ```lox
   func greet(name, greeting="hi") { return greeting & ", " & name }
   print greet("Sam")            // hi, Sam
   print greet("Sam", "yo")      // yo, Sam
   func acc(x, items=[]) { items.append(x); return items }
   print acc(1)                  // [ 1 ]   (fresh list)
   print acc(2)                  // [ 2 ]   (not [1,2] — no shared default)
   func sum(*xs) { var t = 0; foreach (x in xs) { t = t + x } return t }
   print sum()                   // 0
   print sum(1, 2, 3)            // 6
   func lead(a, *rest) { return a & "/" & str(len(rest)) }
   print lead("x", 1, 2)         // x/2
   func f(n, m=n*2) { return m }
   print f(4)                    // 8
   ```
   Plus error cases: too few required args still raises; `func f(a=1, b)` is a
   compile error; `*rest` not last is a compile error.
3. New tests under `tests/new_tests/` (pattern of `test_closure.py`):
   `test_default_params.py` + `test_variadic.py` with `.lox` fixtures covering the
   smoke cases above, run under both fresh-compile and cached `.lxc` paths.
4. Full suite green: `python -m pytest tests/new_tests/ -x -q` (with `LOX_PATH`/`PATH` set).
5. Demo smoke test unaffected: `bash bin/test_examples.sh` (13/13).

## Out of scope

- Keyword arguments at call sites (`f(greeting="yo")`) — positional only here.
- `**kwargs` dict capture.
