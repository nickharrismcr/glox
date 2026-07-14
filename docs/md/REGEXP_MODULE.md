# re Module Documentation

The `re` module provides regular expressions, similar to Python's `re` module, built directly on
Go's standard `regexp` package (RE2 syntax).

**Limitations vs. Python's `re`:**
- RE2 does not support backreferences (`\1` inside a pattern) or lookaround (`(?=...)`, `(?!...)`).
- Replacement strings in `sub()`/`subn()` use Go's `$1` / `${name}` group-reference syntax, not
  Python's `\1` / `\g<name>`.
- `split()` does not interleave captured-group text into the result list (Go's `regexp.Split`
  doesn't support that).

## Usage

```lox
import re
```

## Module functions

### `re.search(pattern, s)` → Match or nil
Finds the first match anywhere in `s`.
- **pattern**: regular expression string
- **s**: string to search
- **Returns**: a [Match object](#match-objects), or `nil` if there is no match

```lox
m = re.search("\d+", "abc123def")
print m.group(0)   // "123"
```

### `re.match(pattern, s)` → Match or nil
Like `search()`, but only matches if the match starts at position 0 of `s` (it does not have to
consume the whole string).

### `re.fullmatch(pattern, s)` → Match or nil
Only matches if the pattern matches the entirety of `s`.

### `re.sub(pattern, repl, s [, count])` → string
Replaces matches of `pattern` in `s` with `repl`.
- **repl**: replacement string; may reference captured groups with `$1`, `$2`, ... or `${name}`
- **count**: maximum number of replacements to make. Omitted or `0` means replace all.

```lox
print re.sub("\d+", "#", "a1 b22 c333")        // "a# b# c#"
print re.sub("\d+", "#", "a1 b22 c333", 1)     // "a# b22 c333"
print re.sub("(\w+)@(\w+)", "$2@$1", "bob@example")  // "example@bob"
```

### `re.subn(pattern, repl, s [, count])` → (string, int)
Same as `sub()`, but returns a tuple of `(new_string, number_of_replacements)`.

```lox
result = re.subn("\d+", "#", "a1 b22 c333")
print result[0]   // "a# b# c#"
print result[1]   // 3
```

### `re.split(pattern, s [, maxsplit])` → list
Splits `s` on each match of `pattern`.
- **maxsplit**: maximum number of splits to make. Omitted or `0` means no limit.

```lox
print re.split(",\s*", "a, b,c,   d")   // ["a", "b", "c", "d"]
print re.split(",\s*", "a, b,c,   d", 1)  // ["a", "b,c,   d"]
```

### `re.findall(pattern, s)` → list
Returns all non-overlapping matches of `pattern` in `s`.
- If `pattern` has no capturing groups, returns a list of the matched substrings.
- If `pattern` has exactly one group, returns a list of that group's matched text.
- If `pattern` has 2+ groups, returns a list of tuples, one per match.

```lox
print re.findall("\d+", "a1 b22 c333")        // ["1", "22", "333"]
print re.findall("(\w)(\d+)", "a1 b22")       // [("a", "1"), ("b", "22")]
```

### `re.compile(pattern)` → Pattern
Precompiles `pattern` into a reusable [Pattern object](#pattern-objects), avoiding recompilation
on every call — useful when the same pattern is applied many times (e.g. in a loop).

```lox
digits = re.compile("\d+")
print digits.search("xx42yy").group(0)   // "42"
print digits.sub("#", "1 2 3")           // "# # #"
```

## Match objects

Returned by `search()`, `match()`, and `fullmatch()` on success.

### `m.group()` / `m.group(n)` / `m.group(name)`
Returns the matched text for group `0` (the whole match), a numbered group `n`, or a named group.
Returns `nil` if that group did not participate in the match (e.g. an unmatched optional group).

### `m.groups()`
Returns a tuple of all groups `1..N` (in order), with `nil` for any group that didn't participate.

### `m.groupdict()`
Returns a dict mapping named-group names (from `(?P<name>...)`) to their matched text.

```lox
m = re.search("(?P<user>\w+)@(?P<host>\w+)", "alice@wonderland")
print m.group("user")        // "alice"
print m.groupdict()["host"]  // "wonderland"
```

### `m.start()` / `m.start(n)`, `m.end()` / `m.end(n)`, `m.span()` / `m.span(n)`
Return the byte offset(s) into the searched string of the whole match (default) or group `n`.
`span()` returns the `(start, end)` tuple.

## Pattern objects

Returned by `re.compile(pattern)`. Exposes the same operations as the module-level functions,
minus the leading `pattern` argument, reusing the precompiled regular expression:

- `search(s)`, `match(s)`, `fullmatch(s)`
- `sub(repl, s [, count])`, `subn(repl, s [, count])`
- `split(s [, maxsplit])`
- `findall(s)`
