# GLox language enhancement suggestions

A roadmap of candidate language features, ordered by value-for-effort. Each entry
notes the rationale, a syntax sketch, where it hooks into the compiler/VM, and a
rough effort estimate. Grounded in the current code — the Pratt parser and rule
table live in [`src/compiler/compile.go`](../src/compiler/compile.go), the scanner
in [`src/compiler/scanner.go`](../src/compiler/scanner.go), the run loop in
[`src/vm/vm.go`](../src/vm/vm.go).

> **Done:** Anonymous functions / lambdas (`func (a, b) { ... }` as an
> expression) — shipped. See the Functions section of the language reference.

---

## Tier 1 — high value, natural fits

### 1. Full compound-assignment set (`*=`, `/=`, `%=`)

**Why.** Only `+=` and `-=` exist today; the scanner has just `TOKEN_PLUS_EQUAL`
/ `TOKEN_MINUS_EQUAL`, handled in `namedVariable` and the property/index setters
in `compile.go`. Rounding out the set is the most-requested small ergonomic win
and follows the existing pattern exactly.

```lox
x *= 2
obj.scale /= 4
i %= n
```

**Hooks.** Add `TOKEN_STAR_EQUAL` / `TOKEN_SLASH_EQUAL` / `TOKEN_PERCENT_EQUAL`
in [`scanner.go`](../src/compiler/scanner.go) (`ScanToken` `*`/`/`/`%` cases),
then extend the two existing compound-assignment sites in
[`compile.go`](../src/compiler/compile.go) (the global/local `namedVariable`
path and the `dot`/`slice` setter path that already switch on
`TOKEN_PLUS_EQUAL`/`TOKEN_MINUS_EQUAL`).

**Effort.** Low — mechanical, mirrors code that already exists.

### 2. Ternary / conditional expression (`cond ? a : b`)

**Why.** There is no expression-level conditional — only the `if` *statement* —
so choosing a value inline forces a temp variable and a 4-line block. `?`/`:`
also composes with the new lambdas for compact callbacks.

```lox
var label = n == 1 ? "item" : "items"
```

**Hooks.** Add a `TOKEN_QUESTION` token (scanner); `:` already exists as
`TOKEN_COLON`. Register a `conditional` infix rule on `TOKEN_QUESTION` at roughly
`PREC_ASSIGNMENT`/`PREC_OR` in the rule table, emitting the same
`OP_JUMP_IF_FALSE` / `OP_JUMP` / `OP_POP` shape as `ifStatement`.

**Effort.** Moderate — one new token, one infix rule, reuse existing jump/patch
helpers.

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

1. **Compound assignment** (`*=`/`/=`/`%=`) — quick, mechanical win.
2. **Ternary** — small, high-use, composes with lambdas.
3. **String interpolation** — biggest day-to-day ergonomics improvement.
4. **Default/variadic params** — rounds out the function model.
5. Bitwise / `switch` as demand arises.

Everything here reuses existing machinery (Pratt rules, jump/patch helpers, the
peephole pass) rather than introducing new subsystems.
