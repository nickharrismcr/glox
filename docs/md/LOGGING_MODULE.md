# logging Module Documentation

The `logging` module provides basic leveled logging: a `Logger` class that
filters messages against its own configured minimum level, then hands a
fully-formatted line to a **writer** — a plain closure that receives the
line as its only argument. The default writer prints to stdout; swap in
something else (most commonly `logging.file_writer(file)`) to send output
elsewhere instead.

## Usage

```lox
import logging
```

## The basic pattern

```lox
import logging
from logging import Logger

log = Logger()              // name="root", level=Logger.INFO, writer=print
log.debug("not shown")      // below the default INFO level -- suppressed
log.info("starting up")     // [2024-01-15 09:30:00] [INFO] root: starting up
log.warn("low disk space")
log.error("connection failed")
```

## Why the writer has to be a closure, not an object

Two things about glox specifically shape this module's design:

- `print` is a reserved statement, not a first-class function value — it
  can't be stored in a field or passed around, so `Logger` can't just take
  "a thing to print with."
- [`os.write(file, text)`](OS_MODULE.md) only accepts a real file object
  from `os.open()` — it isn't duck-typed, so `Logger` can't hand it an
  arbitrary "writer object" either.

So a writer is always a `func(line)` that wraps whichever of those two
`Logger` should ultimately use. The default (`func(line) { print line; }`)
is built in; `logging.file_writer(file)` builds the file-writing
equivalent for you.

## Module functions

### `logging.file_writer(file)` → writer closure
Returns a `func(line)` that appends `line` (plus a trailing newline) to
`file` via `os.write()`, for use as a `Logger`'s `writer`. The caller still
owns the file's lifecycle — open it before constructing the `Logger`, close
it whenever logging is done:

```lox
import logging
import os
from logging import Logger

file = os.open("app.log", "a")
log = Logger("app", Logger.INFO, logging.file_writer(file))
log.info("wrote a line to app.log instead of stdout")
os.close(file)
```

## Logger objects

### `Logger(name, level, writer)`
All three arguments are optional: `name="root"`, `level=Logger.INFO`,
`writer=func(line) { print line; }`. `name` is purely a label baked into
every line this logger produces — there's no logger registry or hierarchy,
just independent `Logger` instances.

### Level constants: `Logger.DEBUG` / `.INFO` / `.WARN` / `.ERROR` / `.CRITICAL`
Plain integers, `10`/`20`/`30`/`40`/`50` — spaced by 10 (matching Python's
`logging` module) so a custom level can be slotted in between two of these
without renumbering anything, e.g. `log.log(25, "...")`.

### `.debug(msg)` / `.info(msg)` / `.warn(msg)` / `.error(msg)` / `.critical(msg)` → nil
Log `msg` at the named level. Each is a thin wrapper over `.log()`.

### `.log(level, msg)` → nil
Logs `msg` at an arbitrary integer `level`. If `level` is below this
logger's configured level, the message is **fully suppressed** — the line
is never formatted and the writer is never called. Otherwise, formats
`"[date time] [LEVEL] name: msg"` (date/time from
[`sys.today()`](BUILTINS.md#system-modules)/`sys.now()`, `LEVEL` from
`Logger.level_name(level)`) and passes that one string to the writer.

### `.set_level(level)` / `.get_level()` → nil / int
Change or query this logger's minimum level after construction.

### `.set_writer(writer)` → nil
Swap this logger's writer closure after construction.

### `Logger.level_name(level)` → string
Static method. Returns the name for a known level constant, or a
`"LEVEL<n>"` fallback for any other integer — never raises, since an
unrecognised level isn't treated as an error.

## Limitations

- No logger hierarchy/registry (no `logging.getLogger(name)`-style lookup
  the way Python's does) — every `Logger()` call makes an independent
  instance; reuse the same one yourself if you want one shared logger.
- No log rotation, multiple simultaneous writers, or structured
  (key=value/JSON) output — one writer, one plain formatted string, by
  design for a basic module.
- Timestamps come from [`sys.today()`/`sys.now()`](BUILTINS.md#system-modules),
  which read the real system clock — unlike `sys.clock()` elsewhere in
  glox, which measures elapsed time since the interpreter started, not
  wall-clock time.
