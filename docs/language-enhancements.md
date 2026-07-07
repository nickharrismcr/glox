# GLox language enhancement suggestions

A roadmap of candidate language features, ordered by value-for-effort. Each entry
notes the rationale, a syntax sketch, where it hooks into the compiler/VM, and a
rough effort estimate. Grounded in the current code — the Pratt parser and rule
table live in [`src/compiler/compile.go`](../src/compiler/compile.go), the scanner
in [`src/compiler/scanner.go`](../src/compiler/scanner.go), the run loop in
[`src/vm/vm.go`](../src/vm/vm.go).

> **Done:** Anonymous functions / lambdas (`func (a, b) { ... }` as an
> expression) — shipped. See the Functions section of the language reference.
>
> **Done:** Full compound-assignment set (`*=`, `/=`, `%=`) — shipped. Works on
> variables and object properties, alongside the existing `+=` / `-=`. See the
> Compound assignment section of the language reference.
>
> **Done:** Ternary / conditional expression (`cond ? a : b`) — shipped, C-style
> and right-associative, only the selected branch is evaluated. See the
> Conditional expression section of the language reference.

---

## Tier 1 — high value, natural fits

### 1. Full compound-assignment set (`*=`, `/=`, `%=`) — *done*

**Why.** Only `+=` and `-=` existed; the scanner had just `TOKEN_PLUS_EQUAL`
/ `TOKEN_MINUS_EQUAL`, handled in `namedVariable` and the property/index setters
in `compile.go`. Rounding out the set was the most-requested small ergonomic win
and followed the existing pattern exactly.

```lox
x *= 2
obj.scale /= 4
i %= n
```

**Hooks.** Added `TOKEN_STAR_EQUAL` / `TOKEN_SLASH_EQUAL` / `TOKEN_PERCENT_EQUAL`
in [`scanner.go`](../src/compiler/scanner.go) (`ScanToken` `*`/`/`/`%` cases),
then extended the two existing compound-assignment sites in
[`compile.go`](../src/compiler/compile.go) — the global/local `namedVariable`
path (`handleCompoundAssignment`) and the property setter path
(`handlePropertyCompoundAssignment`) — to emit `OP_MULTIPLY` / `OP_DIVIDE` /
`OP_MODULUS`. Indexed targets (`a[i] *= x`) remain unsupported, matching the
existing lack of `a[i] += x`.

**Effort.** Low — mechanical, mirrored code that already existed.

### 2. Ternary / conditional expression (`cond ? a : b`) — *done*

**Why.** There was no expression-level conditional — only the `if` *statement* —
so choosing a value inline forced a temp variable and a 4-line block. `?`/`:`
also composes with the lambdas for compact callbacks. Shipped as **C-style**
`cond ? a : b` (not Python's `a if cond else b`): the condition is the left
operand and compiles first, so only the selected branch is evaluated — no
bytecode relocation needed.

```lox
var label = n == 1 ? "item" : "items"
```

**Hooks.** Added a `TOKEN_QUESTION` token (scanner); `:` already existed as
`TOKEN_COLON`. Inserted a `PREC_CONDITIONAL` precedence level (between
`PREC_ASSIGNMENT` and `PREC_OR`) and registered a `conditional` infix rule on
`TOKEN_QUESTION`, emitting the same `OP_JUMP_IF_FALSE` / `OP_JUMP` / `OP_POP`
shape as `and_`/`or_`. Right-associative. Note `OP_JUMP_IF_FALSE` peeks (does
not pop), so the condition is popped on both branches. Truthiness follows the
VM's existing `isFalsey` — as elsewhere in glox, bare integers are falsey, so
use boolean/comparison conditions.

**Effort.** Moderate — one new token, one precedence level, one infix rule,
reused existing jump/patch helpers.

---

## Tier 2 — high value, more work

### 3. String interpolation (f-strings)

**Why.** `format()` (a `fmt.Sprintf` wrapper) exists, but there is no inline
interpolation, and `print` takes a single expression — so `print("a", b)` prints
a *tuple*, a documented footgun. Interpolation removes most `&`-concatenation and
`format()` boilerplate.

```lox
print "total: ${count} (${pct}%)"
```

**Hooks.** Scanner-level: recognise `${ … }` inside string literals and emit a
token sequence the compiler desugars into concatenation / a `format` call. Touches
[`scanner.go`](../src/compiler/scanner.go) `string()` and a small compiler helper.

**Effort.** Moderate-high (scanner + compiler); the runtime is unchanged.

### 4. Default & variadic parameters

**Why.** `function()` parses a fixed parameter list and bumps `Arity`
unconditionally, so every call must pass exactly N arguments. Defaults and a
trailing `*rest` pair especially well with the functional-style APIs
(`map`/`filter`/`reduce`) now that inline lambdas exist.

```lox
func greet(name, greeting="hi") { ... }
func sum(*xs) { ... }
```

**Hooks.** Extend the parameter loop in `function()`
([`compile.go`](../src/compiler/compile.go)) to record defaults / a rest slot,
and adjust argument binding and the arity check at the call sites (`OP_CALL`
handling in [`vm.go`](../src/vm/vm.go)).

**Effort.** Moderate — parameter parsing plus call-time arity/argument handling.

---

## Tier 3 — domain-specific / nice-to-have

### 5. Bitwise operators for ints (`|`, `^`, `<<`, `>>`, bitwise-and)

**Why.** Integers are first-class and the colour encoding packs RGBA into a
single int, so bit ops are the idiomatic tool for the graphics/colour code.
**Caveat:** `&` is already string concatenation and `++` is vector concatenation,
so bitwise-and needs a fresh token or a keyword form (`band`/`bor`/…).

**Effort.** Medium — new tokens + binary rules + integer-only opcodes; the `&`
clash needs a naming decision.

### 6. `switch` / `match`

**Why.** Multiway branching is only expressible as `if`/`else if` chains. Lower
urgency because `else if` already works (statement dispatch re-enters on
`TOKEN_IF`), so this is sugar rather than new capability.

**Effort.** Moderate-high — new statement form, jump-table or comparison-chain
codegen, break semantics.

---

## Recommended order

1. ~~**Compound assignment** (`*=`/`/=`/`%=`)~~ — done.
2. ~~**Ternary**~~ — done (C-style `cond ? a : b`).
3. **String interpolation** — biggest day-to-day ergonomics improvement.
4. **Default/variadic params** — rounds out the function model.
5. Bitwise / `switch` as demand arises.

Everything here reuses existing machinery (Pratt rules, jump/patch helpers, the
peephole pass) rather than introducing new subsystems.
