# thread Module Documentation

The `thread` module runs a closure on a new goroutine-backed `glox` VM instance
within the *same* OS process — unlike [`process`](PROCESS_MODULE.md), which spawns a
whole separate process per worker because there's no way to serialise a closure
across a process boundary, `thread.spawn()` can take an in-memory function directly,
since no process boundary is crossed.

**Important limitation, by design, not an oversight:** only a spawned closure's own
*captured (upvalue) locals* are isolated from the spawning script — deep-copied at
spawn time, so mutating one side afterward never leaks into the other. Top-level
`var`s, class statics, and module attributes are **not** isolated: they're shared,
mutable state across every thread, the same as any other multi-threaded runtime.
Reach for [`sync.Mutex`](SYNC_MODULE.md) when more than one thread needs to touch
one of those safely.

## Usage

```lox
import thread
```

## The basic pattern

```lox
import thread

t = thread.spawn(func() {
    c = thread.channel()
    x = c.recv()
    c.send(x * 2)
    return "done"
})

t.send(21)
print t.recv()   // 42
print t.wait()   // "done"
```

## Module functions

### `thread.spawn(closure, ...args)` → Thread
Deep-copies `closure`'s captured upvalues and `args` (so the new thread shares no
mutable captured state with the caller — see the limitation above), then runs the
copy on a new goroutine-backed VM.
- **closure**: any function value, including an anonymous `func() { ... }` literal
- **args**: extra arguments passed to the closure as its own call arguments
- Raises `ThreadError` if `closure` isn't a function, or if called from the REPL
  (not supported — see Limitations)

### `thread.channel()` → ThreadChannel
Called from *inside* a `thread.spawn()`-ed function, returns a `ThreadChannel` wired
to this thread's own communication channels — the `thread` module's analogue of
`process.parent()`.
- Raises `ThreadError` if called from outside a spawned thread

## Thread objects

The parent-side handle returned by `thread.spawn()`.

### `t.send(value)` → nil
Sends `value` to the thread (for it to read via its own `channel().recv()`). No
serialisation happens — the value is handed over directly, since both sides share
one address space. Raises `ThreadError` if the thread has already finished.

### `t.recv()` → value
Blocks until the thread sends a value via `channel().send()`, and returns it.
Raises `ThreadError` if the thread has finished (no more messages will ever arrive)
or if the thread ended abnormally (an uncaught exception, or a Go-level panic) —
either way the original exception's specific class isn't preserved, only its text.

### `t.try_recv()` → (ok, value)
Non-blocking version of `recv()`. Returns `(false, nil)` immediately if nothing is
waiting, or `(true, value)` if a message was ready.

### `t.wait()` → value
Blocks until the thread's function returns, and returns whatever it returned.
Raises `ThreadError` if the thread ended abnormally instead (see `recv()` above) —
unlike `process`'s `wait()`, which only ever returns an OS exit code, a thread's
`wait()` gives you its actual return value, since there's no process boundary to
lose it across.

### `t.cancel()` → nil
Requests cancellation. **Cooperative only, not a real kill**: it unblocks a thread
currently parked in `channel().send()`/`recv()` (both raise `ThreadError` there
instead), but cannot interrupt a thread stuck in a tight loop that never touches its
channel — there's no preemption point in the interpreter for that. A thread that
truly won't yield can't be force-stopped the way `process.kill()` can force-stop an
OS process.

## ThreadChannel objects

The worker-side handle returned by `thread.channel()`, called from inside the
spawned function.

### `c.send(value)` → nil
Sends `value` back to the parent (readable via `t.recv()`/`t.try_recv()`). Raises
`ThreadError` if the parent has called `cancel()`.

### `c.recv()` → value
Blocks until the parent sends a value via `t.send()`, and returns it. Raises
`ThreadError` if the parent has called `cancel()`, or if the parent handle is gone.

### `c.try_recv()` → (ok, value)
Non-blocking version of `recv()`, same shape as `Thread.try_recv()`.

## Limitations

- **Globals/class statics/module attributes are shared, not isolated** — see above.
  Use [`sync.Mutex`](SYNC_MODULE.md) to serialise access when more than one thread
  touches the same one.
- **`cancel()` is cooperative, not preemptive** — see `t.cancel()` above.
- **No fault isolation from a bare Go panic reaching the interpreter's own bugs the
  way a crashed OS process would give you for free** — a spawned thread's panic is
  caught and surfaced as `ThreadError`, but this is a narrower guarantee than a
  whole separate process dying on its own; it protects against the *thread's*
  Go-level failure, not against, say, a native extension corrupting shared process
  memory.
- **Not supported from the REPL** — `thread.spawn()` raises `ThreadError` if called
  from an interactive session, since the REPL's incremental global-variable growth
  isn't safe to run concurrently with an in-flight thread.
- An uncaught exception (or panic) inside a thread always surfaces as `ThreadError`,
  never the original exception's own class — only its message text carries across.
